import { apiGet } from "./client";
import type { LogLevel } from "./types";

// ------- DTOs -------

export type LogEntryDTO = {
  id: string;
  t: number;
  level: LogLevel;
  message: string;
};

// ------- API functions -------

export function getPanelLogs(): Promise<LogEntryDTO[]> {
  return apiGet<LogEntryDTO[]>("/logs");
}
