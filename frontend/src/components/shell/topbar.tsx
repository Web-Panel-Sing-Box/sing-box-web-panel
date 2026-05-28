
import { useLocation } from "react-router-dom";
import { Menu } from "lucide-react";

import { useStore } from "@/lib/mock/store";
import { StatusDot } from "@/components/ui/status-dot";
import { cn } from "@/lib/utils";

const TITLES: Record<string, string> = {
  "/": "Dashboard",
  "/inbounds": "Inbounds",
  "/clients": "Clients",
  "/settings": "Settings",
  "/logs": "Logs"
};

export function TopBar({ onOpenMobile }: { onOpenMobile: () => void }) {
  const pathname = useLocation().pathname;
  const { metrics, setCoreRunning } = useStore();
  const title = TITLES[pathname] ?? "Sing Grok";
  const running = metrics.coreRunning;

  return (
    <header className="sticky top-0 z-20 flex h-14 items-center justify-between gap-4 border-b border-subtle bg-surface/80 px-4 backdrop-blur-md sm:px-6">
      <div className="flex items-center gap-3">
        <button
          type="button"
          onClick={onOpenMobile}
          className="-ml-1 grid size-9 place-items-center rounded-lg text-ink-secondary transition-colors duration-200 hover:bg-hover hover:text-ink-primary lg:hidden"
          aria-label="Open menu"
        >
          <Menu size={18} />
        </button>
        <h1 className="text-base font-semibold text-ink-primary sm:text-lg">{title}</h1>
      </div>

      <button
        type="button"
        onClick={() => setCoreRunning(!running)}
        className={cn(
          "flex h-9 items-center gap-2 rounded-full border border-subtle bg-canvas/80 px-3 text-xs text-ink-primary transition-colors duration-200 hover:bg-hover"
        )}
        title={running ? "Click to stop (mock)" : "Click to start (mock)"}
      >
        <StatusDot state={running ? "online" : "stopped"} />
        <span>
          Sing-Box: <span className={running ? "text-success" : "text-danger"}>{running ? "Active" : "Stopped"}</span>
        </span>
      </button>
    </header>
  );
}
