import { apiGet, apiPost } from "./client";

// ------- DTOs -------

export type CoreStatusDTO = {
  running: boolean;
  pid: number;
  version: string;
  uptimeSeconds: number;
  lastError?: string;
};

export type MessageResponse = {
  message: string;
};

export type CoreLogsResponse = {
  lines: string[];
  total: number;
  hasMore: boolean;
};

// ------- API functions -------

export function getCoreStatus(): Promise<CoreStatusDTO> {
  return apiGet<CoreStatusDTO>("/core/status");
}

export function startCore(): Promise<MessageResponse> {
  return apiPost<MessageResponse>("/core/start");
}

export function stopCore(): Promise<MessageResponse> {
  return apiPost<MessageResponse>("/core/stop");
}

export function restartCore(): Promise<MessageResponse> {
  return apiPost<MessageResponse>("/core/restart");
}

export function reloadCore(): Promise<MessageResponse> {
  return apiPost<MessageResponse>("/core/reload");
}

export function getCoreVersion(): Promise<{ version: string }> {
  return apiGet<{ version: string }>("/core/version");
}

export function getCoreConfig(): Promise<unknown> {
  return apiGet<unknown>("/core/config");
}

export function getCoreLogs(offset = 0, limit = 200): Promise<CoreLogsResponse> {
  return apiGet<CoreLogsResponse>(`/core/logs?offset=${offset}&limit=${limit}`);
}
