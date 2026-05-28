
import { Search } from "lucide-react";

import { Input } from "@/components/ui/input";
import { Select } from "@/components/ui/select";
import type { ClientStatus } from "@/lib/mock/clients";
import { useStore } from "@/lib/mock/store";

type FilterState = {
  query: string;
  inboundId: string;
  status: ClientStatus | "all";
};

type ClientFilterBarProps = {
  value: FilterState;
  onChange: (next: FilterState) => void;
};

export function ClientFilterBar({ value, onChange }: ClientFilterBarProps) {
  const { inbounds } = useStore();
  return (
    <div className="grid grid-cols-1 gap-3 sm:grid-cols-[1fr_220px_180px]">
      <Input
        value={value.query}
        onChange={(e) => onChange({ ...value, query: e.target.value })}
        placeholder="Search by name or UUID"
        mono
        trailing={<Search size={14} />}
      />
      <Select
        value={value.inboundId}
        options={[{ value: "all", label: "All inbounds" }, ...inbounds.map((i) => ({ value: i.id, label: i.remark }))]}
        onChange={(v) => onChange({ ...value, inboundId: v })}
      />
      <Select<ClientStatus | "all">
        value={value.status}
        options={[
          { value: "all", label: "All statuses" },
          { value: "active", label: "Active" },
          { value: "disabled", label: "Disabled" },
          { value: "expired", label: "Expired" }
        ]}
        onChange={(v) => onChange({ ...value, status: v })}
      />
    </div>
  );
}

export type { FilterState };
