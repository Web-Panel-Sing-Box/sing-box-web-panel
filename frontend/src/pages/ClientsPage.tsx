
import { useState } from "react";
import { Plus } from "lucide-react";

import { Button } from "@/components/ui/button";
import { ClientFilterBar, type FilterState } from "@/components/clients/client-filter-bar";
import { ClientsTable } from "@/components/clients/clients-table";
import { ClientDetailModal } from "@/components/clients/client-detail-modal";
import { useToast } from "@/components/ui/toast";
import type { Client } from "@/lib/mock/clients";

export function ClientsPage() {
  const [filter, setFilter] = useState<FilterState>({ query: "", inboundId: "all", status: "all" });
  const [selected, setSelected] = useState<Client | null>(null);
  const { push } = useToast();

  return (
    <div className="mx-auto flex max-w-[1320px] flex-col gap-6">
      <div className="flex flex-wrap items-center justify-between gap-3">
        <div>
          <h2 className="text-2xl font-semibold text-ink-primary">Clients</h2>
          <p className="mt-1 text-sm text-ink-tertiary">Manage user quotas, expiry, and subscription links</p>
        </div>
        <Button variant="secondary" onClick={() => push("Use 'New configuration' on Inbounds to provision the first client")}>
          <Plus size={16} />
          Add client
        </Button>
      </div>

      <ClientFilterBar value={filter} onChange={setFilter} />
      <ClientsTable filter={filter} onSelect={setSelected} />

      <ClientDetailModal client={selected} onClose={() => setSelected(null)} />
    </div>
  );
}
