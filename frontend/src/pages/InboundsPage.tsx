
import { Suspense, lazy, useCallback, useMemo, useState } from "react";
import { Link, useSearchParams } from "react-router-dom";
import { Plus } from "lucide-react";

import { Button } from "@/components/ui/button";
import { InboundsList } from "@/components/inbounds/inbounds-list";
import type { InboundFormMode } from "@/components/inbounds/inbound-form-modal";
import { useInbounds } from "@/lib/store";
import type { Inbound, Protocol } from "@/lib/store";
import { useI18n } from "@/lib/i18n";

const InboundFormModal = lazy(() =>
  import("@/components/inbounds/inbound-form-modal").then((m) => ({ default: m.InboundFormModal }))
);

const prefetchInboundForm = () => {
  void import("@/components/inbounds/inbound-form-modal");
};

type ModalState = {
  open: boolean;
  mode: InboundFormMode;
  inbound: Inbound | null;
};

const PROTOCOLS: Protocol[] = ["vless", "naive", "hysteria2"];

export function InboundsPage() {
  const [searchParams] = useSearchParams();
  const inbounds = useInbounds();
  const { t } = useI18n();
  const [modal, setModal] = useState<ModalState>({ open: false, mode: "create", inbound: null });
  const protocol = searchParams.get("protocol");
  const protocolFilter = PROTOCOLS.includes(protocol as Protocol) ? (protocol as Protocol) : null;
  const filtered = useMemo(
    () => (protocolFilter ? inbounds.filter((inbound) => inbound.protocol === protocolFilter) : inbounds),
    [inbounds, protocolFilter]
  );

  const openCreate = useCallback(() => setModal({ open: true, mode: "create", inbound: null }), []);
  const openEdit = useCallback((inbound: Inbound) => setModal({ open: true, mode: "edit", inbound }), []);
  const closeModal = useCallback(() => setModal((prev) => ({ ...prev, open: false })), []);
  const openClone = useCallback((inbound: Inbound) => setModal({ open: true, mode: "clone", inbound }), []);

  return (
    <div className="mx-auto flex max-w-[1240px] flex-col gap-6">
      <div className="flex flex-wrap items-center justify-between gap-3">
        <div>
          <h2 className="text-2xl font-semibold text-ink-primary">{t("inbounds.title")}</h2>
          {protocolFilter ? (
            <div className="mt-2 flex items-center gap-2 text-xs text-ink-tertiary">
              <span>{t("inbounds.filterProtocol", { protocol: protocolFilter })}</span>
              <Link to="/inbounds" className="text-ink-secondary transition-colors duration-150 hover:text-ink-primary">
                {t("inbounds.clearFilter")}
              </Link>
            </div>
          ) : null}
        </div>
        <Button
          variant="white"
          onClick={openCreate}
          onMouseEnter={prefetchInboundForm}
          onFocus={prefetchInboundForm}
        >
          <Plus size={16} />
          {t("inbounds.new")}
        </Button>
      </div>

      <InboundsList inbounds={filtered} onEdit={openEdit} />

      {modal.open ? (
        <Suspense fallback={null}>
          <InboundFormModal
            open={modal.open}
            mode={modal.mode}
            inbound={modal.inbound}
            onClose={closeModal}
            onClone={openClone}
          />
        </Suspense>
      ) : null}
    </div>
  );
}
