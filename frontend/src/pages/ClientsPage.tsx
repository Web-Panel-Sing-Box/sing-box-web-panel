import { Suspense, lazy, useCallback, useEffect, useState } from "react";
import { useSearchParams } from "react-router-dom";
import { Plus } from "lucide-react";

import { Button } from "@/components/ui/button";
import {
  ClientFilterBar,
  type FilterState,
} from "@/components/clients/client-filter-bar";
import { ClientsTable } from "@/components/clients/clients-table";
import { useDisclosure } from "@/hooks/useDisclosure";
import type { Client } from "@/lib/store";
import { useI18n } from "@/lib/i18n";

const AddClientModal = lazy(() =>
  import("@/components/clients/add-client-modal").then((m) => ({
    default: m.AddClientModal,
  })),
);
const ClientDetailModal = lazy(() =>
  import("@/components/clients/client-detail-modal").then((m) => ({
    default: m.ClientDetailModal,
  })),
);

const prefetchAddClient = () => {
  void import("@/components/clients/add-client-modal");
};
const prefetchDetailModal = () => {
  void import("@/components/clients/client-detail-modal");
};

export function ClientsPage() {
  const [searchParams, setSearchParams] = useSearchParams();
  const [filter, setFilter] = useState<FilterState>({
    query: "",
    inboundId: "all",
    nodeId: "all",
    status: "all",
  });
  const [selected, setSelected] = useState<Client | null>(null);
  const addModal = useDisclosure(false);
  const { t } = useI18n();

  useEffect(() => {
    const inboundId = searchParams.get("inbound") ?? "all";
    const nodeId = searchParams.get("node") ?? "all";
    setFilter((prev) =>
      prev.inboundId === inboundId && prev.nodeId === nodeId
        ? prev
        : { ...prev, inboundId, nodeId },
    );
  }, [searchParams]);

  const updateFilter = useCallback(
    (next: FilterState) => {
      setFilter(next);
      const params = new URLSearchParams(searchParams);
      if (next.inboundId === "all") params.delete("inbound");
      else params.set("inbound", next.inboundId);
      if (next.nodeId === "all") params.delete("node");
      else params.set("node", next.nodeId);
      setSearchParams(params, { replace: true });
    },
    [searchParams, setSearchParams],
  );

  const closeDetail = useCallback(() => setSelected(null), []);

  return (
    <div
      className="mx-auto flex max-w-[1320px] flex-col gap-6"
      onMouseEnter={prefetchDetailModal}
    >
      <div className="flex flex-wrap items-center justify-between gap-3">
        <h2 className="text-2xl font-semibold text-ink-primary">
          {t("clients.title")}
        </h2>
        <Button
          variant="white"
          onClick={addModal.open}
          onMouseEnter={prefetchAddClient}
          onFocus={prefetchAddClient}
        >
          <Plus size={16} />
          {t("clients.add")}
        </Button>
      </div>

      <ClientFilterBar value={filter} onChange={updateFilter} />
      <ClientsTable filter={filter} onSelect={setSelected} />

      {selected ? (
        <Suspense fallback={null}>
          <ClientDetailModal client={selected} onClose={closeDetail} />
        </Suspense>
      ) : null}
      {addModal.isOpen ? (
        <Suspense fallback={null}>
          <AddClientModal
            open={addModal.isOpen}
            onClose={addModal.close}
            defaultInboundId={
              filter.inboundId !== "all" ? filter.inboundId : undefined
            }
            defaultNodeId={filter.nodeId !== "all" ? filter.nodeId : undefined}
          />
        </Suspense>
      ) : null}
    </div>
  );
}
