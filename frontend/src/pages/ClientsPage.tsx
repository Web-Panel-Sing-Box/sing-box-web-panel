import { useEffect, useState } from "react";
import { useSearchParams } from "react-router-dom";
import { Plus } from "lucide-react";

import { Button } from "@/components/ui/button";
import { AddClientModal } from "@/components/clients/add-client-modal";
import { ClientFilterBar, type FilterState } from "@/components/clients/client-filter-bar";
import { ClientsTable } from "@/components/clients/clients-table";
import { ClientDetailModal } from "@/components/clients/client-detail-modal";
import type { Client } from "@/lib/mock/clients";
import { useI18n } from "@/lib/i18n";

export function ClientsPage() {
  const [searchParams, setSearchParams] = useSearchParams();
  const [filter, setFilter] = useState<FilterState>({ query: "", inboundId: "all", status: "all" });
  const [selected, setSelected] = useState<Client | null>(null);
  const [addOpen, setAddOpen] = useState(false);
  const { t } = useI18n();

  useEffect(() => {
    const inboundId = searchParams.get("inbound") ?? "all";
    setFilter((prev) => (prev.inboundId === inboundId ? prev : { ...prev, inboundId }));
  }, [searchParams]);

  function updateFilter(next: FilterState) {
    setFilter(next);
    const params = new URLSearchParams(searchParams);
    if (next.inboundId === "all") params.delete("inbound");
    else params.set("inbound", next.inboundId);
    setSearchParams(params, { replace: true });
  }

  return (
    <div className="mx-auto flex max-w-[1320px] flex-col gap-6">
      <div className="flex flex-wrap items-center justify-between gap-3">
        <h2 className="text-2xl font-semibold text-ink-primary">{t("clients.title")}</h2>
        <Button variant="white" onClick={() => setAddOpen(true)}>
          <Plus size={16} />
          {t("clients.add")}
        </Button>
      </div>

      <ClientFilterBar value={filter} onChange={updateFilter} />
      <ClientsTable filter={filter} onSelect={setSelected} />

      <ClientDetailModal client={selected} onClose={() => setSelected(null)} />
      <AddClientModal
        open={addOpen}
        onClose={() => setAddOpen(false)}
        defaultInboundId={filter.inboundId !== "all" ? filter.inboundId : undefined}
      />
    </div>
  );
}
