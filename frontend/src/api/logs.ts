import { apiGet, apiPost } from "./client";
import type { LogLevel, LogSource } from "./types";

// ------- DTOs -------

export type LogEntryDTO = {
  id: string;
  t: number;
  level: LogLevel;
  source: LogSource;
  message: string;
  requestId?: string;
  fields?: Record<string, string>;
};

export type LogQuery = {
  level?: LogLevel | "all";
  source?: LogSource | "all";
  q?: string;
  limit?: number;
};

export type FrontendLogRequest = {
  level: LogLevel;
  message: string;
  fields?: Record<string, string>;
};

// ------- API functions -------

export function getPanelLogs(query: LogQuery = {}): Promise<LogEntryDTO[]> {
  const params = new URLSearchParams();
  if (query.level && query.level !== "all") params.set("level", query.level);
  if (query.source && query.source !== "all") params.set("source", query.source);
  if (query.q) params.set("q", query.q);
  if (query.limit) params.set("limit", String(query.limit));
  const suffix = params.toString() ? `?${params.toString()}` : "";
  return apiGet<LogEntryDTO[]>(`/logs${suffix}`);
}

export function postFrontendLog(body: FrontendLogRequest): Promise<{ message: string }> {
  return apiPost<{ message: string }>("/logs/frontend", body);
}
