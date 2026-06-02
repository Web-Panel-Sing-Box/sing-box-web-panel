export type Protocol = "vless" | "naive" | "hysteria2";
// VLESS v2ray transports valid for sing-box. mKCP and XHTTP are Xray-only and
// rejected by sing-box, so they are intentionally excluded.
export type Transmission = "tcp" | "grpc" | "ws" | "httpupgrade";
export type TlsMode = "none" | "tls" | "reality";
export type ClientStatus = "active" | "disabled" | "expired";
export type TrafficReset = "never" | "hourly" | "daily" | "weekly" | "monthly";
export type LogLevel = "info" | "warn" | "error";

// naive `network`
export type Network = "tcp" | "udp" | "both";
// vless `flow`
export type Flow = "" | "xtls-rprx-vision";
// hysteria2 `obfs.type`
export type ObfsType = "none" | "salamander";
// naive `quic_congestion_control`
export type QuicCc = "bbr" | "bbr_standard" | "bbr2" | "bbr2_variant" | "cubic" | "reno";

export const PROTOCOL_OPTIONS: { value: Protocol; label: string }[] = [
  { value: "naive", label: "Naive Proxy" },
  { value: "vless", label: "VLESS" },
  { value: "hysteria2", label: "Hysteria2" },
];

// VLESS-only transport selector.
export const TRANSMISSION_OPTIONS: { value: Transmission; label: string }[] = [
  { value: "tcp", label: "TCP (RAW)" },
  { value: "ws", label: "WebSocket" },
  { value: "grpc", label: "gRPC" },
  { value: "httpupgrade", label: "HTTPUpgrade" },
];

export const NETWORK_OPTIONS: { value: Network; label: string }[] = [
  { value: "both", label: "TCP + UDP" },
  { value: "tcp", label: "TCP" },
  { value: "udp", label: "UDP" },
];

export const FLOW_OPTIONS: { value: Flow; label: string }[] = [
  { value: "", label: "None" },
  { value: "xtls-rprx-vision", label: "xtls-rprx-vision" },
];

export const QUIC_CC_OPTIONS: { value: QuicCc; label: string }[] = [
  { value: "bbr", label: "BBR" },
  { value: "bbr_standard", label: "BBR (standard)" },
  { value: "bbr2", label: "BBRv2" },
  { value: "bbr2_variant", label: "BBRv2 (variant)" },
  { value: "cubic", label: "CUBIC" },
  { value: "reno", label: "Reno" },
];

export const OBFS_OPTIONS: { value: ObfsType; label: string }[] = [
  { value: "none", label: "None" },
  { value: "salamander", label: "Salamander" },
];

export const DEFAULT_TRANSMISSION: Transmission = "tcp";
export const DEFAULT_NETWORK: Network = "both";
export const DEFAULT_QUIC_CC: QuicCc = "bbr";

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
