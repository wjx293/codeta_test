import { Navigate, Outlet, Route, Routes } from 'react-router-dom';
import { AuthProvider, useAuth } from './context/AuthContext';
import { Layout } from './components/Layout';
import { LoginPage } from './pages/LoginPage';
import { TemplatesPage } from './pages/TemplatesPage';
import { TemplateFormPage } from './pages/TemplateFormPage';
import { OrdersPage } from './pages/OrdersPage';
import { UsersPage } from './pages/UsersPage';

function ProtectedRoute() {
  const { admin, loading } = useAuth();

  if (loading) {
    return <div className="state-box full-page">加载中…</div>;
  }

  if (!admin) {
    return <Navigate to="/login" replace />;
  }

  return <Outlet />;
}

function PublicOnlyRoute() {
  const { admin, loading } = useAuth();

  if (loading) {
    return <div className="state-box full-page">加载中…</div>;
  }

  if (admin) {
    return <Navigate to="/templates" replace />;
  }

  return <Outlet />;
}

export default function App() {
  return (
    <AuthProvider>
      <Routes>
        <Route element={<PublicOnlyRoute />}>
          <Route path="/login" element={<LoginPage />} />
        </Route>
        <Route element={<ProtectedRoute />}>
          <Route element={<Layout />}>
            <Route path="/" element={<Navigate to="/templates" replace />} />
            <Route path="/templates" element={<TemplatesPage />} />
            <Route path="/templates/new" element={<TemplateFormPage />} />
            <Route path="/templates/:id/edit" element={<TemplateFormPage />} />
            <Route path="/orders" element={<OrdersPage />} />
            <Route path="/users" element={<UsersPage />} />
          </Route>
        </Route>
        <Route path="*" element={<Navigate to="/templates" replace />} />
      </Routes>
    </AuthProvider>
  );
}
