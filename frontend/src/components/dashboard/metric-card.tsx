
import { Cpu, Database, HardDrive, Network } from "lucide-react";
import { Area, AreaChart, ResponsiveContainer } from "recharts";

import { useStore } from "@/lib/mock/store";
import { Card, CardHeader, CardLabel } from "@/components/ui/card";
import { Progress } from "@/components/ui/progress";
import { AnimatedNumber } from "@/components/ui/animated-number";
import { formatBytes, formatPercent } from "@/lib/format";

export function CpuCard() {
  const { metrics } = useStore();
  const value = metrics.cpu * 100;
  return (
    <Card>
      <CardHeader>
        <CardLabel>CPU</CardLabel>
        <Cpu size={16} className="text-ink-tertiary" />
      </CardHeader>
      <div className="mb-4 text-3xl font-semibold text-ink-primary">
        <AnimatedNumber value={value} format={(n) => `${n.toFixed(1)}%`} />
      </div>
      <Progress value={value} />
      <p className="mt-3 text-xs text-ink-tertiary">8 cores · {formatPercent(value)} usage</p>
    </Card>
  );
}

export function RamCard() {
  const { metrics } = useStore();
  const ram = metrics.ram * 100;
  const swap = metrics.swap * 100;
  return (
    <Card>
      <CardHeader>
        <CardLabel>Memory</CardLabel>
        <Database size={16} className="text-ink-tertiary" />
      </CardHeader>
      <div className="mb-4 text-3xl font-semibold text-ink-primary">
        <AnimatedNumber value={ram} format={(n) => `${n.toFixed(1)}%`} />
      </div>
      <div className="space-y-3">
        <div>
          <div className="mb-1 flex justify-between text-xs text-ink-tertiary">
            <span>RAM</span>
            <span className="font-mono">{formatBytes(metrics.ramUsedBytes)} / {formatBytes(metrics.ramTotalBytes)}</span>
          </div>
          <Progress value={ram} height={6} />
        </div>
        <div>
          <div className="mb-1 flex justify-between text-xs text-ink-tertiary">
            <span>Swap</span>
            <span className="font-mono">{formatBytes(metrics.swapUsedBytes)} / {formatBytes(metrics.swapTotalBytes)}</span>
          </div>
          <div className="w-3/5">
            <Progress value={swap} height={4} />
          </div>
        </div>
      </div>
    </Card>
  );
}

export function DiskCard() {
  const { metrics } = useStore();
  const total = metrics.diskSegments.reduce((acc, s) => acc + s.totalBytes, 0);
  return (
    <Card>
      <CardHeader>
        <CardLabel>Disk</CardLabel>
        <HardDrive size={16} className="text-ink-tertiary" />
      </CardHeader>
      <div className="mb-4 text-3xl font-semibold text-ink-primary">
        <AnimatedNumber
          value={(metrics.diskSegments.reduce((acc, s) => acc + s.usedBytes, 0) / total) * 100}
          format={(n) => `${n.toFixed(1)}%`}
        />
      </div>
      <div className="mb-3 flex h-2 w-full overflow-hidden rounded-full bg-white/10">
        {metrics.diskSegments.map((seg) => (
          <div
            key={seg.label}
            className="h-full transition-all duration-500"
            style={{ width: `${(seg.totalBytes / total) * 100}%`, background: seg.color }}
          />
        ))}
      </div>
      <ul className="grid grid-cols-2 gap-y-1 text-xs">
        {metrics.diskSegments.map((seg) => (
          <li key={seg.label} className="flex items-center gap-2 text-ink-tertiary">
            <span className="size-1.5 rounded-full" style={{ background: seg.color }} />
            <span className="truncate">{seg.label}</span>
            <span className="ml-auto font-mono text-ink-secondary">{formatBytes(seg.totalBytes, 0)}</span>
          </li>
        ))}
      </ul>
    </Card>
  );
}

export function TrafficSplitCard() {
  const { metrics, history } = useStore();
  const spark = history.slice(-20).map((p, i) => ({ i, v: p.down + p.up }));
  return (
    <Card>
      <CardHeader>
        <CardLabel>Traffic</CardLabel>
        <Network size={16} className="text-ink-tertiary" />
      </CardHeader>
      <div className="grid grid-cols-2 gap-4">
        <div>
          <div className="text-xs text-ink-tertiary">Today</div>
          <div className="font-mono text-lg text-ink-primary">{formatBytes(metrics.todayBytes)}</div>
        </div>
        <div>
          <div className="text-xs text-ink-tertiary">This month</div>
          <div className="font-mono text-lg text-ink-primary">{formatBytes(metrics.monthBytes)}</div>
        </div>
      </div>
      <div className="mt-3 h-14 w-full">
        <ResponsiveContainer width="100%" height="100%">
          <AreaChart data={spark}>
            <defs>
              <linearGradient id="spark" x1="0" y1="0" x2="0" y2="1">
                <stop offset="0%" stopColor="#22d3ee" stopOpacity={0.5} />
                <stop offset="100%" stopColor="#22d3ee" stopOpacity={0} />
              </linearGradient>
            </defs>
            <Area type="monotone" dataKey="v" stroke="#22d3ee" strokeWidth={1.5} fill="url(#spark)" isAnimationActive={false} />
          </AreaChart>
        </ResponsiveContainer>
      </div>
    </Card>
  );
}
