import { NavLink, Outlet } from 'react-router-dom';
import {
  LayoutGrid,
  LogOut,
  PlusCircle,
  ShoppingBag,
  Users,
} from 'lucide-react';
import { useAuth } from '../context/AuthContext';

const NAV_ITEMS: Array<{
  to: string;
  end?: boolean;
  label: string;
  icon: typeof LayoutGrid;
  cls: string;
}> = [
  { to: '/templates', end: true, label: '模板管理', icon: LayoutGrid, cls: 'nav-templates' },
  { to: '/templates/new', label: '新建模板', icon: PlusCircle, cls: 'nav-new' },
  { to: '/orders', label: '订单列表', icon: ShoppingBag, cls: 'nav-orders' },
  { to: '/users', label: '用户与会员', icon: Users, cls: 'nav-users' },
];

export function Layout() {
  const { admin, logout } = useAuth();

  return (
    <div className="app-shell">
      <aside className="sidebar">
        <div className="brand">
          <span className="brand-mark">ADM</span>
          <div>
            <h1>管理后台</h1>
            <p>模板商城 · B端</p>
          </div>
        </div>

        <nav className="side-nav">
          {NAV_ITEMS.map((item) => (
            <NavLink
              key={item.to}
              to={item.to}
              end={item.end}
              className={({ isActive }) => `${item.cls}${isActive ? ' active' : ''}`}
            >
              <item.icon size={18} />
              {item.label}
            </NavLink>
          ))}
        </nav>

        <div className="sidebar-footer">
          {admin && (
            <div className="admin-card">
              <div className="admin-row">
                <span className="admin-avatar">{admin.username.charAt(0).toUpperCase()}</span>
                <div className="admin-info">
                  <strong>{admin.username}</strong>
                  <span>ID: {admin.user_id}</span>
                </div>
              </div>
              <button type="button" className="btn btn-ghost btn-block btn-sm" onClick={() => logout()}>
                <LogOut size={14} />
                退出登录
              </button>
            </div>
          )}
        </div>
      </aside>

      <div className="main-area">
        <main className="app-main">
          <Outlet />
        </main>
      </div>
    </div>
  );
}
