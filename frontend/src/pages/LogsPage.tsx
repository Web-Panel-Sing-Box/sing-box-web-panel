
import { useState } from "react";

import { LogFilterBar, type LogFilter } from "@/components/logs/log-filter-bar";
import { LogViewer } from "@/components/logs/log-viewer";
import { useI18n } from "@/lib/i18n";

export function LogsPage() {
  const [filter, setFilter] = useState<LogFilter>({ query: "", level: "all" });
  const { t } = useI18n();
  return (
    <div className="mx-auto flex max-w-[1320px] flex-col gap-4">
      <div>
        <h2 className="text-2xl font-semibold text-ink-primary">{t("logs.title")}</h2>
        <p className="mt-1 text-sm text-ink-tertiary">{t("logs.description")}</p>
      </div>
      <LogFilterBar value={filter} onChange={setFilter} />
      <LogViewer filter={filter} />
    </div>
  );
}
