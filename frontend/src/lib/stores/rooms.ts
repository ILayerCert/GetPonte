import { writable } from 'svelte/store';
import { get, post, del } from '$lib/api';

export interface Room {
	id: string;
	name: string;
	description?: string;
	is_private?: boolean;
	created_by: string;
	created_at: string;
	// enriched from RoomDetail
	member_count?: number;
	members?: any[];
}

export interface RoomsState {
	rooms: Room[];
	activeRoom: Room | null;
	loading: boolean;
	error: string | null;
}

function createRoomsStore() {
	const initial: RoomsState = {
		rooms: [],
		activeRoom: null,
		loading: false,
		error: null
	};

	const { subscribe, set, update } = writable<RoomsState>(initial);

	return {
		subscribe,

		async loadRooms() {
			update(s => ({ ...s, loading: true, error: null }));
			try {
				const rooms = await get<Room[]>('/api/rooms');
				update(s => ({ ...s, rooms: rooms || [], loading: false }));
			} catch (e: any) {
				update(s => ({ ...s, loading: false, error: e.message }));
			}
		},

		async createRoom(name: string, description: string) {
			update(s => ({ ...s, loading: true, error: null }));
			try {
				const room = await post<Room>('/api/rooms', { name, description });
				update(s => ({
					...s,
					rooms: [...s.rooms, room],
					loading: false
				}));
				return room;
			} catch (e: any) {
				update(s => ({ ...s, loading: false, error: e.message }));
				return null;
			}
		},

		async joinRoom(id: string) {
			try {
				await post(`/api/rooms/${id}/join`);
			} catch (e: any) {
				// "already a member" is fine
				console.log('[rooms] join:', e.message);
			}
		},

		async leaveRoom(id: string) {
			try {
				await del(`/api/rooms/${id}/leave`);
				update(s => ({
					...s,
					rooms: s.rooms.filter(r => r.id !== id),
					activeRoom: s.activeRoom?.id === id ? null : s.activeRoom
				}));
			} catch (e: any) {
				update(s => ({ ...s, error: e.message }));
			}
		},

		async getRoom(id: string) {
			try {
				// Backend returns { room: {...}, members: [...] }
				const detail = await get<{ room: Room; members: any[] }>(`/api/rooms/${id}`);
				const room: Room = {
					...detail.room,
					members: detail.members,
					member_count: detail.members?.length || 0
				};
				update(s => ({ ...s, activeRoom: room }));
				return room;
			} catch (e: any) {
				update(s => ({ ...s, error: e.message }));
				return null;
			}
		},

		setActive(room: Room | null) {
			update(s => ({ ...s, activeRoom: room }));
		},

		clear() {
			set(initial);
		}
	};
}

export const rooms = createRoomsStore();
