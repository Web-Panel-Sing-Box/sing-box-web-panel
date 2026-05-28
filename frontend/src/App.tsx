import { BrowserRouter, Route, Routes } from "react-router-dom";

import { Toaster } from "@/components/ui/toaster";
import { ClientsPage } from "@/pages/ClientsPage";
import { DashboardPage } from "@/pages/DashboardPage";
import { InboundsPage } from "@/pages/InboundsPage";
import { LogsPage } from "@/pages/LogsPage";
import { PanelLayout } from "@/pages/PanelLayout";
import { SettingsPage } from "@/pages/SettingsPage";

export function App() {
  return (
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
  );
}
