
import { useMemo } from "react";
import { useState } from "react";
import { Link, useSearchParams } from "react-router-dom";
import { Plus } from "lucide-react";

import { Button } from "@/components/ui/button";
import { InboundsList } from "@/components/inbounds/inbounds-list";
import { InboundFormModal, type InboundFormMode } from "@/components/inbounds/inbound-form-modal";
import { useStore } from "@/lib/mock/store";
import type { Inbound, Protocol } from "@/lib/mock/inbounds";
import { useI18n } from "@/lib/i18n";

type ModalState = {
  open: boolean;
  mode: InboundFormMode;
  inbound: Inbound | null;
};

const PROTOCOLS: Protocol[] = ["vless", "naive", "hysteria2"];

export function InboundsPage() {
  const [searchParams] = useSearchParams();
  const { inbounds } = useStore();
  const { t } = useI18n();
  const [modal, setModal] = useState<ModalState>({ open: false, mode: "create", inbound: null });
  const protocol = searchParams.get("protocol");
  const protocolFilter = PROTOCOLS.includes(protocol as Protocol) ? (protocol as Protocol) : null;
  const filtered = useMemo(
    () => (protocolFilter ? inbounds.filter((inbound) => inbound.protocol === protocolFilter) : inbounds),
    [inbounds, protocolFilter]
  );

  return (
    <div className="mx-auto flex max-w-[1240px] flex-col gap-6">
      <div className="flex flex-wrap items-center justify-between gap-3">
        <div>
          <h2 className="text-2xl font-semibold text-ink-primary">{t("inbounds.title")}</h2>
          <p className="mt-1 text-sm text-ink-tertiary">{t("inbounds.description")}</p>
          {protocolFilter ? (
            <div className="mt-2 flex items-center gap-2 text-xs text-ink-tertiary">
              <span>{t("inbounds.filterProtocol", { protocol: protocolFilter })}</span>
              <Link to="/inbounds" className="text-ink-secondary transition-colors duration-150 hover:text-ink-primary">
                {t("inbounds.clearFilter")}
              </Link>
            </div>
          ) : null}
        </div>
        <Button variant="white" onClick={() => setModal({ open: true, mode: "create", inbound: null })} className="rounded-full">
          <Plus size={16} />
          {t("inbounds.new")}
        </Button>
      </div>

      <InboundsList inbounds={filtered} onEdit={(inbound) => setModal({ open: true, mode: "edit", inbound })} />

      <InboundFormModal
        open={modal.open}
        mode={modal.mode}
        inbound={modal.inbound}
        onClose={() => setModal((prev) => ({ ...prev, open: false }))}
        onClone={(inbound) => setModal({ open: true, mode: "clone", inbound })}
      />
    </div>
  );
}
