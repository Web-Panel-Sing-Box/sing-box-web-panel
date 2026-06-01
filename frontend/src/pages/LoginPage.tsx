import { useState, type FormEvent } from "react";
import { Navigate, useLocation, useNavigate } from "react-router-dom";
import { AnimatePresence, m } from "framer-motion";

import { Button } from "@/components/ui/button";
import { Card } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { useToast } from "@/components/ui/toast";
import { useAuth } from "@/lib/auth";
import { useI18n } from "@/lib/i18n";

function BrandMark() {
  return (
    <div className="grid size-11 place-items-center rounded-xl bg-white/5 text-ink-primary">
      <svg viewBox="0 0 24 24" width="22" height="22" fill="none">
        <path
          d="M5 8.5 12 5l7 3.5v7L12 19l-7-3.5v-7Z"
          stroke="currentColor"
          strokeWidth="1.5"
        />
        <path
          d="M5 8.5 12 12l7-3.5M12 12v7"
          stroke="currentColor"
          strokeWidth="1.5"
        />
      </svg>
    </div>
  );
}

type Step = "credentials" | "twofactor";

export function LoginPage() {
  const { t } = useI18n();
  const { push } = useToast();
  const { isAuthenticated, login, verifyTwoFactor } = useAuth();
  const navigate = useNavigate();
  const location = useLocation();

  const from =
    (location.state as { from?: { pathname?: string } } | null)?.from
      ?.pathname ?? "/";

  const [step, setStep] = useState<Step>("credentials");
  const [username, setUsername] = useState("");
  const [password, setPassword] = useState("");
  const [code, setCode] = useState("");

  if (isAuthenticated) {
    return <Navigate to={from} replace />;
  }

  const handleCredentials = (e: FormEvent) => {
    e.preventDefault();
    const result = login(username.trim(), password);
    if (!result.ok) {
      push(t("login.invalidCredentials"), "error");
      return;
    }
    if (result.needsTwoFactor) {
      setCode("");
      setStep("twofactor");
      return;
    }
    navigate(from, { replace: true });
  };

  const handleVerify = (e: FormEvent) => {
    e.preventDefault();
    if (verifyTwoFactor(code.trim())) {
      navigate(from, { replace: true });
      return;
    }
    push(t("login.invalidCode"), "error");
  };

  return (
    <div className="flex min-h-screen w-full items-center justify-center bg-surface px-4 py-10">
      <Card elevated className="w-full max-w-100 p-7">
        <div className="mb-6 flex justify-center">
          <BrandMark />
        </div>

        <AnimatePresence mode="wait" initial={false}>
          {step === "credentials" ? (
            <m.form
              key="credentials"
              onSubmit={handleCredentials}
              initial={{ opacity: 0, x: -8 }}
              animate={{ opacity: 1, x: 0 }}
              exit={{ opacity: 0, x: -8 }}
              transition={{ duration: 0.18, ease: "easeOut" }}
              className="flex flex-col gap-4"
            >
              <Input
                id="login-username"
                value={username}
                onChange={(e) => setUsername(e.target.value)}
                placeholder={t("login.username")}
                aria-label={t("login.username")}
                autoComplete="username"
                autoFocus
              />

              <Input
                id="login-password"
                type="password"
                value={password}
                onChange={(e) => setPassword(e.target.value)}
                placeholder={t("login.password")}
                aria-label={t("login.password")}
                autoComplete="current-password"
              />

              <Button type="submit" variant="white" className="w-full">
                {t("login.submit")}
              </Button>
              <p className="text-center text-xs text-ink-tertiary">
                {t("login.credentialsHint")}
              </p>
            </m.form>
          ) : (
            <m.form
              key="twofactor"
              onSubmit={handleVerify}
              initial={{ opacity: 0, x: 8 }}
              animate={{ opacity: 1, x: 0 }}
              exit={{ opacity: 0, x: 8 }}
              transition={{ duration: 0.18, ease: "easeOut" }}
              className="flex flex-col gap-4"
            >
              <Input
                id="login-code"
                value={code}
                onChange={(e) =>
                  setCode(e.target.value.replace(/[^0-9]/g, "").slice(0, 6))
                }
                placeholder={t("login.code")}
                aria-label={t("login.code")}
                inputMode="numeric"
                autoComplete="one-time-code"
                maxLength={6}
                mono
                autoFocus
              />

              <Button
                type="submit"
                variant="white"
                className="w-full"
                disabled={code.length !== 6}
              >
                {t("login.verify")}
              </Button>
            </m.form>
          )}
        </AnimatePresence>
      </Card>
    </div>
  );
}
