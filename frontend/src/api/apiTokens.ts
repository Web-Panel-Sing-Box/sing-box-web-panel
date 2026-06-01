import { apiDelete, apiGet, apiPost } from "./client";

export type APITokenDTO = {
  id: string;
  name: string;
  tokenPrefix: string;
  scopes: string;
  enabled: boolean;
  lastUsedAt?: string;
  createdAt: string;
};

export type CreatedAPITokenDTO = APITokenDTO & {
  token: string;
};

export function listAPITokens(): Promise<APITokenDTO[]> {
  return apiGet<APITokenDTO[]>("/api-tokens");
}

export function createAPIToken(body: {
  name: string;
  scopes?: string;
}): Promise<CreatedAPITokenDTO> {
  return apiPost<CreatedAPITokenDTO>("/api-tokens", body);
}

export function setAPITokenEnabled(
  id: string,
  enabled: boolean,
): Promise<{ message: string }> {
  return apiPost<{ message: string }>(`/api-tokens/${id}/toggle`, { enabled });
}

export function deleteAPIToken(id: string): Promise<{ message: string }> {
  return apiDelete<{ message: string }>(`/api-tokens/${id}`);
}
