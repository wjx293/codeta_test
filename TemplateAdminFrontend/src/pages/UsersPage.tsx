import { FormEvent, useCallback, useEffect, useMemo, useState } from 'react';
import { Crown, Search, UserPlus, Users } from 'lucide-react';
import { listUsers, setMember } from '../api/users';
import { ApiError } from '../api/client';
import type { User } from '../types';

export function UsersPage() {
  const [users, setUsers] = useState<User[]>([]);
  const [total, setTotal] = useState(0);
  const [keyword, setKeyword] = useState('');
  const [search, setSearch] = useState('');
  const [page, setPage] = useState(1);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');
  const [message, setMessage] = useState('');
  const [actionId, setActionId] = useState<number | null>(null);

  const pageSize = 15;
  const totalPages = Math.max(1, Math.ceil(total / pageSize));

  const fetchUsers = useCallback(async () => {
    setLoading(true);
    setError('');
    try {
      const data = await listUsers(search, page, pageSize);
      setUsers(data.users ?? []);
      setTotal(data.total ?? 0);
    } catch (err) {
      setError(err instanceof ApiError ? err.message : '加载失败');
    } finally {
      setLoading(false);
    }
  }, [search, page]);

  useEffect(() => {
    fetchUsers();
  }, [fetchUsers]);

  const memberCount = useMemo(
    () => users.filter((u) => u.member_status === 1).length,
    [users],
  );

  function handleSearch(e: FormEvent) {
    e.preventDefault();
    setPage(1);
    setSearch(keyword.trim());
  }

  async function handleToggleMember(user: User) {
    const newStatus = user.member_status === 1 ? 0 : 1;
    const action = newStatus === 1 ? '设为会员' : '取消会员';
    if (!window.confirm(`确定将「${user.username}」${action}？`)) return;

    setActionId(user.id);
    setMessage('');
    setError('');

    try {
      await setMember(user.id, newStatus);
      setMessage(`已${action}：${user.username}`);
      await fetchUsers();
    } catch (err) {
      setError(err instanceof ApiError ? err.message : '操作失败');
    } finally {
      setActionId(null);
    }
  }

  const statCards = [
    {
      label: '注册用户',
      value: total,
      bg: 'var(--primary-soft)',
      icon: Users,
      iconColor: 'var(--primary-deep)',
      trend: 'var(--primary-deep)',
    },
    {
      label: '会员用户',
      value: memberCount,
      bg: 'var(--butter-soft)',
      icon: Crown,
      iconColor: 'var(--butter)',
      trend: 'var(--butter)',
    },
    {
      label: '本页展示',
      value: users.length,
      bg: 'var(--sage-soft)',
      icon: UserPlus,
      iconColor: 'var(--sage)',
      trend: 'var(--sage)',
    },
  ] as const;

  return (
    <div className="page">
      <header className="dashboard-head">
        <div>
          <p className="dashboard-greet">管理用户会员状态与权限</p>
          <h2 className="dashboard-title">用户与会员</h2>
        </div>
        <form className="head-actions" onSubmit={handleSearch}>
          <label className="search-bar">
            <Search size={16} />
            <input
              value={keyword}
              onChange={(e) => setKeyword(e.target.value)}
              placeholder="搜索用户…"
              aria-label="搜索用户"
            />
          </label>
          <button type="submit" className="btn btn-butter">
            搜索
          </button>
        </form>
      </header>

      <div className="stat-grid stat-grid-3">
        {statCards.map((s) => (
          <article key={s.label} className="stat-card" style={{ background: s.bg }}>
            <div className="stat-card-top">
              <span className="stat-orb">
                <s.icon size={22} color={s.iconColor} />
              </span>
              <span className="stat-trend" style={{ color: s.trend }}>
                实时
              </span>
            </div>
            <p className="stat-value">{s.value}</p>
            <p className="stat-label">{s.label}</p>
          </article>
        ))}
      </div>

      {message && <div className="alert alert-success">{message}</div>}
      {error && <div className="alert alert-error">{error}</div>}

      <section className="table-section">
        <div className="table-section-head">
          <h3>用户列表</h3>
          <span>共 {total} 人</span>
        </div>

        {loading ? (
          <div className="state-box">加载中…</div>
        ) : users.length === 0 ? (
          <div className="state-box">暂无用户</div>
        ) : (
          <div className="table-wrap">
            <table className="data-table">
              <thead>
                <tr>
                  <th>ID</th>
                  <th>用户名</th>
                  <th>会员状态</th>
                  <th>注册时间</th>
                  <th>操作</th>
                </tr>
              </thead>
              <tbody>
                {users.map((u) => (
                  <tr key={u.id}>
                    <td>{u.id}</td>
                    <td>{u.username}</td>
                    <td>
                      <span className={`badge ${u.member_status === 1 ? 'badge-member' : ''}`}>
                        {u.member_status === 1 ? '会员' : '普通用户'}
                      </span>
                    </td>
                    <td>{new Date(u.created_at).toLocaleString('zh-CN')}</td>
                    <td className="actions">
                      <button
                        type="button"
                        className={`btn btn-sm ${u.member_status === 1 ? 'btn-danger' : 'btn-primary'}`}
                        disabled={actionId === u.id}
                        onClick={() => handleToggleMember(u)}
                      >
                        {u.member_status === 1 ? '取消会员' : '设为会员'}
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
