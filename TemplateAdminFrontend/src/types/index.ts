export interface ApiResponse<T> {
  code: number;
  message?: string;
  data?: T;
}

export interface AdminUser {
  user_id: string;
  username: string;
  avatar?: string;
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

export interface User {
  id: number;
  username: string;
  member_status: number;
  created_at: string;
}

export interface UserListData {
  users: User[];
  total: number;
}

export interface UploadCredential {
  object_key: string;
  access_key_id: string;
  access_key_secret: string;
  security_token: string;
  expiration: string;
  bucket: string;
  endpoint: string;
  region: string;
}

export interface ConfirmUploadData {
  object_key: string;
  file_size: number;
  confirmed: boolean;
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

export const TEMPLATE_STATUS_LABEL: Record<number, string> = {
  0: '已下架',
  1: '已上架',
};

export function formatPrice(cents: number): string {
  return `¥${(cents / 100).toFixed(2)}`;
}

export function formatFileSize(bytes: number): string {
  if (bytes < 1024) return `${bytes} B`;
  if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(1)} KB`;
  return `${(bytes / (1024 * 1024)).toFixed(2)} MB`;
}

export function getFileExtension(filename: string): string {
  const parts = filename.split('.');
  return parts.length > 1 ? parts.pop()!.toLowerCase() : '';
}
