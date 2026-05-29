import { lazy } from "react";
import { LazyMotion, domMax } from "framer-motion";
import { BrowserRouter, Route, Routes } from "react-router-dom";

import { Toaster } from "@/components/ui/toaster";
import { I18nProvider } from "@/lib/i18n";
import { PanelLayout } from "@/pages/PanelLayout";

const DashboardPage = lazy(() =>
  import("@/pages/DashboardPage").then((m) => ({ default: m.DashboardPage }))
);
const InboundsPage = lazy(() =>
  import("@/pages/InboundsPage").then((m) => ({ default: m.InboundsPage }))
);
const ClientsPage = lazy(() =>
  import("@/pages/ClientsPage").then((m) => ({ default: m.ClientsPage }))
);
const SettingsPage = lazy(() =>
  import("@/pages/SettingsPage").then((m) => ({ default: m.SettingsPage }))
);
const LogsPage = lazy(() =>
  import("@/pages/LogsPage").then((m) => ({ default: m.LogsPage }))
);

export function App() {
  return (
    <I18nProvider>
      <LazyMotion features={domMax} strict>
        <Toaster>
          <BrowserRouter>
            <Routes>
              <Route element={<PanelLayout />}>
                <Route index element={<DashboardPage />} />
                <Route path="inbounds" element={<InboundsPage />} />
                <Route path="clients" element={<ClientsPage />} />
                <Route path="settings" element={<SettingsPage />} />
                <Route path="logs" element={<LogsPage />} />
              </Route>
            </Routes>
          </BrowserRouter>
        </Toaster>
      </LazyMotion>
    </I18nProvider>
  );
}
