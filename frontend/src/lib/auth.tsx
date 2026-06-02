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
  login: (username: string, password: string) => Promise<{ ok: boolean; needsTwoFactor: boolean; tempToken?: string }>;
  verifyTwoFactor: (tempToken: string, code: string) => Promise<boolean>;
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

  useEffect(() => {
    if (!isAuthenticated) return;
    api.getMe().then((me) => setTwoFactorState(me.is_totp_enabled)).catch(() => {});
  }, [isAuthenticated]);

  const login = useCallback<AuthContextValue["login"]>(
    async (username, password) => {
      try {
        const res = await api.login({ username, password });

        if ("requires_totp" in res && res.requires_totp) {
          return { ok: true, needsTwoFactor: true, tempToken: res.temp_token };
        }

        if ("token" in res) {
          api.setToken(res.token);
          window.localStorage.setItem(SESSION_KEY, "1");
          setIsAuthenticated(true);
          return { ok: true, needsTwoFactor: false };
        }

        return { ok: false, needsTwoFactor: false };
      } catch (e: unknown) {
        if (e instanceof api.ApiError && e.status === 403) {
          const body = e.body as { requires_totp?: boolean; temp_token?: string } | undefined;
          if (body?.requires_totp) {
            return { ok: true, needsTwoFactor: true, tempToken: body.temp_token };
          }
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
        return true;
      } catch {
        return false;
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
