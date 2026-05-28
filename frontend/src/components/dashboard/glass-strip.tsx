
import { useStore } from "@/lib/mock/store";
import { StatusDot } from "@/components/ui/status-dot";
import { formatBytes, formatUptime } from "@/lib/format";

export function GlassStrip() {
  const { metrics } = useStore();
  const running = metrics.coreRunning;

  return (
    <div className="glass flex flex-col gap-4 rounded-2xl px-5 py-4 sm:flex-row sm:items-center sm:divide-x sm:divide-white/8 sm:gap-0">
      <div className="flex items-center gap-3 sm:pr-6">
        <div className="flex h-9 items-center gap-2 rounded-full border border-subtle bg-canvas/70 px-3 text-xs">
          <StatusDot state={running ? "online" : "stopped"} />
          <span className="text-ink-primary">{running ? "Active" : "Stopped"}</span>
        </div>
        <span className="font-mono text-xs text-ink-tertiary">{metrics.coreVersion}</span>
      </div>

      <div className="flex flex-col gap-0.5 sm:px-6">
        <span className="text-xs text-ink-secondary">Uptime</span>
        <span className="font-mono text-sm text-ink-primary">{formatUptime(metrics.uptimeSeconds)}</span>
      </div>

      <div className="grid grid-cols-2 gap-x-6 gap-y-1 sm:ml-auto sm:pl-6">
        <span className="text-xs text-ink-secondary">Total sent</span>
        <span className="font-mono text-sm text-ink-primary">{formatBytes(metrics.totalSent)}</span>
        <span className="text-xs text-ink-secondary">Total received</span>
        <span className="font-mono text-sm text-ink-primary">{formatBytes(metrics.totalReceived)}</span>
      </div>
    </div>
  );
}
