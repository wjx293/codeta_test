import { apiRequest } from './client';
import type { OrderListData } from '../types';

export interface OrderQuery {
  status?: number;
  user_id?: number;
  username?: string;
  start_time?: string;
  end_time?: string;
  page?: number;
  page_size?: number;
}

export async function listOrders(query: OrderQuery = {}): Promise<OrderListData> {
  const params = new URLSearchParams();
  if (query.status !== undefined) params.set('status', String(query.status));
  if (query.user_id) params.set('user_id', String(query.user_id));
  if (query.username) params.set('username', query.username);
  if (query.start_time) params.set('start_time', query.start_time);
  if (query.end_time) params.set('end_time', query.end_time);
  params.set('page', String(query.page ?? 1));
  params.set('page_size', String(query.page_size ?? 20));

  return apiRequest<OrderListData>(`/api/admin/orders?${params}`);
}
