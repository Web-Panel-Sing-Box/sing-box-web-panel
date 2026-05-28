export type LogLevel = "info" | "warn" | "error";

export type LogEntry = {
  id: string;
  t: number;
  level: LogLevel;
  message: string;
};

const INFO_LINES = [
  "inbound vless-reality:44321 accepted tcp connection from 203.0.113.42",
  "router matched rule [country:cn] for 198.51.100.7, action: direct",
  "subscription rebuilt for user kira (12 keys)",
  "core healthcheck ok, latency 4ms",
  "dns_outbound resolved cloudflare-dns.com -> 1.1.1.1",
  "inbound naive:38119 client alex_kim authenticated",
  "config_revision committed (sha256: 9f3d…2a17)",
  "traffic flush: 18 rows written in 22ms",
  "core started successfully, 6 inbounds active"
];
const WARN_LINES = [
  "tls handshake retried for sni www.cloudflare.com",
  "client tomek approaching 90% of monthly quota",
  "udp socket recv timeout, retrying",
  "geo database is older than 30 days, consider refreshing"
];
const ERROR_LINES = [
  "inbound vless-ws:27440 failed to bind: address already in use",
  "panic recovered in route handler: nil pointer dereference",
  "sing-box reload failed: invalid reality short_id"
];

function pick<T>(arr: T[]): T {
  return arr[Math.floor(Math.random() * arr.length)];
}

function uid() {
  return Math.random().toString(36).slice(2, 10);
}

export function buildLog(level: LogLevel, message: string, t = Date.now()): LogEntry {
  return { id: uid(), t, level, message };
}

export function seedLogs(): LogEntry[] {
  const now = Date.now();
  const out: LogEntry[] = [];
  for (let i = 0; i < 40; i++) {
    const r = Math.random();
    const level: LogLevel = r < 0.08 ? "error" : r < 0.22 ? "warn" : "info";
    const message =
      level === "info" ? pick(INFO_LINES) : level === "warn" ? pick(WARN_LINES) : pick(ERROR_LINES);
    out.push({ id: uid(), t: now - (40 - i) * 1500, level, message });
  }
  return out;
}

export function nextLog(): LogEntry {
  const r = Math.random();
  const level: LogLevel = r < 0.06 ? "error" : r < 0.18 ? "warn" : "info";
  const message =
    level === "info" ? pick(INFO_LINES) : level === "warn" ? pick(WARN_LINES) : pick(ERROR_LINES);
  return buildLog(level, message);
}
