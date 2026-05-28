export type CoreStatus = {
  mode: string;
  running: boolean;
  detail: string;
};

export type Metrics = {
  cpu_percent: number;
  memory_percent: number;
  upload_bps: number;
  download_bps: number;
  active_users: number;
  core: CoreStatus;
};

export type Inbound = {
  id: number;
  protocol: string;
  tag: string;
  listen: string;
  port: number;
  status: "active" | "disabled";
  tls_enabled: boolean;
  server_name?: string | null;
  reality_enabled: boolean;
  reality_public_key?: string | null;
  reality_short_id?: string | null;
  created_at: string;
  updated_at: string;
};

export type User = {
  id: number;
  inbound_id: number;
  username: string;
  uuid: string;
  password?: string | null;
  total_traffic: number;
  used_traffic: number;
  expire_time?: string | null;
  status: "active" | "disabled" | "expired" | "limited";
  ip_limit: number;
  created_at: string;
  updated_at: string;
};

export type Setting = {
  key: string;
  value: string | null;
  is_secret: boolean;
};

async function request<T>(path: string, init?: RequestInit): Promise<T> {
  const response = await fetch(path, {
    ...init,
    credentials: "include",
    headers: {
      "Content-Type": "application/json",
      ...(init?.headers || {})
    }
  });
  if (!response.ok) {
    const detail = await response.text();
    throw new Error(detail || response.statusText);
  }
  return response.json() as Promise<T>;
}

export const api = {
  login: (username: string, password: string) =>
    request<{ admin: { id: number; username: string } }>("/api/auth/login", {
      method: "POST",
      body: JSON.stringify({ username, password })
    }),
  me: () => request<{ id: number; username: string }>("/api/auth/me"),
  metrics: () => request<Metrics>("/api/dashboard/metrics"),
  users: () => request<User[]>("/api/users"),
  inbounds: () => request<Inbound[]>("/api/inbounds"),
  settings: () => request<Setting[]>("/api/settings"),
  links: (userId: number) => request<{ links: string[]; subscription_url?: string }>(`/api/users/${userId}/links`),
  resetTraffic: (userId: number) =>
    request<User>(`/api/users/${userId}/reset-traffic`, { method: "POST" }),
  disableUser: (userId: number) => request<User>(`/api/users/${userId}/disable`, { method: "POST" }),
  core: (action: "start" | "stop" | "restart" | "reload") =>
    request<CoreStatus | { message: string }>(`/api/core/${action}`, { method: "POST" }),
  logs: () => request<string[]>("/api/logs/sing-box?lines=80")
};
