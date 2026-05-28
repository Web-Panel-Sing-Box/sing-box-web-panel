
import { Card } from "@/components/ui/card";
import { useStore } from "@/lib/mock/store";

import { InboundRow } from "./inbound-row";

export function InboundsList() {
  const { inbounds } = useStore();
  return (
    <Card padded={false} className="overflow-hidden">
      <div className="grid grid-cols-[110px_110px_1fr_90px_auto_auto_24px] items-center gap-3 border-b border-subtle px-4 py-3 text-[11px] uppercase tracking-wider text-ink-tertiary sm:px-5">
        <span>Protocol</span>
        <span>Port</span>
        <span>Remark</span>
        <span className="hidden md:inline">Clients</span>
        <span>Enabled</span>
        <span>Status</span>
        <span />
      </div>
      <div>
        {inbounds.map((inbound) => (
          <InboundRow key={inbound.id} inbound={inbound} />
        ))}
      </div>
    </Card>
  );
}
