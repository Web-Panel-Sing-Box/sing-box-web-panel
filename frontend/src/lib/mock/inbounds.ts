export type Protocol = "vless" | "naive" | "hysteria2";
export type Transmission = "tcp" | "mkcp" | "grpc" | "ws" | "xhttp" | "httpupgrade";
export type TlsMode = "none" | "tls" | "reality";

export type Inbound = {
  id: string;
  remark: string;
  protocol: Protocol;
  port: number;
  transmission: Transmission;
  tls: TlsMode;
  sni?: string;
  dest?: string;
  enabled: boolean;
  clientCount: number;
  createdAt: string;
};

export const SEED_INBOUNDS: Inbound[] = [
  {
    id: "ib_01",
    remark: "berlin-edge-01",
    protocol: "vless",
    port: 44321,
    transmission: "tcp",
    tls: "reality",
    sni: "www.cloudflare.com",
    dest: "www.cloudflare.com:443",
    enabled: true,
    clientCount: 18,
    createdAt: "2026-03-12T14:11:00Z"
  },
  {
    id: "ib_02",
    remark: "tokyo-relay-02",
    protocol: "hysteria2",
    port: 51005,
    transmission: "tcp",
    tls: "tls",
    sni: "panel.example",
    enabled: true,
    clientCount: 22,
    createdAt: "2026-03-20T09:45:00Z"
  },
  {
    id: "ib_03",
    remark: "amsterdam-naive-01",
    protocol: "naive",
    port: 38119,
    transmission: "tcp",
    tls: "tls",
    enabled: true,
    clientCount: 31,
    createdAt: "2026-04-02T18:00:00Z"
  },
  {
    id: "ib_04",
    remark: "frankfurt-ws-01",
    protocol: "vless",
    port: 27440,
    transmission: "ws",
    tls: "tls",
    sni: "panel.example",
    enabled: false,
    clientCount: 9,
    createdAt: "2026-04-18T12:24:00Z"
  },
  {
    id: "ib_05",
    remark: "singapore-grpc-01",
    protocol: "vless",
    port: 47711,
    transmission: "grpc",
    tls: "reality",
    sni: "www.microsoft.com",
    dest: "www.microsoft.com:443",
    enabled: true,
    clientCount: 28,
    createdAt: "2026-05-04T07:32:00Z"
  },
  {
    id: "ib_06",
    remark: "warsaw-hy2-01",
    protocol: "hysteria2",
    port: 19087,
    transmission: "tcp",
    tls: "tls",
    enabled: true,
    clientCount: 16,
    createdAt: "2026-05-10T22:01:00Z"
  }
];

export const PROTOCOL_OPTIONS: { value: Protocol; label: string }[] = [
  { value: "naive", label: "Naive Proxy" },
  { value: "vless", label: "VLESS" },
  { value: "hysteria2", label: "Hysteria2" }
];

export const TRANSMISSION_OPTIONS: { value: Transmission; label: string }[] = [
  { value: "tcp", label: "TCP (RAW)" },
  { value: "mkcp", label: "mKCP" },
  { value: "grpc", label: "gRPC" },
  { value: "ws", label: "WebSocket" },
  { value: "xhttp", label: "XHTTP" },
  { value: "httpupgrade", label: "HTTPUpgrade" }
];

export const TRAFFIC_RESET_OPTIONS = [
  { value: "never", label: "Never" },
  { value: "hourly", label: "Hourly" },
  { value: "daily", label: "Daily" },
  { value: "weekly", label: "Weekly" },
  { value: "monthly", label: "Monthly" }
];
