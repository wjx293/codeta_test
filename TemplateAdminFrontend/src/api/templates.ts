import { apiRequest } from './client';
import type { Template, TemplateListData } from '../types';

export interface CreateTemplatePayload {
  name: string;
  description: string;
  is_free: number;
  price: number;
  file_type: string;
  file_size: number;
  file_oss_key: string;
  thumbnail_oss_key?: string;
  thumbnail_file_size?: number;
}

export interface UpdateTemplatePayload {
  name: string;
  description: string;
  is_free: number;
  price: number;
  thumbnail_oss_key?: string;
  thumbnail_file_size?: number;
}

export async function listTemplates(
  status = -1,
  page = 1,
  pageSize = 20,
): Promise<TemplateListData> {
  const params = new URLSearchParams({
    status: String(status),
    page: String(page),
    page_size: String(pageSize),
  });
  return apiRequest<TemplateListData>(`/api/admin/templates?${params}`);
}

export async function createTemplate(payload: CreateTemplatePayload): Promise<void> {
  await apiRequest<unknown>('/api/admin/templates', { method: 'POST', body: payload });
}

export async function updateTemplate(id: number, payload: UpdateTemplatePayload): Promise<void> {
  await apiRequest<unknown>(`/api/admin/templates/${id}`, { method: 'PUT', body: payload });
}

export async function updateTemplateStatus(id: number, status: number): Promise<void> {
  await apiRequest<unknown>(`/api/admin/templates/${id}/status`, {
    method: 'PUT',
    body: { status },
  });
}

export type { Template };
