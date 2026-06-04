import { apiGet, apiPost } from "./client";

export type PanelVersionDTO = {
  currentVersion: string;
  latestVersion: string;
  updateAvailable: boolean;
  releaseURL: string;
  checkedAt: string;
  status: "up_to_date" | "update_available" | "running" | "failed" | "updated" | "development" | "check_failed" | "not_configured";
};

export function getPanelVersion(): Promise<PanelVersionDTO> {
  return apiGet<PanelVersionDTO>("/panel/version");
}

export function startPanelUpdate(): Promise<PanelVersionDTO> {
  return apiPost<PanelVersionDTO>("/panel/update");
}
