import { apiGet } from "./client";
import type { DiskSegment } from "./types";

// ------- DTOs -------

export type MetricsDTO = {
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

export type TrafficPoint = {
  t: number;
  up: number;
  down: number;
};

export type TrafficHistoryDTO = TrafficPoint[];

// ------- API functions -------

export function getMetrics(): Promise<MetricsDTO> {
  return apiGet<MetricsDTO>("/dashboard/metrics");
}

export function getTrafficHistory(): Promise<TrafficHistoryDTO> {
  return apiGet<TrafficHistoryDTO>("/dashboard/traffic");
}
