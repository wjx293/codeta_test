import { apiRequest } from './client';
import type { AuthData } from '../types';

export async function register(username: string, password: string): Promise<AuthData> {
  return apiRequest<AuthData>('/auth/register', {
    method: 'POST',
    body: { username, password },
    skipAuth: true,
  });
}

export async function login(username: string, password: string): Promise<AuthData> {
  return apiRequest<AuthData>('/auth/login', {
    method: 'POST',
    body: { username, password },
    skipAuth: true,
  });
}

export async function refreshToken(refreshToken: string): Promise<AuthData> {
  return apiRequest<AuthData>('/auth/refresh', {
    method: 'POST',
    body: { refresh_token: refreshToken },
    skipAuth: true,
  });
}

export async function logout(token: string): Promise<void> {
  await apiRequest<{ message: string }>('/auth/logout', {
    method: 'POST',
    token,
  });
}

export async function getCurrentUser(token: string) {
  return apiRequest<import('../types').UserInfo>('/users/me', { token });
}
