import { apiGet, apiPost, apiPut, apiDelete } from "./client";
import type { ClientStatus } from "./types";

// ------- DTOs -------

export type ClientDTO = {
  id: string;
  nodeId?: string;
  remoteId?: string;
  name: string;
  uuid: string;
  inboundId: string;
  usedDown: number;
  usedUp: number;
  totalQuota: number;
  expiry: string;
  status: ClientStatus;
  subscription: string;
  subToken?: string;
  enabled?: boolean;
  startAfterFirstUse: boolean;
};

export type ClientCreateRequest = {
  name: string;
  inboundId: string;
  totalQuota?: number;
  expiry?: string;
  startAfterFirstUse?: boolean;
};

export type ClientUpdateRequest = {
  name?: string;
  totalQuota?: number;
  expiry?: string;
  startAfterFirstUse?: boolean;
};

export type ClientSetStatusRequest = {
  status: ClientStatus;
};

export type ClientLinksDTO = {
  shareLink: string;
  subscription: string;
};

export type MessageResponse = {
  message: string;
};

// ------- API functions -------

export function listClients(inboundId?: string): Promise<ClientDTO[]> {
  const query = inboundId ? `?inboundId=${encodeURIComponent(inboundId)}` : "";
  return apiGet<ClientDTO[]>(`/clients${query}`);
}

export function getClient(id: string): Promise<ClientDTO> {
  return apiGet<ClientDTO>(`/clients/${id}`);
}

export function createClient(body: ClientCreateRequest): Promise<ClientDTO> {
  return apiPost<ClientDTO>("/clients", body);
}

export function updateClient(id: string, body: ClientUpdateRequest): Promise<ClientDTO> {
  return apiPut<ClientDTO>(`/clients/${id}`, body);
}

export function deleteClient(id: string): Promise<MessageResponse> {
  return apiDelete<MessageResponse>(`/clients/${id}`);
}

export function resetClientTraffic(id: string): Promise<ClientDTO> {
  return apiPost<ClientDTO>(`/clients/${id}/reset-traffic`);
}

export function setClientStatus(id: string, body: ClientSetStatusRequest): Promise<ClientDTO> {
  return apiPost<ClientDTO>(`/clients/${id}/status`, body);
}

export function getClientLinks(id: string): Promise<ClientLinksDTO> {
  return apiGet<ClientLinksDTO>(`/clients/${id}/links`);
}
