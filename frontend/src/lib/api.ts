const BASE_URL = import.meta.env.VITE_API_URL || 'http://localhost:8080';

function getToken(): string | null {
	if (typeof window === 'undefined') return null;
	return localStorage.getItem('ponte_token');
}

interface ApiOptions extends RequestInit {
	skipAuth?: boolean;
}

export async function api<T = any>(path: string, options: ApiOptions = {}): Promise<T> {
	const { skipAuth, ...fetchOptions } = options;
	const headers: Record<string, string> = {
		'Content-Type': 'application/json',
		...(fetchOptions.headers as Record<string, string> || {})
	};

	if (!skipAuth) {
		const token = getToken();
		if (token) {
			headers['Authorization'] = `Bearer ${token}`;
		}
	}

	const response = await fetch(`${BASE_URL}${path}`, {
		...fetchOptions,
		headers
	});

	if (!response.ok) {
		const body = await response.json().catch(() => ({ message: response.statusText }));
		throw new ApiError(response.status, body.message || body.error || 'Request failed');
	}

	if (response.status === 204) return undefined as T;
	return response.json();
}

export class ApiError extends Error {
	constructor(public status: number, message: string) {
		super(message);
		this.name = 'ApiError';
	}
}

export const get = <T = any>(path: string) => api<T>(path, { method: 'GET' });
export const post = <T = any>(path: string, body?: any) =>
	api<T>(path, { method: 'POST', body: body ? JSON.stringify(body) : undefined });
export const put = <T = any>(path: string, body?: any) =>
	api<T>(path, { method: 'PUT', body: body ? JSON.stringify(body) : undefined });
export const del = <T = any>(path: string) => api<T>(path, { method: 'DELETE' });
