
import { Search } from "lucide-react";

import { Input } from "@/components/ui/input";
import { Select } from "@/components/ui/select";
import type { ClientStatus } from "@/lib/mock/clients";
import { useInbounds } from "@/lib/mock/store";
import { useI18n } from "@/lib/i18n";
import type { ClientFilterState } from "@/hooks/useClientFilter";

type FilterState = ClientFilterState;

type ClientFilterBarProps = {
  value: FilterState;
  onChange: (next: FilterState) => void;
};

export function ClientFilterBar({ value, onChange }: ClientFilterBarProps) {
  const inbounds = useInbounds();
  const { t } = useI18n();
  return (
    <div className="grid grid-cols-1 gap-3 sm:grid-cols-[1fr_220px_180px]">
      <Input
        value={value.query}
        onChange={(e) => onChange({ ...value, query: e.target.value })}
        placeholder={t("clients.search")}
        mono
        trailing={<Search size={14} />}
      />
      <Select
        value={value.inboundId}
        options={[{ value: "all", label: t("clients.allInbounds") }, ...inbounds.map((i) => ({ value: i.id, label: i.remark }))]}
        onChange={(v) => onChange({ ...value, inboundId: v })}
      />
      <Select<ClientStatus | "all">
        value={value.status}
        options={[
          { value: "all", label: t("clients.allStatuses") },
          { value: "active", label: t("common.active") },
          { value: "disabled", label: t("common.disabled") },
          { value: "expired", label: t("common.expired") }
        ]}
        onChange={(v) => onChange({ ...value, status: v })}
      />
    </div>
  );
}

export type { FilterState };
