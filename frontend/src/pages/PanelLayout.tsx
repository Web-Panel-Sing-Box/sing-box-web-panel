import { useState } from "react";
import { Outlet } from "react-router-dom";

import { Sidebar } from "@/components/shell/sidebar";
import { TopBar } from "@/components/shell/topbar";
import { PageTransition } from "@/components/shell/page-transition";
import { MockStoreProvider } from "@/lib/mock/store";

export function PanelLayout() {
  const [mobileOpen, setMobileOpen] = useState(false);
  return (
    <MockStoreProvider>
      <div className="flex min-h-screen w-full bg-surface">
        <Sidebar mobileOpen={mobileOpen} onCloseMobile={() => setMobileOpen(false)} />
        <div className="flex min-w-0 flex-1 flex-col">
          <TopBar onOpenMobile={() => setMobileOpen(true)} />
          <main className="flex-1 px-4 py-6 sm:px-6 lg:px-8">
            <PageTransition>
              <Outlet />
            </PageTransition>
          </main>
        </div>
      </div>
    </MockStoreProvider>
  );
}
