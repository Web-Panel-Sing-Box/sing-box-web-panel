import { apiDelete, apiGet, apiPost, apiPut } from "./client";

export type NodeStatus = "unknown" | "online" | "offline";

export type NodeDTO = {
  id: string;
  name: string;
  remark: string;
  scheme: "http" | "https";
  address: string;
  port: number;
  basePath: string;
  enabled: boolean;
  allowPrivateAddress: boolean;
  status: NodeStatus;
  lastHeartbeatAt?: string;
  latencyMs: number;
  panelVersion: string;
  coreVersion: string;
  cpuPct: number;
  ramPct: number;
  uptimeSeconds: number;
  lastError?: string;
  hasApiToken: boolean;
  createdAt: string;
  updatedAt: string;
};

export type NodeRequest = {
  name: string;
  remark?: string;
  scheme: "http" | "https";
  address: string;
  port: number;
  basePath?: string;
  apiToken?: string;
  enabled?: boolean;
  allowPrivateAddress?: boolean;
};

export type NodeSyncResult = {
  nodeId: number;
  inboundCount: number;
  clientCount: number;
  syncedAt: string;
};

export function listNodes(): Promise<NodeDTO[]> {
  return apiGet<NodeDTO[]>("/nodes");
}

export function createNode(body: NodeRequest): Promise<NodeDTO> {
  return apiPost<NodeDTO>("/nodes", body);
}

export function updateNode(id: string, body: NodeRequest): Promise<NodeDTO> {
  return apiPut<NodeDTO>(`/nodes/${id}`, body);
}

export function deleteNode(id: string): Promise<{ message: string }> {
  return apiDelete<{ message: string }>(`/nodes/${id}`);
}

export function toggleNode(id: string): Promise<NodeDTO> {
  return apiPost<NodeDTO>(`/nodes/${id}/toggle`);
}

export function probeNode(id: string): Promise<NodeDTO> {
  return apiPost<NodeDTO>(`/nodes/${id}/probe`);
}

export function syncNode(id: string): Promise<NodeSyncResult> {
  return apiPost<NodeSyncResult>(`/nodes/${id}/sync`);
}
