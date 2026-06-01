
import { Area, AreaChart, CartesianGrid, ResponsiveContainer, Tooltip, XAxis, YAxis } from "recharts";

import { useMetrics } from "@/lib/store";
import { Card, CardHeader, CardTitle } from "@/components/ui/card";
import { formatSpeed } from "@/lib/format";
import { useI18n } from "@/lib/i18n";

function fmtAxis(value: number) {
  if (value >= 1024 ** 3) return `${(value / 1024 ** 3).toFixed(1)}G`;
  if (value >= 1024 ** 2) return `${(value / 1024 ** 2).toFixed(0)}M`;
  if (value >= 1024) return `${(value / 1024).toFixed(0)}K`;
  return `${value}`;
}

function fmtTime(t: number) {
  const d = new Date(t);
  return d.toLocaleTimeString("en-US", { hour12: false, minute: "2-digit", second: "2-digit" });
}

function ChartTooltip({ active, payload, incoming, outgoing }: any) {
  if (!active || !payload?.length) return null;
  const down = payload.find((p: any) => p.dataKey === "down")?.value ?? 0;
  const up = payload.find((p: any) => p.dataKey === "up")?.value ?? 0;
  return (
    <div className="rounded-lg border border-black/10 bg-white px-3 py-2 text-canvas shadow-pop">
      <div className="mb-1 flex items-center justify-between gap-4 text-[11px] text-canvas/60">
        <span>{fmtTime(payload[0]?.payload?.t ?? Date.now())}</span>
      </div>
      <div className="flex items-center gap-2 text-xs">
        <span className="size-2 rounded-full" style={{ background: "#22d3ee" }} />
        <span className="text-canvas/70">{incoming}</span>
        <span className="ml-auto font-mono text-canvas">{formatSpeed(down)}</span>
      </div>
      <div className="mt-1 flex items-center gap-2 text-xs">
        <span className="size-2 rounded-full" style={{ background: "#a78bfa" }} />
        <span className="text-canvas/70">{outgoing}</span>
        <span className="ml-auto font-mono text-canvas">{formatSpeed(up)}</span>
      </div>
    </div>
  );
}

export function TrafficChart() {
  const { history } = useMetrics();
  const { t } = useI18n();
  return (
    <Card>
      <CardHeader>
        <div>
          <CardTitle>{t("dashboard.traffic")}</CardTitle>
          <p className="mt-0.5 text-xs text-ink-tertiary">{t("dashboard.trafficSubtitle")}</p>
        </div>
        <div className="flex items-center gap-4 text-xs text-ink-secondary">
          <span className="flex items-center gap-2">
            <span className="size-2 rounded-full" style={{ background: "#22d3ee" }} />
            {t("dashboard.incoming")}
          </span>
          <span className="flex items-center gap-2">
            <span className="size-2 rounded-full" style={{ background: "#a78bfa" }} />
            {t("dashboard.outgoing")}
          </span>
        </div>
      </CardHeader>
      <div className="h-72 w-full">
        <ResponsiveContainer width="100%" height="100%">
          <AreaChart data={history} margin={{ left: -8, right: 8, top: 12, bottom: 0 }}>
            <defs>
              <linearGradient id="down-fill" x1="0" y1="0" x2="0" y2="1">
                <stop offset="0%" stopColor="#22d3ee" stopOpacity={0.32} />
                <stop offset="100%" stopColor="#22d3ee" stopOpacity={0} />
              </linearGradient>
              <linearGradient id="up-fill" x1="0" y1="0" x2="0" y2="1">
                <stop offset="0%" stopColor="#a78bfa" stopOpacity={0.28} />
                <stop offset="100%" stopColor="#a78bfa" stopOpacity={0} />
              </linearGradient>
            </defs>
            <CartesianGrid stroke="#262626" vertical={false} />
            <XAxis
              dataKey="t"
              tickFormatter={fmtTime}
              stroke="#6b7280"
              fontSize={11}
              tickLine={false}
              axisLine={false}
              interval="preserveStartEnd"
              minTickGap={48}
            />
            <YAxis
              stroke="#6b7280"
              fontSize={11}
              tickLine={false}
              axisLine={false}
              tickFormatter={fmtAxis}
              width={48}
            />
            <Tooltip
              content={<ChartTooltip incoming={t("dashboard.incoming")} outgoing={t("dashboard.outgoing")} />}
              cursor={{ stroke: "#ffffff44", strokeWidth: 1 }}
            />
            <Area
              type="monotone"
              dataKey="down"
              stroke="#22d3ee"
              strokeWidth={1.6}
              fill="url(#down-fill)"
              isAnimationActive={false}
              dot={false}
              activeDot={{ r: 4, fill: "#ffffff", stroke: "#22d3ee", strokeWidth: 1.5 }}
            />
            <Area
              type="monotone"
              dataKey="up"
              stroke="#a78bfa"
              strokeWidth={1.6}
              fill="url(#up-fill)"
              isAnimationActive={false}
              dot={false}
              activeDot={{ r: 4, fill: "#ffffff", stroke: "#a78bfa", strokeWidth: 1.5 }}
            />
          </AreaChart>
        </ResponsiveContainer>
      </div>
    </Card>
  );
}
