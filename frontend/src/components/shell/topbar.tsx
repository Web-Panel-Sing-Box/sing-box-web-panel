import { useLocation } from "react-router-dom";
import { Menu } from "lucide-react";

import { useMetrics, useStoreActions } from "@/lib/store";
import { StatusDot } from "@/components/ui/status-dot";
import { cn } from "@/lib/utils";

const TITLES: Record<string, string> = {
  "/dashboard": "Dashboard",
  "/inbounds": "Inbounds",
  "/clients": "Clients",
  "/settings": "Settings",
  "/logs": "Logs",
};

export function TopBar({ onOpenMobile }: { onOpenMobile: () => void }) {
  const pathname = useLocation().pathname;
  const { metrics } = useMetrics();
  const { setCoreRunning } = useStoreActions();
  const title = TITLES[pathname] ?? "Shilka";
  const running = metrics.coreRunning;

  return (
    <header className="sticky top-0 z-20 flex h-14 items-center justify-between gap-4 border-b border-subtle bg-surface/80 px-4 backdrop-blur-md sm:px-6">
      <div className="flex items-center gap-3">
        <button
          onClick={onOpenMobile}
          className="lg:hidden text-muted hover:text-brand"
          aria-label="Open sidebar"
        >
          <Menu size={20} />
        </button>
        <h1 className="text-sm font-semibold">{title}</h1>
      </div>

      <div className="flex items-center gap-2">
        <StatusDot state={running ? "online" : "stopped"} />
        <span className="text-xs text-muted">
          {running ? "Core running" : "Core stopped"}
        </span>
      </div>
    </header>
  );
}
