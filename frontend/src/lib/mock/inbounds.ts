export type Protocol = "vless" | "naive" | "hysteria2";

// sing-box v2ray transports valid for VLESS. "tcp" == raw/none.
// mKCP and XHTTP are Xray-only and rejected by sing-box, so they are not listed.
export type VlessTransport = "tcp" | "ws" | "grpc" | "http" | "httpupgrade";

export type TlsMode = "none" | "tls" | "reality";

// naive `network`
export type Network = "tcp" | "udp" | "both";

// vless `flow`
export type Flow = "" | "xtls-rprx-vision";

// hysteria2 `obfs.type`
export type ObfsType = "none" | "salamander";

// naive `quic_congestion_control`
export type QuicCc = "bbr" | "bbr_standard" | "bbr2" | "bbr2_variant" | "cubic" | "reno";

export type Inbound = {
  id: string;
  remark: string;
  protocol: Protocol;
  port: number;
  tls: TlsMode;
  sni?: string;
  dest?: string; // reality handshake target (vless + reality)

  // vless
  transport?: VlessTransport;
  flow?: Flow;

  // naive + hysteria2 share username/password auth
  username?: string;
  password?: string;

  // naive
  network?: Network;
  quicCc?: QuicCc;

  // hysteria2
  obfsType?: ObfsType;
  obfsPassword?: string;
  upMbps?: number;
  downMbps?: number;

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
    transport: "tcp",
    flow: "xtls-rprx-vision",
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
    tls: "tls",
    sni: "panel.example",
    username: "user-001",
    password: "goofy_ahh_password",
    obfsType: "salamander",
    obfsPassword: "cry_me_a_r1ver",
    upMbps: 100,
    downMbps: 100,
    enabled: true,
    clientCount: 22,
    createdAt: "2026-03-20T09:45:00Z"
  },
  {
    id: "ib_03",
    remark: "amsterdam-naive-01",
    protocol: "naive",
    port: 38119,
    tls: "tls",
    sni: "panel.example",
    username: "user-001",
    password: "password",
    network: "both",
    quicCc: "bbr",
    enabled: true,
    clientCount: 31,
    createdAt: "2026-04-02T18:00:00Z"
  },
  {
    id: "ib_04",
    remark: "frankfurt-ws-01",
    protocol: "vless",
    port: 27440,
    transport: "ws",
    flow: "",
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
    transport: "grpc",
    flow: "",
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
    tls: "tls",
    sni: "panel.example",
    username: "user-001",
    password: "warsaw_pw_01",
    obfsType: "none",
    upMbps: 200,
    downMbps: 200,
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

// VLESS-only. sing-box v2ray transports.
export const VLESS_TRANSPORT_OPTIONS: { value: VlessTransport; label: string }[] = [
  { value: "tcp", label: "TCP (RAW)" },
  { value: "ws", label: "WebSocket" },
  { value: "grpc", label: "gRPC" },
  { value: "http", label: "HTTP/2" },
  { value: "httpupgrade", label: "HTTPUpgrade" }
];

export const FLOW_OPTIONS: { value: Flow; label: string }[] = [
  { value: "", label: "None" },
  { value: "xtls-rprx-vision", label: "xtls-rprx-vision" }
];

export const NETWORK_OPTIONS: { value: Network; label: string }[] = [
  { value: "both", label: "TCP + UDP" },
  { value: "tcp", label: "TCP" },
  { value: "udp", label: "UDP" }
];

export const QUIC_CC_OPTIONS: { value: QuicCc; label: string }[] = [
  { value: "bbr", label: "BBR" },
  { value: "bbr_standard", label: "BBR (standard)" },
  { value: "bbr2", label: "BBRv2" },
  { value: "bbr2_variant", label: "BBRv2 (variant)" },
  { value: "cubic", label: "CUBIC" },
  { value: "reno", label: "Reno" }
];

export const OBFS_OPTIONS: { value: ObfsType; label: string }[] = [
  { value: "none", label: "None" },
  { value: "salamander", label: "Salamander" }
];

export const TRAFFIC_RESET_OPTIONS = [
  { value: "never", label: "Never" },
  { value: "hourly", label: "Hourly" },
  { value: "daily", label: "Daily" },
  { value: "weekly", label: "Weekly" },
  { value: "monthly", label: "Monthly" }
];

// Default transport / network when switching to a protocol that needs one.
export const DEFAULT_VLESS_TRANSPORT: VlessTransport = "tcp";
export const DEFAULT_NETWORK: Network = "both";
export const DEFAULT_QUIC_CC: QuicCc = "bbr";
