import { useCallback, useEffect, useState } from 'react';
import { Search } from 'lucide-react';
import { listOrders } from '../api/orders';
import { ApiError } from '../api/client';
import type { Order } from '../types';
import { ORDER_STATUS_LABEL, ORDER_TYPE_LABEL, formatPrice } from '../types';

const STATUS_OPTIONS = [
  { value: -1, label: '全部' },
  { value: 1, label: '未支付' },
  { value: 2, label: '已获得下载资格' },
  { value: 3, label: '已取消' },
];

export function OrdersPage() {
  const [orders, setOrders] = useState<Order[]>([]);
  const [total, setTotal] = useState(0);
  const [status, setStatus] = useState(-1);
  const [username, setUsername] = useState('');
  const [userId, setUserId] = useState('');
  const [page, setPage] = useState(1);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');

  const pageSize = 15;
  const totalPages = Math.max(1, Math.ceil(total / pageSize));

  const fetchOrders = useCallback(async () => {
    setLoading(true);
    setError('');
    try {
      const data = await listOrders({
        status,
        username: username.trim() || undefined,
        user_id: userId ? Number(userId) : undefined,
        page,
        page_size: pageSize,
      });
      setOrders(data.orders ?? []);
      setTotal(data.total ?? 0);
    } catch (err) {
      setError(err instanceof ApiError ? err.message : '加载失败');
    } finally {
      setLoading(false);
    }
  }, [status, username, userId, page]);

  useEffect(() => {
    fetchOrders();
  }, [fetchOrders]);

  function handleSearch(e: React.FormEvent) {
    e.preventDefault();
    setPage(1);
    fetchOrders();
  }

  return (
    <div className="page">
      <header className="dashboard-head">
        <div>
          <p className="dashboard-greet">查看与管理全部用户订单</p>
          <h2 className="dashboard-title">订单列表</h2>
        </div>
        <form className="head-actions" onSubmit={handleSearch}>
          <label className="filter-pill">
            状态
            <select value={status} onChange={(e) => setStatus(Number(e.target.value))}>
              {STATUS_OPTIONS.map((o) => (
                <option key={o.value} value={o.value}>
                  {o.label}
                </option>
              ))}
            </select>
          </label>
          <label className="filter-pill">
            用户名
            <input
              value={username}
              onChange={(e) => setUsername(e.target.value)}
              placeholder="模糊搜索"
            />
          </label>
          <label className="filter-pill">
            用户 ID
            <input
              value={userId}
              onChange={(e) => setUserId(e.target.value)}
              placeholder="精确匹配"
            />
          </label>
          <button type="submit" className="btn btn-primary">
            <Search size={16} />
            查询
          </button>
        </form>
      </header>

      {error && <div className="alert alert-error">{error}</div>}

      <section className="table-section">
        <div className="table-section-head">
          <h3>全部订单</h3>
          <span>共 {total} 笔</span>
        </div>

        {loading ? (
          <div className="state-box">加载中…</div>
        ) : orders.length === 0 ? (
          <div className="state-box">暂无订单</div>
        ) : (
          <div className="table-wrap">
            <table className="data-table">
              <thead>
                <tr>
                  <th>订单号</th>
                  <th>用户</th>
                  <th>模板</th>
                  <th>类型</th>
                  <th>金额</th>
                  <th>状态</th>
                  <th>创建时间</th>
                </tr>
              </thead>
              <tbody>
                {orders.map((o) => (
                  <tr key={o.id}>
                    <td className="mono">{o.order_no}</td>
                    <td>
                      {o.username}
                      <span className="sub-text">#{o.user_id}</span>
                    </td>
                    <td>{o.template_name}</td>
                    <td>{ORDER_TYPE_LABEL[o.order_type] ?? o.order_type}</td>
                    <td>{formatPrice(o.price_snapshot)}</td>
                    <td>
                      <span className={`status status-${o.status}`}>
                        {ORDER_STATUS_LABEL[o.status] ?? o.status}
                      </span>
                    </td>
                    <td>{new Date(o.created_at).toLocaleString('zh-CN')}</td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )}
      </section>

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
