import { apiRequest } from './client';
import type { DownloadData, TemplateListData } from '../types';

export async function listTemplates(
  token: string,
  page = 1,
  pageSize = 20,
): Promise<TemplateListData> {
  const params = new URLSearchParams({
    page: String(page),
    page_size: String(pageSize),
  });
  return apiRequest<TemplateListData>(`/templates?${params}`, { token });
}

export async function downloadTemplate(
  token: string,
  templateId: number,
): Promise<DownloadData> {
  return apiRequest<DownloadData>(`/templates/${templateId}/download`, {
    method: 'POST',
    token,
  });
}
