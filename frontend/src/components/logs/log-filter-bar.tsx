
import { Pause, Play, Search } from "lucide-react";

import { Input } from "@/components/ui/input";
import { Select } from "@/components/ui/select";
import { Button } from "@/components/ui/button";
import { useStore } from "@/lib/mock/store";

type LogFilter = {
  query: string;
  level: "all" | "info" | "warn" | "error";
};

type Props = {
  value: LogFilter;
  onChange: (next: LogFilter) => void;
};

export function LogFilterBar({ value, onChange }: Props) {
  const { paused, setPaused } = useStore();
  return (
    <div className="grid grid-cols-1 gap-3 sm:grid-cols-[1fr_180px_auto]">
      <Input
        value={value.query}
        onChange={(e) => onChange({ ...value, query: e.target.value })}
        placeholder="Search logs"
        mono
        trailing={<Search size={14} />}
      />
      <Select<LogFilter["level"]>
        value={value.level}
        onChange={(v) => onChange({ ...value, level: v })}
        options={[
          { value: "all", label: "All levels" },
          { value: "info", label: "Info" },
          { value: "warn", label: "Warn" },
          { value: "error", label: "Error" }
        ]}
      />
      <Button variant="secondary" onClick={() => setPaused(!paused)}>
        {paused ? <Play size={14} /> : <Pause size={14} />}
        {paused ? "Resume" : "Pause"}
      </Button>
    </div>
  );
}

export type { LogFilter };
