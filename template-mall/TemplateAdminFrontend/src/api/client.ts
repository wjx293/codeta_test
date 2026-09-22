import type { ApiResponse } from '../types';

const API_BASE = import.meta.env.VITE_API_BASE_URL ?? '';

export class ApiError extends Error {
  code: number;
  status: number;

  constructor(message: string, code: number, status: number) {
    super(message);
    this.name = 'ApiError';
    this.code = code;
    this.status = status;
  }
}

type RequestOptions = {
  method?: string;
  body?: unknown;
};

export async function apiRequest<T>(path: string, options: RequestOptions = {}): Promise<T> {
  const { method = 'GET', body } = options;

  const res = await fetch(`${API_BASE}${path}`, {
    method,
    headers: { 'Content-Type': 'application/json' },
    credentials: 'include',
    body: body !== undefined ? JSON.stringify(body) : undefined,
  });

  let json: ApiResponse<T>;
  try {
    json = (await res.json()) as ApiResponse<T>;
  } catch {
    throw new ApiError('响应解析失败', 500, res.status);
  }

  if (!res.ok || json.code !== 0) {
    throw new ApiError(json.message || '请求失败', json.code ?? res.status, res.status);
  }

  return json.data as T;
}

export function getApiBaseUrl(): string {
  return API_BASE || (import.meta.env.DEV ? '' : 'http://localhost:8081');
}

export function getAuthMode(): 'mock' | 'wps' {
  const mode = import.meta.env.VITE_AUTH_MODE || 'mock';
  return mode === 'wps' ? 'wps' : 'mock';
}
