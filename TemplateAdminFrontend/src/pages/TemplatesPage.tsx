import { useCallback, useEffect, useMemo, useState } from 'react';
import { Link } from 'react-router-dom';
import {
  BarChart3,
  Crown,
  LayoutGrid,
  Search,
  ShoppingCart,
  Upload,
  Users,
} from 'lucide-react';
import { listTemplates, updateTemplateStatus } from '../api/templates';
import { listOrders } from '../api/orders';
import { listUsers } from '../api/users';
import { useAuth } from '../context/AuthContext';
import { ApiError } from '../api/client';
import type { Template } from '../types';
import { TEMPLATE_STATUS_LABEL, formatFileSize, formatPrice } from '../types';

const STATUS_OPTIONS = [
  { value: -1, label: '全部' },
  { value: 1, label: '已上架' },
  { value: 0, label: '已下架' },
];

interface DashboardStats {
  onShelf: number;
  todayOrders: number;
  totalUsers: number;
  memberUsers: number;
}

function todayRange() {
  const start = new Date();
  start.setHours(0, 0, 0, 0);
  const end = new Date();
  end.setHours(23, 59, 59, 999);
  return { start: start.toISOString(), end: end.toISOString() };
}

function greeting(name: string) {
  const hour = new Date().getHours();
  const period = hour < 12 ? '上午好' : hour < 18 ? '下午好' : '晚上好';
  return `${period}，${name} ☀️`;
}

