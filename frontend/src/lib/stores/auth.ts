import { writable, get } from 'svelte/store';
import { api } from '$lib/api';

export interface User {
	id: string;
	username: string;
	email: string;
	avatar_url?: string;
}

export interface AuthState {
	token: string | null;
	user: User | null;
	loading: boolean;
	error: string | null;
}

function createAuthStore() {
	const initial: AuthState = {
		token: null,
		user: null,
		loading: false,
		error: null
	};

	const { subscribe, set, update } = writable<AuthState>(initial);

	// Restore from localStorage on init
	if (typeof window !== 'undefined') {
		const savedToken = localStorage.getItem('ponte_token');
		const savedUser = localStorage.getItem('ponte_user');
		if (savedToken && savedUser) {
			try {
				set({
					token: savedToken,
					user: JSON.parse(savedUser),
					loading: false,
					error: null
				});
			} catch {
				localStorage.removeItem('ponte_token');
				localStorage.removeItem('ponte_user');
			}
		}
	}

	return {
		subscribe,
		login: async (email: string, password: string) => {
			update(s => ({ ...s, loading: true, error: null }));
			try {
				const res = await api<{ token: string; user: User }>('/api/auth/login', {
					method: 'POST',
					body: JSON.stringify({ email, password }),
					skipAuth: true
				});
				localStorage.setItem('ponte_token', res.token);
				localStorage.setItem('ponte_user', JSON.stringify(res.user));
				set({ token: res.token, user: res.user, loading: false, error: null });
				return true;
			} catch (e: any) {
				update(s => ({ ...s, loading: false, error: e.message || 'Login failed' }));
				return false;
			}
		},
		register: async (username: string, email: string, password: string) => {
			update(s => ({ ...s, loading: true, error: null }));
			try {
				const res = await api<{ token: string; user: User }>('/api/auth/register', {
					method: 'POST',
					body: JSON.stringify({ username, email, password }),
					skipAuth: true
				});
				localStorage.setItem('ponte_token', res.token);
				localStorage.setItem('ponte_user', JSON.stringify(res.user));
				set({ token: res.token, user: res.user, loading: false, error: null });
				return true;
			} catch (e: any) {
				update(s => ({ ...s, loading: false, error: e.message || 'Registration failed' }));
				return false;
			}
		},
		logout: () => {
			localStorage.removeItem('ponte_token');
			localStorage.removeItem('ponte_user');
			set({ token: null, user: null, loading: false, error: null });
		},
		clearError: () => {
			update(s => ({ ...s, error: null }));
		},
		getToken: (): string | null => {
			return get({ subscribe }).token;
		}
	};
}

export const auth = createAuthStore();
