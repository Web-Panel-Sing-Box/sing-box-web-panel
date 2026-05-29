import { useMemo } from "react";

import type { LogEntry } from "@/lib/mock/logs";

export type LogFilter = {
  query: string;
  level: "all" | "info" | "warn" | "error";
};

export function useLogFilter(logs: LogEntry[], filter: LogFilter): LogEntry[] {
  return useMemo(() => {
    const q = filter.query.trim().toLowerCase();
    return logs.filter((l) => {
      if (filter.level !== "all" && l.level !== filter.level) return false;
      if (!q) return true;
      return l.message.toLowerCase().includes(q);
    });
  }, [logs, filter]);
}
