import { FormEvent, useState } from 'react';
import { Link, useNavigate } from 'react-router-dom';
import { Sparkles } from 'lucide-react';
import { useAuth } from '../context/AuthContext';
import { ApiError } from '../api/client';

export function LoginPage() {
  const { login } = useAuth();
  const navigate = useNavigate();
  const [username, setUsername] = useState('');
  const [password, setPassword] = useState('');
  const [error, setError] = useState('');
  const [submitting, setSubmitting] = useState(false);

  async function handleSubmit(e: FormEvent) {
    e.preventDefault();
    setError('');
    setSubmitting(true);
    try {
      await login(username, password);
      navigate('/', { replace: true });
    } catch (err) {
      setError(err instanceof ApiError ? err.message : '登录失败');
    } finally {
      setSubmitting(false);
    }
  }

  return (
    <div className="auth-scene">
      <span className="auth-blob auth-blob-1" aria-hidden />
      <span className="auth-blob auth-blob-2" aria-hidden />
      <span className="auth-blob auth-blob-3" aria-hidden />

      <div className="auth-wrap">
        <div className="auth-brand-mini">
          <span className="brand-mark">
            <Sparkles size={20} />
          </span>
          <strong style={{ fontFamily: 'var(--font-head)', fontSize: '1.125rem' }}>模板商城</strong>
        </div>

        <div className="auth-card">
          <h2>欢迎回来</h2>
          <p className="auth-subtitle">登录后即可下载模板与管理订单</p>
          <form onSubmit={handleSubmit} className="auth-form">
            <label>
              用户名
              <input
                value={username}
                onChange={(e) => setUsername(e.target.value)}
                placeholder="请输入用户名"
                autoComplete="username"
                required
              />
            </label>
            <label>
              密码
              <input
                type="password"
                value={password}
                onChange={(e) => setPassword(e.target.value)}
                placeholder="请输入密码"
                autoComplete="current-password"
                required
              />
            </label>
            {error && <div className="alert alert-error">{error}</div>}
            <button type="submit" className="btn btn-primary btn-block" disabled={submitting}>
              {submitting ? '登录中…' : '登录'}
            </button>
          </form>
          <p className="auth-footer">
            还没有账号？<Link to="/register">立即注册</Link>
          </p>
        </div>
      </div>
    </div>
  );
}
