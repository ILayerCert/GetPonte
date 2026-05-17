export type MessageType = 'chat_message' | 'user_joined' | 'user_left' | 'typing' | 'webrtc_signal' | 'room_update' | 'error';

/**
 * WSMessage matches the backend OutboundMessage format:
 *   { type, message_id, user_id, username, content, room_id, timestamp }
 */
export interface WSMessage {
	type: MessageType;
	message_id?: string;
	user_id?: string;
	username?: string;
	content?: string;
	room_id?: string;
	timestamp?: string;
	// For webrtc_signal forwarding
	payload?: any;
}

type MessageHandler = (msg: WSMessage) => void;

export class WebSocketClient {
	private ws: WebSocket | null = null;
	private handlers: Map<MessageType, MessageHandler[]> = new Map();
	private reconnectAttempts = 0;
	private maxReconnectAttempts = 10;
	private reconnectTimer: ReturnType<typeof setTimeout> | null = null;
	private token: string;
	private roomId: string;
	private url: string;
	private _connected = false;
	private intentionalClose = false;

	constructor(token: string, roomId: string) {
		this.token = token;
		this.roomId = roomId;
		const wsBase = import.meta.env.VITE_WS_URL || 'ws://localhost:8080';
		this.url = `${wsBase}/ws?token=${encodeURIComponent(token)}&room_id=${encodeURIComponent(roomId)}`;
	}

	get connected(): boolean {
		return this._connected;
	}

	connect(): void {
		this.intentionalClose = false;
		try {
			this.ws = new WebSocket(this.url);

			this.ws.onopen = () => {
				this._connected = true;
				this.reconnectAttempts = 0;
				console.log(`[WS] Connected to room ${this.roomId}`);
			};

			this.ws.onmessage = (event) => {
				try {
					const msg: WSMessage = JSON.parse(event.data);
					this.dispatch(msg);
				} catch (e) {
					console.error('[WS] Failed to parse message:', e);
				}
			};

			this.ws.onclose = (event) => {
				this._connected = false;
				console.log(`[WS] Disconnected (code: ${event.code})`);
				if (!this.intentionalClose) {
					this.scheduleReconnect();
				}
			};

			this.ws.onerror = (error) => {
				console.error('[WS] Error:', error);
			};
		} catch (e) {
			console.error('[WS] Connection failed:', e);
			this.scheduleReconnect();
		}
	}

	private scheduleReconnect(): void {
		if (this.reconnectAttempts >= this.maxReconnectAttempts) {
			console.error('[WS] Max reconnect attempts reached');
			return;
		}
		const delay = Math.min(1000 * Math.pow(2, this.reconnectAttempts), 30000);
		this.reconnectAttempts++;
		console.log(`[WS] Reconnecting in ${delay}ms (attempt ${this.reconnectAttempts})`);
		this.reconnectTimer = setTimeout(() => this.connect(), delay);
	}

	on(type: MessageType, handler: MessageHandler): void {
		const existing = this.handlers.get(type) || [];
		existing.push(handler);
		this.handlers.set(type, existing);
	}

	off(type: MessageType, handler: MessageHandler): void {
		const existing = this.handlers.get(type) || [];
		this.handlers.set(type, existing.filter(h => h !== handler));
	}

	private dispatch(msg: WSMessage): void {
		const handlers = this.handlers.get(msg.type) || [];
		handlers.forEach(h => h(msg));
	}

	/**
	 * Send a message to the backend.
	 * Backend expects: { type: "chat_message", content: "..." }
	 */
	send(type: MessageType, data: any): void {
		if (!this.ws || this.ws.readyState !== WebSocket.OPEN) {
			console.warn('[WS] Not connected, cannot send');
			return;
		}
		this.ws.send(JSON.stringify({ type, content: data.content || '', ...data }));
	}

	disconnect(): void {
		this.intentionalClose = true;
		if (this.reconnectTimer) {
			clearTimeout(this.reconnectTimer);
			this.reconnectTimer = null;
		}
		if (this.ws) {
			this.ws.close(1000, 'Client disconnect');
			this.ws = null;
		}
		this._connected = false;
		this.handlers.clear();
	}
}
