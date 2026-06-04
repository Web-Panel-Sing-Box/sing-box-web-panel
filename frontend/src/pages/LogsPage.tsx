import { useState } from "react";

import { LogFilterBar, type LogFilter } from "@/components/logs/log-filter-bar";
import { LogViewer } from "@/components/logs/log-viewer";
import { useI18n } from "@/lib/i18n";

export function LogsPage() {
  const [filter, setFilter] = useState<LogFilter>({ query: "", level: "all", source: "all" });
  const { t } = useI18n();
  return (
    <div className="mx-auto flex h-[calc(100dvh-72px)] min-h-[420px] w-full max-w-[1320px] flex-col gap-4 lg:h-[calc(100dvh-48px)]">
      <h2 className="text-2xl font-semibold text-ink-primary">
        {t("logs.title")}
      </h2>
      <LogFilterBar value={filter} onChange={setFilter} />
      <LogViewer filter={filter} />
    </div>
  );
}
