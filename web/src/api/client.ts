const API_BASE = '/api';

async function request<T>(path: string, options?: RequestInit): Promise<T> {
  const res = await fetch(`${API_BASE}${path}`, {
    credentials: 'include',
    headers: {
      'Content-Type': 'application/json',
      ...options?.headers,
    },
    ...options,
  });

  if (!res.ok) {
    const body = await res.json().catch(() => ({ error: res.statusText }));
    throw new ApiError(res.status, body.error || 'Request failed');
  }

  return res.json();
}

export class ApiError extends Error {
  status: number;
  constructor(status: number, message: string) {
    super(message);
    this.status = status;
  }
}

export interface User {
  id: string;
  email: string;
  name: string;
  role: string;
  orgId: string;
  orgSlug: string;
  orgName: string;
}

export const api = {
  auth: {
    signup: (data: { email: string; name: string; password: string; orgName: string; orgSlug: string }) =>
      request<User>('/auth/signup', { method: 'POST', body: JSON.stringify(data) }),
    login: (data: { email: string; password: string }) =>
      request<User>('/auth/login', { method: 'POST', body: JSON.stringify(data) }),
    logout: () =>
      request<{ status: string }>('/auth/logout', { method: 'POST' }),
    me: () =>
      request<User>('/auth/me'),
  },
};
