
import { Card } from "@/components/ui/card";
import type { Inbound } from "@/lib/mock/inbounds";
import { useI18n } from "@/lib/i18n";

import { InboundHeader, InboundRow } from "./inbound-row";

export function InboundsList({ inbounds, onEdit }: { inbounds: Inbound[]; onEdit: (inbound: Inbound) => void }) {
  const { t } = useI18n();
  return (
    <Card padded={false} className="overflow-x-auto">
      <div className="min-w-[980px]">
        <InboundHeader>
          <span>{t("common.protocol")}</span>
          <span>{t("common.port")}</span>
          <span>{t("common.remark")}</span>
          <span>{t("common.transport")}</span>
          <span>{t("common.clients")}</span>
          <span className="text-center">{t("common.enabled")}</span>
          <span>{t("common.status")}</span>
        </InboundHeader>
        {inbounds.map((inbound) => (
          <InboundRow key={inbound.id} inbound={inbound} onEdit={onEdit} />
        ))}
        {inbounds.length === 0 ? (
          <div className="px-5 py-10 text-center text-sm text-ink-tertiary">{t("inbounds.noMatch")}</div>
        ) : null}
      </div>
    </Card>
  );
}
