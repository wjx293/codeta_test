import { useCallback, useEffect, useState } from 'react';
import { FileText } from 'lucide-react';
import { cancelOrder, listOrders } from '../api/orders';
import { downloadTemplate } from '../api/templates';
import { useAuth } from '../context/AuthContext';
import { ApiError } from '../api/client';
import type { Order } from '../types';
import { ORDER_STATUS_LABEL, ORDER_TYPE_LABEL, formatPrice } from '../types';

const STATUS_OPTIONS = [
  { value: -1, label: '全部' },
  { value: 1, label: '未支付' },
  { value: 2, label: '已获得下载资格' },
  { value: 3, label: '已取消' },
];

function statusBadge(order: Order): { label: string; className: string } {
  if (order.status === 2) {
    return { label: ORDER_STATUS_LABEL[2], className: 'status-badge-paid' };
  }
  if (order.status === 1) {
    return { label: ORDER_STATUS_LABEL[1], className: 'status-badge-pending' };
  }
  if (order.status === 3) {
    return { label: ORDER_STATUS_LABEL[3], className: 'status-badge-cancel' };
  }
  return { label: ORDER_STATUS_LABEL[order.status] ?? String(order.status), className: 'status-badge-done' };
}

export function OrdersPage() {
  const { accessToken } = useAuth();
  const [orders, setOrders] = useState<Order[]>([]);
  const [total, setTotal] = useState(0);
  const [status, setStatus] = useState(-1);
  const [page, setPage] = useState(1);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');
  const [message, setMessage] = useState('');
  const [actionNo, setActionNo] = useState<string | null>(null);

  const pageSize = 10;
  const totalPages = Math.max(1, Math.ceil(total / pageSize));

  const fetchOrders = useCallback(async () => {
    if (!accessToken) return;
    setLoading(true);
    setError('');
    try {
      const data = await listOrders(accessToken, status, page, pageSize);
      setOrders(data.orders ?? []);
      setTotal(data.total ?? 0);
    } catch (err) {
      setError(err instanceof ApiError ? err.message : '加载订单失败');
    } finally {
      setLoading(false);
    }
  }, [accessToken, status, page]);

  useEffect(() => {
    fetchOrders();
  }, [fetchOrders]);

  useEffect(() => {
    setPage(1);
  }, [status]);

  async function handleCancel(order: Order) {
    if (!accessToken) return;
    if (!window.confirm(`确定取消订单 ${order.order_no}？`)) return;

    setActionNo(order.order_no);
    setMessage('');
    setError('');

    try {
      await cancelOrder(accessToken, order.order_no);
      setMessage('订单已取消');
      await fetchOrders();
    } catch (err) {
      setError(err instanceof ApiError ? err.message : '取消失败');
    } finally {
      setActionNo(null);
    }
  }

  async function handleRedownload(order: Order) {
    if (!accessToken || order.status !== 2) return;

    setActionNo(order.order_no);
    setMessage('');
    setError('');

    try {
      const result = await downloadTemplate(accessToken, order.template_id);
      if (result.download_url) {
        window.open(result.download_url, '_blank', 'noopener,noreferrer');
        setMessage(`「${order.template_name}」下载已开始`);
      } else {
        setError('暂无下载链接');
      }
    } catch (err) {
      setError(err instanceof ApiError ? err.message : '下载失败');
    } finally {
      setActionNo(null);
    }
  }

  return (
    <div className="page">
      <div className="page-head-row">
        <div>
          <h2>我的订单</h2>
          <p>查看购买记录与下载状态</p>
        </div>
        <label className="filter-pill">
          状态筛选
          <select value={status} onChange={(e) => setStatus(Number(e.target.value))}>
            {STATUS_OPTIONS.map((opt) => (
              <option key={opt.value} value={opt.value}>
                {opt.label}
              </option>
            ))}
          </select>
        </label>
      </div>

      {message && <div className="alert alert-success">{message}</div>}
      {error && <div className="alert alert-error">{error}</div>}

      {loading ? (
        <div className="state-box">加载中…</div>
      ) : orders.length === 0 ? (
        <div className="state-box">暂无订单</div>
      ) : (
        <div className="order-list">
          {orders.map((o) => {
            const badge = statusBadge(o);
            const amount =
              o.price_snapshot === 0 ? '免费' : formatPrice(o.price_snapshot);
            return (
              <article key={o.id} className="order-card">
                <div className="order-card-left">
                  <span className="order-icon">
                    <FileText size={22} />
                  </span>
                  <div className="order-info">
                    <h3>{o.template_name}</h3>
                    <p>{o.order_no}</p>
                    <p style={{ marginTop: 4, fontFamily: 'inherit' }}>
                      {ORDER_TYPE_LABEL[o.order_type] ?? o.order_type} ·{' '}
                      {new Date(o.created_at).toLocaleString('zh-CN')}
                    </p>
                  </div>
                </div>
                <div className="order-card-right">
                  <span className="order-amount">{amount}</span>
                  <span className={`status-badge ${badge.className}`}>{badge.label}</span>
                  {o.status === 2 && (
                    <button
                      type="button"
                      className="btn btn-periwinkle btn-sm"
                      disabled={actionNo === o.order_no}
                      onClick={() => handleRedownload(o)}
                    >
                      下载
                    </button>
                  )}
                  {o.status === 1 && o.order_type === 3 && (
                    <button
                      type="button"
                      className="btn btn-coral btn-sm"
                      disabled={actionNo === o.order_no}
                      onClick={() => handleCancel(o)}
                    >
                      取消
                    </button>
                  )}
                </div>
              </article>
            );
          })}
        </div>
      )}

      {totalPages > 1 && (
        <div className="pagination">
          <button
            type="button"
            className="btn btn-ghost"
            disabled={page <= 1}
            onClick={() => setPage((p) => p - 1)}
          >
            上一页
          </button>
          <span>
            第 {page} / {totalPages} 页
          </span>
          <button
            type="button"
            className="btn btn-ghost"
            disabled={page >= totalPages}
            onClick={() => setPage((p) => p + 1)}
          >
            下一页
          </button>
        </div>
      )}
    </div>
  );
}
