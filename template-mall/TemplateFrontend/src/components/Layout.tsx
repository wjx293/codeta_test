import { NavLink, Outlet } from 'react-router-dom';
import { Bell, Search, Sparkles } from 'lucide-react';
import { useAuth } from '../context/AuthContext';

export function Layout() {
  const { user, logout } = useAuth();

  return (
    <div className="app-shell">
      <div className="header-wrap">
        <header className="soft-nav">
          <div className="brand">
            <span className="brand-mark">
              <Sparkles size={22} />
            </span>
            <div>
              <h1>模板商城</h1>
              <p>简约美学 · 从容办公</p>
            </div>
          </div>

          <nav className="nav-center">
            <NavLink to="/" end className={({ isActive }) => `nav-pill${isActive ? ' active' : ''}`}>
              模板列表
            </NavLink>
            <NavLink
              to="/orders"
              className={({ isActive }) => `nav-pill orders${isActive ? ' active' : ''}`}
            >
              我的订单
            </NavLink>
          </nav>

          <div className="nav-right">
            <label className="search-bar">
              <Search size={16} />
              <input type="search" placeholder="搜索模板…" aria-label="搜索模板" />
            </label>
            <button type="button" className="icon-btn" aria-label="通知">
              <Bell size={20} />
            </button>
            {user && (
              <div className="user-bar">
                <span className="user-avatar">{user.username.charAt(0).toUpperCase()}</span>
                <span className="user-name">{user.username}</span>
                <span className={`badge ${user.member_status === 1 ? 'badge-member' : 'badge-paid'}`}>
                  {user.member_status === 1 ? '会员' : '普通用户'}
                </span>
                <button type="button" className="btn btn-ghost btn-sm" onClick={() => logout()}>
                  退出
                </button>
              </div>
            )}
          </div>
        </header>
      </div>
      <main className="app-main">
        <Outlet />
      </main>
    </div>
  );
}
