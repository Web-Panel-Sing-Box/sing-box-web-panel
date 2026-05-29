import { Suspense, useState } from "react";
import { Outlet } from "react-router-dom";
import { Menu } from "lucide-react";

import { Sidebar } from "@/components/shell/sidebar";
import { PageTransition } from "@/components/shell/page-transition";
import { MockStoreProvider } from "@/lib/mock/store";
import { useI18n } from "@/lib/i18n";

function RouteFallback() {
  return (
    <div className="mx-auto flex max-w-[1320px] animate-pulse flex-col gap-6">
      <div className="h-7 w-40 rounded-md bg-elevated/60" />
      <div className="h-[420px] rounded-2xl border border-subtle bg-canvas/40" />
    </div>
  );
}

export function PanelLayout() {
  const [mobileOpen, setMobileOpen] = useState(false);
  const { t } = useI18n();
  return (
    <MockStoreProvider>
      <div className="flex min-h-screen w-full bg-surface">
        <Sidebar mobileOpen={mobileOpen} onCloseMobile={() => setMobileOpen(false)} />
        <div className="flex min-w-0 flex-1 flex-col">
          <button
            type="button"
            onClick={() => setMobileOpen(true)}
            className="fixed left-3 top-3 z-30 grid size-10 place-items-center rounded-lg border border-subtle bg-canvas/90 text-ink-secondary shadow-pop backdrop-blur transition-colors duration-200 hover:bg-hover hover:text-ink-primary lg:hidden"
            aria-label={t("mobile.openMenu")}
          >
            <Menu size={18} />
          </button>
          <main className="flex-1 px-4 pb-6 pt-16 sm:px-6 lg:px-8 lg:py-8">
            <PageTransition>
              <Suspense fallback={<RouteFallback />}>
                <Outlet />
              </Suspense>
            </PageTransition>
          </main>
        </div>
      </div>
    </MockStoreProvider>
  );
}
