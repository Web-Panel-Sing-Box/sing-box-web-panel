import { memo, useCallback, type ReactNode } from "react";
import { Link } from "react-router-dom";

import { Toggle } from "@/components/ui/toggle";
import { StatusDot } from "@/components/ui/status-dot";
import { cn } from "@/lib/utils";
import type { Inbound, Network, Protocol, TlsMode, VlessTransport } from "@/lib/mock/inbounds";
import { useStoreActions } from "@/lib/mock/store";
import { useToast } from "@/components/ui/toast";
import { useI18n } from "@/lib/i18n";

const ROW_GRID =
  "grid-cols-[minmax(96px,1fr)_minmax(90px,0.8fr)_minmax(180px,1.5fr)_minmax(150px,1.2fr)_minmax(100px,0.8fr)_minmax(90px,0.7fr)_minmax(110px,0.8fr)]";

function InboundRowImpl({ inbound, onEdit }: { inbound: Inbound; onEdit: (inbound: Inbound) => void }) {
  const { toggleInbound } = useStoreActions();
  const { push } = useToast();
  const { t } = useI18n();

  const open = useCallback(() => {
    onEdit(inbound);
  }, [onEdit, inbound]);

  const onToggle = useCallback(
    (v: boolean) => {
      toggleInbound(inbound.id);
      push(
        t("inbounds.toggled", { remark: inbound.remark, state: v ? t("common.enabled") : t("common.disabled") }),
        "success"
      );
    },
    [toggleInbound, inbound.id, inbound.remark, push, t]
  );

  return (
    <div
      role="button"
      tabIndex={0}
      onClick={open}
      onKeyDown={(event) => {
        if (event.key === "Enter" || event.key === " ") open();
      }}
      className={cn(
        "grid w-full cursor-pointer items-center gap-3 border-b border-subtle px-4 py-3 text-left transition-colors duration-200 last:border-b-0 hover:bg-hover sm:px-5",
        ROW_GRID
      )}
    >
      <ProtocolChip protocol={inbound.protocol} />
      <span className="font-mono text-sm text-ink-secondary">{inbound.port}</span>
      <span className="truncate text-sm text-ink-primary">{inbound.remark}</span>
      <TransportChip inbound={inbound} />
      <Link
        to={`/clients?inbound=${inbound.id}`}
        onClick={(event) => event.stopPropagation()}
        className="w-fit rounded-full border border-subtle bg-canvas px-2.5 py-1 text-xs text-ink-secondary transition-colors duration-200 hover:border-white/20 hover:text-ink-primary"
      >
        {t("inbounds.clientCount", { count: inbound.clientCount })}
      </Link>
      <div onClick={(event) => event.stopPropagation()} className="flex justify-center">
        <Toggle checked={inbound.enabled} onChange={onToggle} />
      </div>
      <span
        className={cn(
          "inline-flex w-fit items-center gap-2 rounded-full border border-subtle bg-canvas px-2.5 py-1 text-[11px]",
          inbound.enabled ? "text-success" : "text-ink-tertiary"
        )}
      >
        <StatusDot state={inbound.enabled ? "online" : "neutral"} />
        {inbound.enabled ? t("common.active") : t("common.disabled")}
      </span>
    </div>
  );
}

export function InboundHeader({ children }: { children: ReactNode }) {
  return (
    <div className={cn("grid items-center gap-3 border-b border-subtle px-4 py-3 text-[11px] uppercase tracking-wider text-ink-tertiary sm:px-5", ROW_GRID)}>
      {children}
    </div>
  );
}

export const InboundRow = memo(InboundRowImpl);

export function ProtocolChip({ protocol }: { protocol: Inbound["protocol"] }) {
  const palette: Record<Protocol, string> = {
    vless: "text-cyan border-cyan/30",
    naive: "text-amber border-amber/30",
    hysteria2: "text-violet border-violet/30"
  };
  return (
    <span
      className={cn(
        "inline-flex h-7 w-fit items-center rounded-full border bg-canvas px-2.5 font-mono text-[11px] uppercase tracking-wider",
        palette[protocol]
      )}
    >
      {protocol}
    </span>
  );
}

function TransportChip({ inbound }: { inbound: Inbound }) {
  return (
    <span className="inline-flex h-7 w-fit items-center rounded-full border border-white/15 bg-canvas px-2.5 font-mono text-[11px] uppercase tracking-wider text-ink-secondary">
      {connectionLabel(inbound)} · {tlsLabel(inbound.tls)}
    </span>
  );
}

function connectionLabel(inbound: Inbound) {
  if (inbound.protocol === "vless") return transportLabel(inbound.transport ?? "tcp");
  if (inbound.protocol === "naive") return networkLabel(inbound.network ?? "both");
  return "QUIC"; // hysteria2
}

function transportLabel(transport: VlessTransport) {
  const labels: Record<VlessTransport, string> = {
    tcp: "TCP",
    ws: "WS",
    grpc: "gRPC",
    http: "HTTP/2",
    httpupgrade: "HTTPUpgrade"
  };
  return labels[transport];
}

function networkLabel(network: Network) {
  const labels: Record<Network, string> = {
    tcp: "TCP",
    udp: "UDP",
    both: "TCP+UDP"
  };
  return labels[network];
}

function tlsLabel(tls: TlsMode) {
  const labels: Record<TlsMode, string> = {
    none: "None",
    tls: "TLS",
    reality: "Reality"
  };
  return labels[tls];
}
