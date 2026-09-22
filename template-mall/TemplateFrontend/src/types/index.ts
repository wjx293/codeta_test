export interface ApiResponse<T> {
  code: number;
  message?: string;
  data?: T;
}

export interface AuthData {
  access_token: string;
  refresh_token: string;
  user_id: number;
  username: string;
  member_status: number;
}

export interface UserInfo {
  user_id: number;
  username: string;
  member_status: number;
}

export interface Template {
  id: number;
  template_no: string;
  name: string;
  description: string;
  is_free: number;
  price: number;
  status: number;
  file_type: string;
  file_size: number;
  thumbnail_download_url: string;
  created_at: string;
  updated_at: string;
}

export interface TemplateListData {
  templates: Template[];
  total: number;
}

export interface DownloadData {
  order_no: string;
  order_type: number;
  order_status: number;
  price_snapshot: number;
  download_url: string;
  payment_no: string;
}

export interface Order {
  id: number;
  order_no: string;
  user_id: number;
  username: string;
  template_id: number;
  template_name: string;
  order_type: number;
  price_snapshot: number;
  status: number;
  created_at: string;
  updated_at: string;
}

export interface OrderListData {
  orders: Order[];
  total: number;
}

export const ORDER_TYPE_LABEL: Record<number, string> = {
  1: '免费',
  2: '会员',
  3: '零售',
};

export const ORDER_STATUS_LABEL: Record<number, string> = {
  1: '未支付',
  2: '已获得下载资格',
  3: '已取消',
};

export function formatPrice(cents: number): string {
  return `¥${(cents / 100).toFixed(2)}`;
}

export function formatFileSize(bytes: number): string {
  if (bytes < 1024) return `${bytes} B`;
  if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(1)} KB`;
  return `${(bytes / (1024 * 1024)).toFixed(2)} MB`;
}
