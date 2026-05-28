"use client";

import { useEffect, useState } from "react";

import { DashboardShell } from "@/components/dashboard-shell";
import { LoginPanel } from "@/components/login-panel";
import { api } from "@/lib/api";

type AuthState = "checking" | "authenticated" | "anonymous";

export default function Home() {
  const [authState, setAuthState] = useState<AuthState>("checking");

  useEffect(() => {
    api
      .me()
      .then(() => setAuthState("authenticated"))
      .catch(() => setAuthState("anonymous"));
  }, []);

  if (authState === "checking") {
    return (
      <main className="grid min-h-screen place-items-center bg-void text-xs text-zinc-500">
        BOOTSTRAP...
      </main>
    );
  }

  if (authState === "anonymous") {
    return <LoginPanel onAuthenticated={() => setAuthState("authenticated")} />;
  }

  return <DashboardShell authState={authState} onAuthState={setAuthState} />;
}
