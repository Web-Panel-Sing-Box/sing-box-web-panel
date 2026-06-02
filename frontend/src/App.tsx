import { lazy, Suspense } from "react";
import { LazyMotion, MotionConfig, domMax } from "framer-motion";
import { BrowserRouter, Navigate, Route, Routes } from "react-router-dom";

import { RequireAuth } from "@/components/auth/require-auth";
import { Toaster } from "@/components/ui/toaster";
import { AuthProvider } from "@/lib/auth";
import { I18nProvider } from "@/lib/i18n";
import { PanelLayout } from "@/pages/PanelLayout";

const LoginPage = lazy(() =>
  import("@/pages/LoginPage").then((m) => ({ default: m.LoginPage })),
);
const DashboardPage = lazy(() =>
  import("@/pages/DashboardPage").then((m) => ({ default: m.DashboardPage })),
);
const InboundsPage = lazy(() =>
  import("@/pages/InboundsPage").then((m) => ({ default: m.InboundsPage })),
);
const ClientsPage = lazy(() =>
  import("@/pages/ClientsPage").then((m) => ({ default: m.ClientsPage })),
);
const SettingsPage = lazy(() =>
  import("@/pages/SettingsPage").then((m) => ({ default: m.SettingsPage })),
);
const LogsPage = lazy(() =>
  import("@/pages/LogsPage").then((m) => ({ default: m.LogsPage })),
);

export function App() {
  return (
    <I18nProvider>
      <LazyMotion features={domMax} strict>
        <MotionConfig reducedMotion="user">
        <Toaster>
          <AuthProvider>
            <BrowserRouter>
              <Suspense
                fallback={<div className="min-h-screen w-full bg-surface" />}
              >
                <Routes>
                  <Route path="/" element={<Navigate to="/dashboard" replace />} />
                  <Route path="/login" element={<LoginPage />} />
                  <Route element={<RequireAuth />}>
                    <Route element={<PanelLayout />}>
                      <Route index element={<Navigate to="/dashboard" replace />} />
                      <Route path="dashboard" element={<DashboardPage />} />
                      <Route path="inbounds" element={<InboundsPage />} />
                      <Route path="clients" element={<ClientsPage />} />
                      <Route path="settings" element={<SettingsPage />} />
                      <Route path="logs" element={<LogsPage />} />
                    </Route>
                  </Route>
                </Routes>
              </Suspense>
            </BrowserRouter>
          </AuthProvider>
        </Toaster>
        </MotionConfig>
      </LazyMotion>
    </I18nProvider>
  );
}
