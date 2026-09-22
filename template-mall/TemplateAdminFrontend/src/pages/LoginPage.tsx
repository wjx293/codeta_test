import { useAuth } from '../context/AuthContext';

export function LoginPage() {
  const { wpsLogin } = useAuth();

  return (
    <div className="auth-scene">
      <span className="auth-blob auth-blob-1" aria-hidden />
      <span className="auth-blob auth-blob-2" aria-hidden />

      <div className="auth-wrap">
        <div className="auth-brand-mini">
          <span className="brand-mark">ADM</span>
          <strong style={{ fontFamily: 'var(--font-head)', fontSize: '1.125rem' }}>管理后台</strong>
        </div>

        <div className="auth-card">
          <h2>管理员登录</h2>
          <p className="auth-subtitle">使用 WPS 企业账号登录管理后台</p>

          <div className="auth-actions">
            <button type="button" className="btn btn-primary btn-block" onClick={wpsLogin}>
              WPS 企业账号登录
            </button>
          </div>

          <p className="auth-hint">
            WPS 回调地址须与开放平台一致：http://localhost:3001/api/auth/callback
          </p>
        </div>
      </div>
    </div>
  );
}
