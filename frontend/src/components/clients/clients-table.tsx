
import { useMemo, useState } from "react";

import { Card } from "@/components/ui/card";
import { Progress } from "@/components/ui/progress";
import { StatusDot } from "@/components/ui/status-dot";
import { formatBytes, formatDate, truncateMiddle } from "@/lib/format";
import { useStore } from "@/lib/mock/store";
import type { Client } from "@/lib/mock/clients";
import { cn } from "@/lib/utils";

import type { FilterState } from "./client-filter-bar";

type Props = {
  filter: FilterState;
  onSelect: (client: Client) => void;
};

export function ClientsTable({ filter, onSelect }: Props) {
  const { clients, inbounds } = useStore();
  const inboundMap = useMemo(() => new Map(inbounds.map((i) => [i.id, i])), [inbounds]);

  const rows = useMemo(() => {
    const q = filter.query.trim().toLowerCase();
    return clients.filter((c) => {
      if (filter.inboundId !== "all" && c.inboundId !== filter.inboundId) return false;
      if (filter.status !== "all" && c.status !== filter.status) return false;
      if (!q) return true;
      return c.name.toLowerCase().includes(q) || c.uuid.toLowerCase().includes(q);
    });
  }, [clients, filter]);

  return (
    <Card padded={false} className="overflow-hidden">
      <div className="grid grid-cols-[1fr_220px_140px_180px_120px_60px] items-center gap-3 border-b border-subtle px-5 py-3 text-[11px] uppercase tracking-wider text-ink-tertiary">
        <span>Name</span>
        <span>Data usage</span>
        <span>Expiry</span>
        <span>Inbound</span>
        <span>Status</span>
        <span />
      </div>
      <div className="divide-y divide-subtle">
        {rows.length === 0 ? (
          <div className="px-5 py-10 text-center text-sm text-ink-tertiary">No clients match the current filter.</div>
        ) : null}
        {rows.map((c) => {
          const inbound = inboundMap.get(c.inboundId);
          const total = c.usedDown + c.usedUp;
          const pct = c.totalQuota > 0 ? (total / c.totalQuota) * 100 : 0;
          return (
            <Row
              key={c.id}
              client={c}
              inboundLabel={inbound?.remark ?? "—"}
              pct={pct}
              total={total}
              onClick={() => onSelect(c)}
            />
          );
        })}
      </div>
    </Card>
  );
}

function Row({
  client,
  inboundLabel,
  pct,
  total,
  onClick
}: {
  client: Client;
  inboundLabel: string;
  pct: number;
  total: number;
  onClick: () => void;
}) {
  const [hover, setHover] = useState(false);
  return (
    <button
      type="button"
      onClick={onClick}
      onMouseEnter={() => setHover(true)}
      onMouseLeave={() => setHover(false)}
      className={cn(
        "grid w-full grid-cols-[1fr_220px_140px_180px_120px_60px] items-center gap-3 px-5 py-3 text-left transition-colors duration-200",
        hover && "bg-elevated"
      )}
    >
      <div className="min-w-0">
        <div className="truncate text-sm text-ink-primary">{client.name}</div>
        <div className="truncate font-mono text-[11px] text-ink-tertiary">{truncateMiddle(client.uuid, 8, 6)}</div>
      </div>
      <div className="min-w-0">
        <div className="mb-1 flex justify-between font-mono text-[11px] text-ink-tertiary">
          <span>{formatBytes(total)}</span>
          <span>{formatBytes(client.totalQuota)}</span>
        </div>
        <Progress value={pct} />
      </div>
      <span className="text-xs text-ink-secondary">{formatDate(client.expiry)}</span>
      <span className="truncate text-xs text-ink-secondary">{inboundLabel}</span>
      <StatusBadge status={client.status} />
      <span className="text-right text-ink-tertiary">›</span>
    </button>
  );
}

function StatusBadge({ status }: { status: Client["status"] }) {
  const map: Record<Client["status"], { label: string; dot: "online" | "stopped" | "neutral"; cls: string }> = {
    active: { label: "Active", dot: "online", cls: "text-success" },
    disabled: { label: "Disabled", dot: "neutral", cls: "text-ink-tertiary" },
    expired: { label: "Expired", dot: "stopped", cls: "text-danger" }
  };
  const v = map[status];
  return (
    <span className="inline-flex items-center gap-2 rounded-full border border-subtle bg-canvas px-2.5 py-1 text-[11px]">
      <StatusDot state={v.dot} />
      <span className={v.cls}>{v.label}</span>
    </span>
  );
}
