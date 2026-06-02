import { apiGet, apiPut } from "./client";

export type SettingsDTO = Record<string, string>;

export function getSettings(): Promise<SettingsDTO> {
  return apiGet<SettingsDTO>("/settings");
}

export function saveSettings(settings: SettingsDTO): Promise<{ ok: string }> {
  return apiPut<{ ok: string }>("/settings", settings);
}
