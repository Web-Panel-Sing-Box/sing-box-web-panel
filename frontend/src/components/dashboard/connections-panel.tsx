
import { useStore } from "@/lib/mock/store";
import { Card, CardHeader, CardLabel } from "@/components/ui/card";
import { AnimatedNumber } from "@/components/ui/animated-number";
import { StatusDot } from "@/components/ui/status-dot";

export function InboundsActiveCard() {
  const { inbounds } = useStore();
  const active = inbounds.filter((i) => i.enabled);
  const protocols = Array.from(new Set(active.map((i) => i.protocol)));
  return (
    <Card>
      <CardHeader>
        <CardLabel>Inbounds active</CardLabel>
        <span className="text-xs text-ink-tertiary">{inbounds.length} total</span>
      </CardHeader>
      <div className="text-5xl font-semibold leading-none tracking-tight text-ink-primary">
        <AnimatedNumber value={active.length} />
      </div>
      <p className="mt-2 text-sm text-ink-secondary">across {protocols.length} protocols</p>
      <div className="mt-4 flex flex-wrap gap-2">
        {protocols.map((p) => (
          <span
            key={p}
            className="rounded-full border border-subtle bg-canvas px-2.5 py-1 font-mono text-[11px] uppercase tracking-wider text-ink-secondary"
          >
            {p}
          </span>
        ))}
      </div>
    </Card>
  );
}

export function ClientsTelemetryCard() {
  const { clients, metrics } = useStore();
  return (
    <Card>
      <CardHeader>
        <CardLabel>Clients telemetry</CardLabel>
        <span className="text-xs text-ink-tertiary">realtime</span>
      </CardHeader>
      <div className="flex items-baseline gap-2">
        <span className="text-xs text-ink-secondary">Total users</span>
      </div>
      <div className="text-5xl font-semibold leading-none tracking-tight text-ink-primary">
        <AnimatedNumber value={clients.length} />
      </div>
      <div className="mt-4 flex items-center gap-2 text-sm text-ink-secondary">
        <StatusDot state="online" />
        <span>Online now:</span>
        <span className="font-mono text-ink-primary">
          <AnimatedNumber value={metrics.onlineNow} />
        </span>
      </div>
    </Card>
  );
}
