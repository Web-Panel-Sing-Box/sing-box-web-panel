
import type { MouseEventHandler } from "react";
import { Link, useNavigate } from "react-router-dom";
import { Network, Users } from "lucide-react";

import { useStore } from "@/lib/mock/store";
import { Card, CardHeader, CardLabel } from "@/components/ui/card";
import { AnimatedNumber } from "@/components/ui/animated-number";
import { StatusDot } from "@/components/ui/status-dot";
import type { Inbound, Protocol } from "@/lib/mock/inbounds";
import { cn } from "@/lib/utils";
import { useI18n } from "@/lib/i18n";

export function InboundsActiveCard() {
  const { inbounds } = useStore();
  const navigate = useNavigate();
  const { t } = useI18n();
  const active = inbounds.filter((i) => i.enabled);
  const protocols = Array.from(new Set(active.map((i) => i.protocol)));
  return (
    <Card
      role="link"
      tabIndex={0}
      onClick={() => navigate("/inbounds")}
      onKeyDown={(event) => {
        if (event.key === "Enter" || event.key === " ") navigate("/inbounds");
      }}
      className="group cursor-pointer transition-colors duration-200 hover:bg-hover"
    >
      <CardHeader>
        <CardLabel className="inline-flex items-center gap-2">
          <Network size={15} />
          {t("dashboard.inboundsActive")}
        </CardLabel>
        <span className="text-xs text-ink-tertiary">{t("dashboard.total", { count: inbounds.length })}</span>
      </CardHeader>
      <div className="text-5xl font-semibold leading-none tracking-tight text-ink-primary">
        <AnimatedNumber value={active.length} />
      </div>
      <p className="mt-2 text-sm text-ink-secondary">{t("dashboard.acrossProtocols", { count: protocols.length })}</p>
      <div className="mt-4 flex flex-wrap gap-2">
        {protocols.map((p) => (
          <ProtocolLink
            key={p}
            protocol={p}
            onClick={(event) => event.stopPropagation()}
          />
        ))}
      </div>
    </Card>
  );
}

export function ClientsTelemetryCard() {
  const { clients, metrics } = useStore();
  const navigate = useNavigate();
  const { t } = useI18n();
  return (
    <Card
      role="link"
      tabIndex={0}
      onClick={() => navigate("/clients")}
      onKeyDown={(event) => {
        if (event.key === "Enter" || event.key === " ") navigate("/clients");
      }}
      className="group cursor-pointer transition-colors duration-200 hover:bg-hover"
    >
      <CardHeader>
        <CardLabel className="inline-flex items-center gap-2">
          <Users size={15} />
          {t("dashboard.clientsTelemetry")}
        </CardLabel>
        <span className="text-xs text-ink-tertiary">{t("dashboard.realtime")}</span>
      </CardHeader>
      <div className="flex items-baseline gap-2">
        <span className="text-xs text-ink-secondary">{t("dashboard.totalUsers")}</span>
      </div>
      <div className="text-5xl font-semibold leading-none tracking-tight text-ink-primary">
        <AnimatedNumber value={clients.length} />
      </div>
      <div className="mt-4 flex items-center gap-2 text-sm text-ink-secondary">
        <StatusDot state="online" />
        <span>{t("dashboard.onlineNow")}</span>
        <span className="font-mono text-ink-primary">
          <AnimatedNumber value={metrics.onlineNow} />
        </span>
      </div>
    </Card>
  );
}

function ProtocolLink({
  protocol,
  onClick
}: {
  protocol: Inbound["protocol"];
  onClick: MouseEventHandler<HTMLAnchorElement>;
}) {
  const palette: Record<Protocol, string> = {
    vless: "hover:border-cyan/50 hover:text-cyan",
    naive: "hover:border-amber/50 hover:text-amber",
    hysteria2: "hover:border-violet/50 hover:text-violet"
  };
  return (
    <Link
      to={`/inbounds?protocol=${protocol}`}
      onClick={onClick}
      className={cn(
        "rounded-full border border-subtle bg-canvas px-2.5 py-1 font-mono text-[11px] uppercase tracking-wider text-ink-secondary transition-colors duration-200",
        palette[protocol]
      )}
    >
      {protocol}
    </Link>
  );
}
