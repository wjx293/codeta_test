import { apiRequest } from './client';
import type { UserListData } from '../types';

export async function listUsers(
  username = '',
  page = 1,
  pageSize = 20,
): Promise<UserListData> {
  const params = new URLSearchParams({
    page: String(page),
    page_size: String(pageSize),
  });
  if (username) params.set('username', username);

  return apiRequest<UserListData>(`/api/admin/users?${params}`);
}

export async function setMember(userId: number, memberStatus: number): Promise<void> {
  await apiRequest<unknown>(`/api/admin/users/${userId}/member`, {
    method: 'PUT',
    body: { member_status: memberStatus },
  });
}
