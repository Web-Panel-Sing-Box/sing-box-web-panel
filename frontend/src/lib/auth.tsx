import {
  createContext,
  useCallback,
  useContext,
  useMemo,
  useState,
} from "react";
import * as api from "@/api";

const SESSION_KEY = "sing-grok:auth";

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
      } catch {
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

// Legacy exports for components still being migrated from mock.
// TODO: replace with real API calls (setupTOTP, confirmTOTP).
const MOCK_TOTP_SECRET = "JBSWY3DPEHPK3PXP";
export const TWO_FACTOR_SECRET = MOCK_TOTP_SECRET;

export function buildOtpAuthUri(account = "admin", issuer = "Sing box") {
  const label = encodeURIComponent(`${issuer}:${account}`);
  const params = new URLSearchParams({ secret: MOCK_TOTP_SECRET, issuer });
  return `otpauth://totp/${label}?${params.toString()}`;
}
