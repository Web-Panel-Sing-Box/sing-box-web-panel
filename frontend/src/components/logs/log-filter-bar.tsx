
import { Pause, Play, Search } from "lucide-react";

import { Input } from "@/components/ui/input";
import { Select } from "@/components/ui/select";
import { Button } from "@/components/ui/button";
import { useRuntime, useStoreActions } from "@/lib/store";
import { useI18n } from "@/lib/i18n";

type LogFilter = {
  query: string;
  level: "all" | "info" | "warn" | "error";
  source: "all" | "panel" | "core" | "frontend";
};

type Props = {
  value: LogFilter;
  onChange: (next: LogFilter) => void;
};

export function LogFilterBar({ value, onChange }: Props) {
  const { paused } = useRuntime();
  const { setPaused } = useStoreActions();
  const { t } = useI18n();
  return (
    <div className="grid grid-cols-1 gap-3 sm:grid-cols-[1fr_160px_160px_auto]">
      <Input
        value={value.query}
        onChange={(e) => onChange({ ...value, query: e.target.value })}
        placeholder={t("logs.search")}
        mono
        trailing={<Search size={14} />}
      />
      <Select<LogFilter["level"]>
        value={value.level}
        onChange={(v) => onChange({ ...value, level: v })}
        options={[
          { value: "all", label: t("logs.allLevels") },
          { value: "info", label: "Info" },
          { value: "warn", label: "Warn" },
          { value: "error", label: "Error" }
        ]}
      />
      <Select<LogFilter["source"]>
        value={value.source}
        onChange={(v) => onChange({ ...value, source: v })}
        options={[
          { value: "all", label: t("logs.allSources") },
          { value: "panel", label: t("logs.sourcePanel") },
          { value: "core", label: t("logs.sourceCore") },
          { value: "frontend", label: t("logs.sourceFrontend") }
        ]}
      />
      <Button variant="secondary" onClick={() => setPaused(!paused)}>
        {paused ? <Play size={14} /> : <Pause size={14} />}
        {paused ? t("logs.resume") : t("logs.pause")}
      </Button>
    </div>
  );
}

export type { LogFilter };
