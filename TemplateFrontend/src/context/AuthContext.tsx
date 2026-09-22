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
import type { AuthData, UserInfo } from '../types';

const STORAGE_KEY = 'template_mall_auth';

interface StoredAuth {
  access_token: string;
  refresh_token: string;
  user_id: number;
  username: string;
  member_status: number;
}

interface AuthContextValue {
  user: UserInfo | null;
  accessToken: string | null;
  loading: boolean;
  login: (username: string, password: string) => Promise<void>;
  register: (username: string, password: string) => Promise<void>;
  logout: () => Promise<void>;
  refreshUser: () => Promise<void>;
}

const AuthContext = createContext<AuthContextValue | null>(null);

function loadStoredAuth(): StoredAuth | null {
  try {
    const raw = localStorage.getItem(STORAGE_KEY);
    if (!raw) return null;
    return JSON.parse(raw) as StoredAuth;
  } catch {
    return null;
  }
}

function saveAuth(data: AuthData) {
  localStorage.setItem(
    STORAGE_KEY,
    JSON.stringify({
      access_token: data.access_token,
      refresh_token: data.refresh_token,
      user_id: data.user_id,
      username: data.username,
      member_status: data.member_status,
    }),
  );
}

function clearAuth() {
  localStorage.removeItem(STORAGE_KEY);
}

export function AuthProvider({ children }: { children: ReactNode }) {
  const [stored, setStored] = useState<StoredAuth | null>(() => loadStoredAuth());
  const [loading, setLoading] = useState(true);

  const user: UserInfo | null = stored
    ? {
        user_id: stored.user_id,
        username: stored.username,
        member_status: stored.member_status,
      }
    : null;

  const accessToken = stored?.access_token ?? null;

  const applyAuth = useCallback((data: AuthData) => {
    saveAuth(data);
    setStored(loadStoredAuth());
  }, []);

  const refreshUser = useCallback(async () => {
    if (!stored?.access_token) return;
    try {
      const me = await authApi.getCurrentUser(stored.access_token);
      const updated: StoredAuth = {
        ...stored,
        member_status: me.member_status,
        username: me.username,
        user_id: me.user_id,
      };
      localStorage.setItem(STORAGE_KEY, JSON.stringify(updated));
      setStored(updated);
    } catch (err) {
      if (err instanceof ApiError && (err.status === 401 || err.code === 401)) {
        if (stored.refresh_token) {
          try {
            const refreshed = await authApi.refreshToken(stored.refresh_token);
            applyAuth(refreshed);
            return;
          } catch {
            clearAuth();
            setStored(null);
          }
        } else {
          clearAuth();
          setStored(null);
        }
      }
    }
  }, [stored, applyAuth]);

  useEffect(() => {
    refreshUser().finally(() => setLoading(false));
  }, []); // eslint-disable-line react-hooks/exhaustive-deps

  const login = useCallback(
    async (username: string, password: string) => {
      const data = await authApi.login(username, password);
      applyAuth(data);
    },
    [applyAuth],
  );

  const register = useCallback(
    async (username: string, password: string) => {
      const data = await authApi.register(username, password);
      applyAuth(data);
    },
    [applyAuth],
  );

  const logout = useCallback(async () => {
    if (stored?.access_token) {
      try {
        await authApi.logout(stored.access_token);
      } catch {
        // 幂等登出，忽略错误
      }
    }
    clearAuth();
    setStored(null);
  }, [stored]);

  const value = useMemo(
    () => ({
      user,
      accessToken,
      loading,
      login,
      register,
      logout,
      refreshUser,
    }),
    [user, accessToken, loading, login, register, logout, refreshUser],
  );

  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>;
}

export function useAuth(): AuthContextValue {
  const ctx = useContext(AuthContext);
  if (!ctx) throw new Error('useAuth must be used within AuthProvider');
  return ctx;
}
