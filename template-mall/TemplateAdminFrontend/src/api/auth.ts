import { apiRequest, getAuthMode } from './client';
import type { AdminUser } from '../types';

export async function mockLogin(): Promise<AdminUser> {
  return apiRequest<AdminUser>('/api/auth/mock-login', { method: 'POST' });
}

export function wpsLogin(): void {
  window.location.href = '/api/auth/login';
}

export async function logout(): Promise<void> {
  await apiRequest<{ message: string }>('/api/auth/logout', { method: 'POST' });
}

export async function getCurrentAdmin(): Promise<AdminUser> {
  return apiRequest<AdminUser>('/api/auth/me');
}

export function isMockAuthMode(): boolean {
  return getAuthMode() === 'mock';
}
