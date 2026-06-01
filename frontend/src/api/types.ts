export type Protocol = "vless" | "naive" | "hysteria2";
export type Transmission = "tcp" | "mkcp" | "grpc" | "ws" | "xhttp" | "httpupgrade";
export type TlsMode = "none" | "tls" | "reality";
export type ClientStatus = "active" | "disabled" | "expired";
export type TrafficReset = "never" | "hourly" | "daily" | "weekly" | "monthly";
export type LogLevel = "info" | "warn" | "error";

export const PROTOCOL_OPTIONS: { value: Protocol; label: string }[] = [
  { value: "naive", label: "Naive Proxy" },
  { value: "vless", label: "VLESS" },
  { value: "hysteria2", label: "Hysteria2" },
];

export const TRANSMISSION_OPTIONS: { value: Transmission; label: string }[] = [
  { value: "tcp", label: "TCP (RAW)" },
  { value: "mkcp", label: "mKCP" },
  { value: "grpc", label: "gRPC" },
  { value: "ws", label: "WebSocket" },
  { value: "xhttp", label: "XHTTP" },
  { value: "httpupgrade", label: "HTTPUpgrade" },
];

export const TRAFFIC_RESET_OPTIONS: { value: TrafficReset; label: string }[] = [
  { value: "never", label: "Never" },
  { value: "hourly", label: "Hourly" },
  { value: "daily", label: "Daily" },
  { value: "weekly", label: "Weekly" },
  { value: "monthly", label: "Monthly" },
];

export type DiskSegment = {
  label: string;
  usedBytes: number;
  totalBytes: number;
  color: string;
};

export type MeResponse = {
  id: number;
  username: string;
  is_totp_enabled: boolean;
  totp_confirmed_at?: string;
  created_at: string;
};

export type ErrorResponse = {
  error: string;
};
