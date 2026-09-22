import {
  createContext,
  useCallback,
  useContext,
  useEffect,
  useMemo,
  useState,
  type ReactNode,
} from 'react';
import * as authApi from '../api/auth';
import { ApiError } from '../api/client';
import type { AdminUser } from '../types';

interface AuthContextValue {
  admin: AdminUser | null;
  loading: boolean;
  mockLogin: () => Promise<void>;
  wpsLogin: () => void;
  logout: () => Promise<void>;
  refreshAdmin: () => Promise<void>;
  isMockMode: boolean;
}

const AuthContext = createContext<AuthContextValue | null>(null);

export function AuthProvider({ children }: { children: ReactNode }) {
  const [admin, setAdmin] = useState<AdminUser | null>(null);
  const [loading, setLoading] = useState(true);
  const isMockMode = authApi.isMockAuthMode();

  const refreshAdmin = useCallback(async () => {
    try {
      const me = await authApi.getCurrentAdmin();
      setAdmin(me);
    } catch (err) {
      if (err instanceof ApiError && (err.status === 401 || err.code === 401)) {
        setAdmin(null);
      }
    }
  }, []);

  useEffect(() => {
    refreshAdmin().finally(() => setLoading(false));
  }, [refreshAdmin]);

  const mockLogin = useCallback(async () => {
    const me = await authApi.mockLogin();
    setAdmin(me);
  }, []);

  const wpsLogin = useCallback(() => {
    authApi.wpsLogin();
  }, []);

  const logout = useCallback(async () => {
    try {
      await authApi.logout();
    } catch {
      // 幂等
    }
    setAdmin(null);
  }, []);

  const value = useMemo(
    () => ({ admin, loading, mockLogin, wpsLogin, logout, refreshAdmin, isMockMode }),
    [admin, loading, mockLogin, wpsLogin, logout, refreshAdmin, isMockMode],
  );

  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>;
}

export function useAuth(): AuthContextValue {
  const ctx = useContext(AuthContext);
  if (!ctx) throw new Error('useAuth must be used within AuthProvider');
  return ctx;
}
