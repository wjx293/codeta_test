import { Navigate, Outlet, Route, Routes } from 'react-router-dom';
import { AuthProvider, useAuth } from './context/AuthContext';
import { Layout } from './components/Layout';
import { LoginPage } from './pages/LoginPage';
import { RegisterPage } from './pages/RegisterPage';
import { TemplatesPage } from './pages/TemplatesPage';
import { OrdersPage } from './pages/OrdersPage';

function ProtectedRoute() {
  const { accessToken, loading } = useAuth();

  if (loading) {
    return <div className="state-box full-page">加载中…</div>;
  }

  if (!accessToken) {
    return <Navigate to="/login" replace />;
  }

  return <Outlet />;
}

function PublicOnlyRoute() {
  const { accessToken, loading } = useAuth();

  if (loading) {
    return <div className="state-box full-page">加载中…</div>;
  }

  if (accessToken) {
    return <Navigate to="/" replace />;
  }

  return <Outlet />;
}

export default function App() {
  return (
    <AuthProvider>
      <Routes>
        <Route element={<PublicOnlyRoute />}>
          <Route path="/login" element={<LoginPage />} />
          <Route path="/register" element={<RegisterPage />} />
        </Route>
        <Route element={<ProtectedRoute />}>
          <Route element={<Layout />}>
            <Route index element={<TemplatesPage />} />
            <Route path="orders" element={<OrdersPage />} />
          </Route>
        </Route>
        <Route path="*" element={<Navigate to="/" replace />} />
      </Routes>
    </AuthProvider>
  );
}
