import { apiGet } from "./client";

// ------- DTOs -------

export type PanelInfo = {
  name: string;
  version: string;
};

export type HealthStatus = {
  status: string;
};

// ------- API functions -------

export function getPanelInfo(): Promise<PanelInfo> {
  return apiGet<PanelInfo>("/");
}

export function getHealth(): Promise<HealthStatus> {
  return apiGet<HealthStatus>("/health");
}
