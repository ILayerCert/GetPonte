import type { WebSocketClient } from './websocket';

interface PeerConnection {
	peerId: string;
	peerName: string;
	peer: any; // SimplePeer instance
	stream: MediaStream | null;
}

type OnStreamCallback = (peerId: string, peerName: string, stream: MediaStream) => void;
type OnDisconnectCallback = (peerId: string) => void;

export class WebRTCManager {
	private peers: Map<string, PeerConnection> = new Map();
	private localStream: MediaStream | null = null;
	private ws: WebSocketClient;
	private userId: string;
	private onStream: OnStreamCallback;
	private onDisconnect: OnDisconnectCallback;
	private SimplePeer: any = null;

	constructor(
		ws: WebSocketClient,
		userId: string,
		onStream: OnStreamCallback,
		onDisconnect: OnDisconnectCallback
	) {
		this.ws = ws;
		this.userId = userId;
		this.onStream = onStream;
		this.onDisconnect = onDisconnect;

		// Listen for WebRTC signals
		this.ws.on('webrtc_signal', (msg) => {
			this.handleSignal(msg.sender_id!, msg.payload);
		});

		// When a new user joins, initiate connection
		this.ws.on('user_joined', (msg) => {
			if (msg.sender_id !== this.userId && this.localStream) {
				this.createPeer(msg.sender_id!, msg.sender_name || 'Unknown', true);
			}
		});

		this.ws.on('user_left', (msg) => {
			if (msg.sender_id) {
				this.removePeer(msg.sender_id);
			}
		});
	}

	async loadSimplePeer(): Promise<void> {
		if (!this.SimplePeer) {
			const mod = await import('simple-peer');
			this.SimplePeer = mod.default || mod;
		}
	}

	addLocalStream(stream: MediaStream): void {
		this.localStream = stream;
	}

	async joinCall(): Promise<void> {
		await this.loadSimplePeer();
		// Signal to room that we're joining the call
		this.ws.send('webrtc_signal', { type: 'join_call', user_id: this.userId });
	}

	private createPeer(peerId: string, peerName: string, initiator: boolean): void {
		if (this.peers.has(peerId) || !this.SimplePeer) return;

		const peer = new this.SimplePeer({
			initiator,
			stream: this.localStream || undefined,
			trickle: true,
			config: {
				iceServers: [
					{ urls: 'stun:stun.l.google.com:19302' },
					{ urls: 'stun:stun1.l.google.com:19302' }
				]
			}
		});

		const conn: PeerConnection = { peerId, peerName, peer, stream: null };

		peer.on('signal', (data: any) => {
			this.ws.send('webrtc_signal', {
				type: 'signal',
				target_id: peerId,
				signal: data
			});
		});

		peer.on('stream', (stream: MediaStream) => {
			conn.stream = stream;
			this.onStream(peerId, peerName, stream);
		});

		peer.on('close', () => {
			this.removePeer(peerId);
		});

		peer.on('error', (err: Error) => {
			console.error(`[WebRTC] Peer ${peerId} error:`, err);
			this.removePeer(peerId);
		});

		this.peers.set(peerId, conn);
	}

	handleSignal(senderId: string, payload: any): void {
		if (payload.type === 'join_call') {
			// New peer wants to join, create a non-initiator peer
			if (senderId !== this.userId && this.localStream) {
				this.createPeer(senderId, payload.sender_name || 'Unknown', false);
			}
			return;
		}

		if (payload.type === 'signal' && payload.signal) {
			const conn = this.peers.get(senderId);
			if (conn) {
				try {
					conn.peer.signal(payload.signal);
				} catch (e) {
					console.error('[WebRTC] Signal error:', e);
				}
			} else if (this.localStream) {
				// Create peer on-demand if we receive a signal but don't have the peer yet
				this.createPeer(senderId, payload.sender_name || 'Unknown', false);
				setTimeout(() => {
					const newConn = this.peers.get(senderId);
					if (newConn) {
						try {
							newConn.peer.signal(payload.signal);
						} catch (e) {
							console.error('[WebRTC] Delayed signal error:', e);
						}
					}
				}, 100);
			}
		}
	}

	private removePeer(peerId: string): void {
		const conn = this.peers.get(peerId);
		if (conn) {
			try { conn.peer.destroy(); } catch (_) {}
			this.peers.delete(peerId);
			this.onDisconnect(peerId);
		}
	}

	replaceStream(newStream: MediaStream): void {
		this.localStream = newStream;
		this.peers.forEach(conn => {
			try {
				// Remove old tracks and add new ones
				conn.peer.removeStream(conn.peer.streams?.[0]);
				conn.peer.addStream(newStream);
			} catch (e) {
				console.error('[WebRTC] Replace stream error:', e);
			}
		});
	}

	leaveCall(): void {
		this.peers.forEach((conn, id) => {
			try { conn.peer.destroy(); } catch (_) {}
			this.onDisconnect(id);
		});
		this.peers.clear();
		this.localStream = null;
	}

	getPeerCount(): number {
		return this.peers.size;
	}
}
