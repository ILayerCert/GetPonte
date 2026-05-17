import { writable, get } from 'svelte/store';

export interface RemotePeer {
	peerId: string;
	peerName: string;
	stream: MediaStream;
}

export interface MediaState {
	localStream: MediaStream | null;
	screenStream: MediaStream | null;
	remotePeers: RemotePeer[];
	audioEnabled: boolean;
	videoEnabled: boolean;
	screenSharing: boolean;
	inCall: boolean;
}

function createMediaStore() {
	const initial: MediaState = {
		localStream: null,
		screenStream: null,
		remotePeers: [],
		audioEnabled: true,
		videoEnabled: true,
		screenSharing: false,
		inCall: false
	};

	const { subscribe, set, update } = writable<MediaState>(initial);

	return {
		subscribe,

		async getUserMedia(): Promise<MediaStream | null> {
			try {
				const stream = await navigator.mediaDevices.getUserMedia({
					audio: true,
					video: {
						width: { ideal: 1280 },
						height: { ideal: 720 },
						frameRate: { ideal: 30 }
					}
				});
				update(s => ({ ...s, localStream: stream, inCall: true, audioEnabled: true, videoEnabled: true }));
				return stream;
			} catch (e) {
				console.error('Failed to get user media:', e);
				// Try audio only
				try {
					const audioStream = await navigator.mediaDevices.getUserMedia({ audio: true, video: false });
					update(s => ({ ...s, localStream: audioStream, inCall: true, audioEnabled: true, videoEnabled: false }));
					return audioStream;
				} catch (e2) {
					console.error('Failed to get audio:', e2);
					return null;
				}
			}
		},

		toggleAudio() {
			update(s => {
				if (s.localStream) {
					s.localStream.getAudioTracks().forEach(t => {
						t.enabled = !s.audioEnabled;
					});
				}
				return { ...s, audioEnabled: !s.audioEnabled };
			});
		},

		toggleVideo() {
			update(s => {
				if (s.localStream) {
					s.localStream.getVideoTracks().forEach(t => {
						t.enabled = !s.videoEnabled;
					});
				}
				return { ...s, videoEnabled: !s.videoEnabled };
			});
		},

		async toggleScreenShare(): Promise<MediaStream | null> {
			const state = get({ subscribe });
			if (state.screenSharing && state.screenStream) {
				state.screenStream.getTracks().forEach(t => t.stop());
				update(s => ({ ...s, screenStream: null, screenSharing: false }));
				return null;
			}
			try {
				const screen = await navigator.mediaDevices.getDisplayMedia({
					video: true,
					audio: false
				});
				screen.getVideoTracks()[0].onended = () => {
					update(s => ({ ...s, screenStream: null, screenSharing: false }));
				};
				update(s => ({ ...s, screenStream: screen, screenSharing: true }));
				return screen;
			} catch (e) {
				console.error('Screen share failed:', e);
				return null;
			}
		},

		addRemotePeer(peerId: string, peerName: string, stream: MediaStream) {
			update(s => {
				const filtered = s.remotePeers.filter(p => p.peerId !== peerId);
				return { ...s, remotePeers: [...filtered, { peerId, peerName, stream }] };
			});
		},

		removeRemotePeer(peerId: string) {
			update(s => ({
				...s,
				remotePeers: s.remotePeers.filter(p => p.peerId !== peerId)
			}));
		},

		stopAll() {
			const state = get({ subscribe });
			if (state.localStream) {
				state.localStream.getTracks().forEach(t => t.stop());
			}
			if (state.screenStream) {
				state.screenStream.getTracks().forEach(t => t.stop());
			}
			set(initial);
		}
	};
}

export const media = createMediaStore();
