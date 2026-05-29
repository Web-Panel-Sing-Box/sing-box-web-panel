import {
  createContext,
  useCallback,
  useContext,
  useMemo,
  useState
} from "react";

const SESSION_KEY = "sing-grok:auth";
const TWOFA_KEY = "sing-grok:twofa";

// Mock credentials for the in-memory build. Swap for real /api/auth/* later.
const MOCK_USERNAME = "admin";
const MOCK_PASSWORD = "admin";
const MOCK_TOTP = "123456";
const MOCK_TOTP_SECRET = "JBSWY3DPEHPK3PXP";

export const DEMO_USERNAME = MOCK_USERNAME;
export const DEMO_PASSWORD = MOCK_PASSWORD;
export const DEMO_TOTP = MOCK_TOTP;
export const TWO_FACTOR_SECRET = MOCK_TOTP_SECRET;

/** Builds the otpauth:// URI an authenticator app would scan. */
export function buildOtpAuthUri(account = MOCK_USERNAME, issuer = "Sing box") {
  const label = encodeURIComponent(`${issuer}:${account}`);
  const params = new URLSearchParams({
    secret: MOCK_TOTP_SECRET,
    issuer
  });
  return `otpauth://totp/${label}?${params.toString()}`;
}

type LoginResult = { ok: boolean; needsTwoFactor: boolean };

type AuthContextValue = {
  isAuthenticated: boolean;
  twoFactorEnabled: boolean;
  login: (username: string, password: string) => LoginResult;
  verifyTwoFactor: (code: string) => boolean;
  logout: () => void;
  setTwoFactorEnabled: (on: boolean) => void;
};

const AuthContext = createContext<AuthContextValue | null>(null);

function readFlag(key: string, on: string) {
  if (typeof window === "undefined") return false;
  return window.localStorage.getItem(key) === on;
}

export function AuthProvider({ children }: { children: React.ReactNode }) {
  const [isAuthenticated, setIsAuthenticated] = useState(() => readFlag(SESSION_KEY, "1"));
  const [twoFactorEnabled, setTwoFactorState] = useState(() => readFlag(TWOFA_KEY, "on"));

  const startSession = useCallback(() => {
    if (typeof window !== "undefined") window.localStorage.setItem(SESSION_KEY, "1");
    setIsAuthenticated(true);
  }, []);

  const login = useCallback<AuthContextValue["login"]>(
    (username, password) => {
      if (username !== MOCK_USERNAME || password !== MOCK_PASSWORD) {
        return { ok: false, needsTwoFactor: false };
      }
      if (twoFactorEnabled) {
        // Hold off on the session — the LoginPage drives the second step.
        return { ok: true, needsTwoFactor: true };
      }
      startSession();
      return { ok: true, needsTwoFactor: false };
    },
    [twoFactorEnabled, startSession]
  );

  const verifyTwoFactor = useCallback<AuthContextValue["verifyTwoFactor"]>(
    (code) => {
      if (code !== MOCK_TOTP) return false;
      startSession();
      return true;
    },
    [startSession]
  );

  const logout = useCallback(() => {
    if (typeof window !== "undefined") window.localStorage.removeItem(SESSION_KEY);
    setIsAuthenticated(false);
  }, []);

  const setTwoFactorEnabled = useCallback<AuthContextValue["setTwoFactorEnabled"]>((on) => {
    if (typeof window !== "undefined") {
      if (on) window.localStorage.setItem(TWOFA_KEY, "on");
      else window.localStorage.removeItem(TWOFA_KEY);
    }
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
