import {
  createContext,
  useCallback,
  useContext,
  useEffect,
  useMemo,
  useState,
} from "react";
import * as api from "@/api";

const SESSION_KEY = "shilka:auth";

type AuthContextValue = {
  isAuthenticated: boolean;
  twoFactorEnabled: boolean;
  login: (username: string, password: string) => Promise<{ ok: boolean; needsTwoFactor: boolean; tempToken?: string; rateLimited?: boolean }>;
  verifyTwoFactor: (tempToken: string, code: string) => Promise<{ ok: boolean; rateLimited?: boolean }>;
  logout: () => void;
  setTwoFactorEnabled: (on: boolean) => void;
};

const AuthContext = createContext<AuthContextValue | null>(null);

function readFlag(on: string) {
  try {
    return window.localStorage.getItem(SESSION_KEY) === on;
  } catch {
    return false;
  }
}

export function AuthProvider({ children }: { children: React.ReactNode }) {
  const [isAuthenticated, setIsAuthenticated] = useState(() => readFlag("1"));
  const [twoFactorEnabled, setTwoFactorState] = useState(false);

  const login = useCallback<AuthContextValue["login"]>(
    async (username, password) => {
      try {
        const res = await api.login({ username, password });
        api.setToken(res.token);
        window.localStorage.setItem(SESSION_KEY, "1");
        setIsAuthenticated(true);
        return { ok: true, needsTwoFactor: false };
      } catch (err) {
        // 2FA enabled: backend returns 403 + { requires_totp, temp_token }.
        if (err instanceof api.ApiError && err.status === 403) {
          const body = err.body as { requires_totp?: boolean; temp_token?: string };
          if (body?.requires_totp) {
            return { ok: true, needsTwoFactor: true, tempToken: body.temp_token };
          }
        }
        // Brute-force limiter: distinguish from bad credentials.
        if (err instanceof api.ApiError && err.status === 429) {
          return { ok: false, needsTwoFactor: false, rateLimited: true };
        }
        return { ok: false, needsTwoFactor: false };
      }
    },
    []
  );

  const verifyTwoFactor = useCallback<AuthContextValue["verifyTwoFactor"]>(
    async (tempToken, code) => {
      try {
        const res = await api.loginTOTP({ temp_token: tempToken, code });
        api.setToken(res.token);
        window.localStorage.setItem(SESSION_KEY, "1");
        setIsAuthenticated(true);
        return { ok: true };
      } catch (err) {
        // Brute-force limiter: distinguish from a wrong code.
        if (err instanceof api.ApiError && err.status === 429) {
          return { ok: false, rateLimited: true };
        }
        return { ok: false };
      }
    },
    []
  );

  const logout = useCallback(() => {
    api.clearToken();
    window.localStorage.removeItem(SESSION_KEY);
    setIsAuthenticated(false);
    api.logout().catch(() => {});
  }, []);

  const setTwoFactorEnabled = useCallback((on: boolean) => {
    setTwoFactorState(on);
  }, []);

  // Hydrate 2FA state from the server so Settings reflects reality across reloads.
  useEffect(() => {
    if (!isAuthenticated) {
      setTwoFactorState(false);
      return;
    }
    let cancelled = false;
    api
      .getMe()
      .then((me) => {
        if (!cancelled) setTwoFactorState(me.is_totp_enabled);
      })
      .catch(() => {});
    return () => {
      cancelled = true;
    };
  }, [isAuthenticated]);

  const value = useMemo<AuthContextValue>(
    () => ({ isAuthenticated, twoFactorEnabled, login, verifyTwoFactor, logout, setTwoFactorEnabled }),
    [isAuthenticated, twoFactorEnabled, login, verifyTwoFactor, logout, setTwoFactorEnabled]
  );

  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>;
}

export function useAuth() {
  const ctx = useContext(AuthContext);
  if (!ctx) throw new Error("useAuth must be used inside <AuthProvider />");
  return ctx;
}
