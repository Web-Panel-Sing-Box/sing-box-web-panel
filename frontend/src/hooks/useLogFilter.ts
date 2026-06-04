import { useMemo } from "react";

import type { LogEntry } from "@/lib/store";

export type LogFilter = {
  query: string;
  level: "all" | "info" | "warn" | "error";
  source: "all" | "panel" | "core" | "frontend";
};

export function useLogFilter(logs: LogEntry[], filter: LogFilter): LogEntry[] {
  return useMemo(() => {
    const q = filter.query.trim().toLowerCase();
    return logs.filter((l) => {
      if (filter.level !== "all" && l.level !== filter.level) return false;
      if (filter.source !== "all" && l.source !== filter.source) return false;
      if (!q) return true;
      if (l.message.toLowerCase().includes(q)) return true;
      if (l.requestId?.toLowerCase().includes(q)) return true;
      if (l.source.toLowerCase().includes(q)) return true;
      return Object.entries(l.fields ?? {}).some(([key, value]) =>
        key.toLowerCase().includes(q) || value.toLowerCase().includes(q)
      );
    });
  }, [logs, filter]);
}