export function TemplatesPage() {
  const { admin } = useAuth();
  const [templates, setTemplates] = useState<Template[]>([]);
  const [total, setTotal] = useState(0);
  const [status, setStatus] = useState(-1);
  const [page, setPage] = useState(1);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');
  const [message, setMessage] = useState('');
  const [actionId, setActionId] = useState<number | null>(null);
  const [search, setSearch] = useState('');
  const [stats, setStats] = useState<DashboardStats>({
    onShelf: 0,
    todayOrders: 0,
    totalUsers: 0,
    memberUsers: 0,
  });

  const pageSize = 10;
  const totalPages = Math.max(1, Math.ceil(total / pageSize));

  const fetchStats = useCallback(async () => {
    try {
      const { start, end } = todayRange();
      const [onShelfData, todayOrderData, userData] = await Promise.all([
        listTemplates(1, 1, 1),
        listOrders({ start_time: start, end_time: end, page: 1, page_size: 1 }),
        listUsers('', 1, 200),
      ]);
      const members = (userData.users ?? []).filter((u) => u.member_status === 1).length;
      setStats({
        onShelf: onShelfData.total ?? 0,
        todayOrders: todayOrderData.total ?? 0,
        totalUsers: userData.total ?? 0,
        memberUsers: members,
      });
    } catch {
      // 统计失败不影响主列表
    }
  }, []);

  const fetchTemplates = useCallback(async () => {
    setLoading(true);
    setError('');
    try {
      const data = await listTemplates(status, page, pageSize);
      setTemplates(data.templates ?? []);
      setTotal(data.total ?? 0);
    } catch (err) {
      setError(err instanceof ApiError ? err.message : '加载失败');
    } finally {
      setLoading(false);
    }
  }, [status, page]);

  useEffect(() => {
    fetchStats();
  }, [fetchStats]);

  useEffect(() => {
    fetchTemplates();
  }, [fetchTemplates]);

  useEffect(() => {
    setPage(1);
  }, [status]);

  const filteredTemplates = useMemo(() => {
    const q = search.trim().toLowerCase();
    if (!q) return templates;
    return templates.filter((t) => t.name.toLowerCase().includes(q));
  }, [templates, search]);

  async function handleToggleStatus(template: Template) {
    const newStatus = template.status === 1 ? 0 : 1;
    const action = newStatus === 1 ? '上架' : '下架';
    if (!window.confirm(`确定${action}「${template.name}」？`)) return;

    setActionId(template.id);
    setMessage('');
    setError('');

    try {
      await updateTemplateStatus(template.id, newStatus);
      setMessage(`已${action}「${template.name}」`);
      await Promise.all([fetchTemplates(), fetchStats()]);
    } catch (err) {
      setError(err instanceof ApiError ? err.message : `${action}失败`);
    } finally {
      setActionId(null);
    }
  }

  const statCards = [
    {
      label: '上架模板',
      value: stats.onShelf,
      bg: 'var(--primary-soft)',
      icon: LayoutGrid,
      iconColor: 'var(--primary-deep)',
      trend: 'var(--primary-deep)',
    },
    {
      label: '今日订单',
      value: stats.todayOrders,
      bg: 'var(--coral-soft)',
      icon: ShoppingCart,
      iconColor: 'var(--coral)',
      trend: 'var(--coral)',
    },
    {
      label: '活跃用户',
      value: stats.totalUsers,
      bg: 'var(--sage-soft)',
      icon: Users,
      iconColor: 'var(--sage)',
      trend: 'var(--sage)',
    },
    {
      label: '会员用户',
      value: stats.memberUsers,
      bg: 'var(--butter-soft)',
      icon: Crown,
      iconColor: 'var(--butter)',
      trend: 'var(--butter)',
    },
  ] as const;

  return (
    <div className="page">
      <header className="dashboard-head">
        <div>
          <p className="dashboard-greet">{greeting(admin?.username ?? 'admin')}</p>
          <h2 className="dashboard-title">运营概览</h2>
        </div>
        <div className="head-actions">
          <label className="search-bar">
            <Search size={16} />
            <input
              type="search"
              placeholder="搜索模板或订单…"
              value={search}
              onChange={(e) => setSearch(e.target.value)}
              aria-label="搜索模板"
            />
          </label>
          <Link to="/templates/new" className="btn btn-primary">
            + 新建模板
          </Link>
        </div>
      </header>

      <div className="stat-grid">
        {statCards.map((s) => (
          <article key={s.label} className="stat-card" style={{ background: s.bg }}>
            <div className="stat-card-top">
              <span className="stat-orb">
                <s.icon size={22} color={s.iconColor} />
              </span>
              <span className="stat-trend" style={{ color: s.trend }}>
                ↑ 实时
              </span>
            </div>
            <p className="stat-value">{s.value}</p>
            <p className="stat-label">{s.label}</p>
          </article>
        ))}
      </div>

      <section className="quick-section">
        <h3>快捷操作</h3>
        <div className="quick-actions">
          <Link to="/templates/new" className="btn btn-sage">
            <Upload size={16} />
            上传模板
          </Link>
          <Link to="/orders" className="btn btn-coral">
            <ShoppingCart size={16} />
            处理订单
          </Link>
          <Link to="/users" className="btn btn-butter">
            <Crown size={16} />
            会员管理
          </Link>
          <button type="button" className="btn btn-periwinkle" disabled>
            <BarChart3 size={16} />
            数据报表
          </button>
        </div>
      </section>

      <section className="table-section">
        <div className="table-section-head">
          <h3>最近模板</h3>
          <div className="header-actions">
            <label>
              状态
              <select value={status} onChange={(e) => setStatus(Number(e.target.value))}>
                {STATUS_OPTIONS.map((o) => (
                  <option key={o.value} value={o.value}>
                    {o.label}
                  </option>
                ))}
              </select>
            </label>
            <span>共 {total} 个</span>
          </div>
        </div>

        {message && <div className="alert alert-success" style={{ margin: '0 26px 12px' }}>{message}</div>}
        {error && <div className="alert alert-error" style={{ margin: '0 26px 12px' }}>{error}</div>}

        {loading ? (
          <div className="state-box">加载中…</div>
        ) : filteredTemplates.length === 0 ? (
          <div className="state-box">暂无模板</div>
        ) : (
          <div className="table-wrap">
            <table className="data-table">
              <thead>
                <tr>
                  <th>编号</th>
                  <th>名称</th>
                  <th>类型</th>
                  <th>定价</th>
                  <th>大小</th>
                  <th>状态</th>
                  <th>操作</th>
                </tr>
              </thead>
              <tbody>
                {filteredTemplates.map((t, index) => (
                  <tr key={t.id}>
                    <td>{(page - 1) * pageSize + index + 1}</td>
                    <td>{t.name}</td>
                    <td>{t.file_type}</td>
                    <td>{t.is_free === 1 ? '免费' : formatPrice(t.price)}</td>
                    <td>{formatFileSize(t.file_size)}</td>
                    <td>
                      <span className={`status status-tpl-${t.status}`}>
                        {TEMPLATE_STATUS_LABEL[t.status] ?? t.status}
                      </span>
                    </td>
                    <td className="actions">
                      <Link
                        to={`/templates/${t.id}/edit`}
                        state={{ template: t }}
                        className="btn btn-sm btn-ghost"
                      >
                        编辑
                      </Link>
                      <button
                        type="button"
                        className={`btn btn-sm ${t.status === 1 ? 'btn-danger' : 'btn-primary'}`}
                        disabled={actionId === t.id}
                        onClick={() => handleToggleStatus(t)}
                      >
                        {t.status === 1 ? '下架' : '上架'}
                      </button>
                    </td>
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
