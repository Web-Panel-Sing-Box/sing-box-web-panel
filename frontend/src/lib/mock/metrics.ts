export type DiskSegment = {
  label: string;
  usedBytes: number;
  totalBytes: number;
  color: string;
};

export type Metrics = {
  cpu: number;
  ram: number;
  swap: number;
  ramUsedBytes: number;
  ramTotalBytes: number;
  swapUsedBytes: number;
  swapTotalBytes: number;
  uptimeSeconds: number;
  uploadBps: number;
  downloadBps: number;
  todayBytes: number;
  monthBytes: number;
  totalSent: number;
  totalReceived: number;
  diskSegments: DiskSegment[];
  inboundsActive: number;
  totalUsers: number;
  onlineNow: number;
  coreRunning: boolean;
  coreVersion: string;
};

const GB = 1024 ** 3;

export function initialMetrics(): Metrics {
  return {
    cpu: 0.18,
    ram: 0.41,
    swap: 0.07,
    ramUsedBytes: 6.6 * GB,
    ramTotalBytes: 16 * GB,
    swapUsedBytes: 0.14 * GB,
    swapTotalBytes: 2 * GB,
    uptimeSeconds: 12 * 86400 + 4 * 3600 + 33 * 60,
    uploadBps: 4_200_000,
    downloadBps: 8_400_000,
    todayBytes: 6.4 * GB,
    monthBytes: 184 * GB,
    totalSent: 1.42 * 1024 * GB,
    totalReceived: 3.81 * 1024 * GB,
    inboundsActive: 5,
    totalUsers: 124,
    onlineNow: 18,
    coreRunning: true,
    coreVersion: "sing-box 1.10.2",
    diskSegments: [
      { label: "system", usedBytes: 14 * GB, totalBytes: 40 * GB, color: "#ffffffcc" },
      { label: "panel", usedBytes: 3.2 * GB, totalBytes: 10 * GB, color: "#10a37f" },
      { label: "logs", usedBytes: 1.7 * GB, totalBytes: 8 * GB, color: "#22d3ee" },
      { label: "free", usedBytes: 0, totalBytes: 42 * GB, color: "#ffffff14" }
    ]
  };
}

const clamp = (v: number, lo: number, hi: number) => Math.min(hi, Math.max(lo, v));
const drift = (v: number, range: number, lo: number, hi: number) =>
  clamp(v + (Math.random() - 0.5) * range, lo, hi);

export function tickMetrics(prev: Metrics): Metrics {
  const cpu = drift(prev.cpu, 0.06, 0.05, 0.92);
  const ram = drift(prev.ram, 0.02, 0.2, 0.85);
  const swap = drift(prev.swap, 0.01, 0.02, 0.4);
  const uploadBps = clamp(prev.uploadBps + (Math.random() - 0.5) * 600_000, 800_000, 14_000_000);
  const downloadBps = clamp(prev.downloadBps + (Math.random() - 0.5) * 900_000, 1_200_000, 22_000_000);
  const dt = 1;
  return {
    ...prev,
    cpu,
    ram,
    swap,
    uploadBps,
    downloadBps,
    uptimeSeconds: prev.uptimeSeconds + dt,
    todayBytes: prev.todayBytes + (uploadBps + downloadBps) * dt * 0.05,
    monthBytes: prev.monthBytes + (uploadBps + downloadBps) * dt * 0.05,
    totalSent: prev.totalSent + uploadBps * dt * 0.5,
    totalReceived: prev.totalReceived + downloadBps * dt * 0.5,
    ramUsedBytes: ram * prev.ramTotalBytes,
    swapUsedBytes: swap * prev.swapTotalBytes,
    onlineNow: clamp(Math.round(prev.onlineNow + (Math.random() - 0.5) * 1.4), 14, 22)
  };
}

export type TrafficPoint = { t: number; up: number; down: number };

export function seedTrafficHistory(): TrafficPoint[] {
  const now = Date.now();
  return Array.from({ length: 30 }, (_, i) => ({
    t: now - (30 - i) * 1000,
    up: 3_000_000 + Math.random() * 4_000_000,
    down: 6_000_000 + Math.random() * 8_000_000
  }));
}
