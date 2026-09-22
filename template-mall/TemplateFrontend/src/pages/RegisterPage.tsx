import { FormEvent, useState } from 'react';
import { Link, useNavigate } from 'react-router-dom';
import { Sparkles } from 'lucide-react';
import { useAuth } from '../context/AuthContext';
import { ApiError } from '../api/client';

export function RegisterPage() {
  const { register } = useAuth();
  const navigate = useNavigate();
  const [username, setUsername] = useState('');
  const [password, setPassword] = useState('');
  const [confirm, setConfirm] = useState('');
  const [error, setError] = useState('');
  const [submitting, setSubmitting] = useState(false);

  async function handleSubmit(e: FormEvent) {
    e.preventDefault();
    setError('');

    if (password.length < 6) {
      setError('密码至少 6 位');
      return;
    }
    if (password !== confirm) {
      setError('两次输入的密码不一致');
      return;
    }

    setSubmitting(true);
    try {
      await register(username, password);
      navigate('/', { replace: true });
    } catch (err) {
      setError(err instanceof ApiError ? err.message : '注册失败');
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
          <h2>创建账号</h2>
          <p className="auth-subtitle">注册成为会员，享受更多下载权益</p>
          <form onSubmit={handleSubmit} className="auth-form">
            <label>
              用户名
              <input
                value={username}
                onChange={(e) => setUsername(e.target.value)}
                placeholder="3-20 位字母数字"
                autoComplete="username"
                minLength={3}
                maxLength={32}
                required
              />
            </label>
            <label>
              密码
              <input
                type="password"
                value={password}
                onChange={(e) => setPassword(e.target.value)}
                placeholder="至少 6 位"
                autoComplete="new-password"
                minLength={6}
                required
              />
            </label>
            <label>
              确认密码
              <input
                type="password"
                value={confirm}
                onChange={(e) => setConfirm(e.target.value)}
                placeholder="再次输入密码"
                autoComplete="new-password"
                required
              />
            </label>
            {error && <div className="alert alert-error">{error}</div>}
            <button type="submit" className="btn btn-primary btn-block" disabled={submitting}>
              {submitting ? '注册中…' : '注册'}
            </button>
          </form>
          <p className="auth-footer">
            已有账号？<Link to="/login">返回登录</Link>
          </p>
        </div>
      </div>
    </div>
  );
}
