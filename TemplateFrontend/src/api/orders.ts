import { apiRequest } from './client';
import type { OrderListData } from '../types';

export async function listOrders(
  token: string,
  status = -1,
  page = 1,
  pageSize = 20,
): Promise<OrderListData> {
  const params = new URLSearchParams({
    status: String(status),
    page: String(page),
    page_size: String(pageSize),
  });
  return apiRequest<OrderListData>(`/orders?${params}`, { token });
}

export async function cancelOrder(token: string, orderNo: string): Promise<void> {
  await apiRequest<{ message: string }>(`/orders/${orderNo}/cancel`, {
    method: 'POST',
    token,
  });
}

/** 获取用户已支付零售购买的模板 ID（order_type=3, status=2） */
export async function fetchPurchasedTemplateIds(token: string): Promise<Set<number>> {
  const ids = new Set<number>();
  let page = 1;
  const pageSize = 100;

  while (true) {
    const data = await listOrders(token, 2, page, pageSize);
    for (const order of data.orders ?? []) {
      if (order.order_type === 3) {
        ids.add(order.template_id);
      }
    }
    const total = data.total ?? 0;
    if (page * pageSize >= total || !data.orders?.length) {
      break;
    }
    page += 1;
  }

  return ids;
}
