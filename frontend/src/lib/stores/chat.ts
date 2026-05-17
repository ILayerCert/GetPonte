import { writable, get } from 'svelte/store';
import { get as apiGet } from '$lib/api';
import type { WebSocketClient, WSMessage } from '$lib/websocket';

export interface ChatMessage {
	id: string;
	room_id: string;
	sender_id: string;
	sender_name: string;
	content: string;
	timestamp: string;
}

export interface ChatState {
	messages: Map<string, ChatMessage[]>; // roomId -> messages
	typingUsers: Map<string, Set<string>>; // roomId -> set of usernames
	loading: boolean;
}

function createChatStore() {
	const initial: ChatState = {
		messages: new Map(),
		typingUsers: new Map(),
		loading: false
	};

	const { subscribe, set, update } = writable<ChatState>(initial);

	let typingTimers: Map<string, ReturnType<typeof setTimeout>> = new Map();
	let wsClient: WebSocketClient | null = null;

	return {
		subscribe,

		connectWS(ws: WebSocketClient, roomId: string) {
			wsClient = ws;

			// Backend sends: { type, message_id, user_id, username, content, room_id, timestamp }
			ws.on('chat_message', (msg: WSMessage) => {
				const chatMsg: ChatMessage = {
					id: msg.message_id || crypto.randomUUID(),
					room_id: msg.room_id || roomId,
					sender_id: msg.user_id || '',
					sender_name: msg.username || 'Unknown',
					content: msg.content || '',
					timestamp: msg.timestamp || new Date().toISOString()
				};
				update(s => {
					const msgs = new Map(s.messages);
					const existing = msgs.get(roomId) || [];
					msgs.set(roomId, [...existing, chatMsg]);
					return { ...s, messages: msgs };
				});
			});

			ws.on('typing', (msg: WSMessage) => {
				const senderName = msg.username;
				if (!senderName) return;
				const key = `${roomId}:${senderName}`;
				update(s => {
					const typing = new Map(s.typingUsers);
					const users = new Set(typing.get(roomId) || []);
					users.add(senderName);
					typing.set(roomId, users);
					return { ...s, typingUsers: typing };
				});

				// Clear typing after 3 seconds
				if (typingTimers.has(key)) clearTimeout(typingTimers.get(key)!);
				typingTimers.set(key, setTimeout(() => {
					update(s => {
						const typing = new Map(s.typingUsers);
						const users = new Set(typing.get(roomId) || []);
						users.delete(senderName);
						typing.set(roomId, users);
						return { ...s, typingUsers: typing };
					});
				}, 3000));
			});

			ws.on('user_joined', (msg: WSMessage) => {
				console.log(`[Chat] ${msg.username} joined the room`);
			});

			ws.on('user_left', (msg: WSMessage) => {
				console.log(`[Chat] ${msg.username} left the room`);
			});
		},

		async loadHistory(roomId: string) {
			update(s => ({ ...s, loading: true }));
			try {
				const messages = await apiGet<ChatMessage[]>(`/api/rooms/${roomId}/messages`);
				update(s => {
					const msgs = new Map(s.messages);
					// API returns messages with: id, room_id, user_id, username, content, msg_type, created_at
					const mapped = (messages || []).map((m: any) => ({
						id: m.id,
						room_id: m.room_id || roomId,
						sender_id: m.user_id || '',
						sender_name: m.username || 'Unknown',
						content: m.content || '',
						timestamp: m.created_at || m.timestamp || ''
					}));
					msgs.set(roomId, mapped);
					return { ...s, messages: msgs, loading: false };
				});
			} catch (e) {
				console.error('Failed to load chat history:', e);
				update(s => {
					const msgs = new Map(s.messages);
					msgs.set(roomId, []);
					return { ...s, messages: msgs, loading: false };
				});
			}
		},

		sendMessage(roomId: string, content: string) {
			if (!wsClient || !content.trim()) return;
			// Backend expects: { type: "chat_message", content: "..." }
			wsClient.send('chat_message', { content: content.trim() });
		},

		sendTyping(roomId: string) {
			if (!wsClient) return;
			wsClient.send('typing', {});
		},

		getMessages(roomId: string): ChatMessage[] {
			const state = get({ subscribe });
			return state.messages.get(roomId) || [];
		},

		clear() {
			typingTimers.forEach(t => clearTimeout(t));
			typingTimers.clear();
			wsClient = null;
			set(initial);
		}
	};
}

export const chat = createChatStore();
