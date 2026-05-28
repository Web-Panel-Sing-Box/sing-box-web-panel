import { GlassStrip } from "@/components/dashboard/glass-strip";
import {
  CpuCard,
  DiskCard,
  RamCard,
  TrafficSplitCard
} from "@/components/dashboard/metric-card";
import { ClientsTelemetryCard, InboundsActiveCard } from "@/components/dashboard/connections-panel";
import { TrafficChart } from "@/components/dashboard/traffic-chart";
import { QuickLinks } from "@/components/dashboard/quick-links";

export function DashboardPage() {
  return (
    <div className="mx-auto flex max-w-[1440px] flex-col gap-6">
      <GlassStrip />

      <section className="grid grid-cols-1 gap-4 md:grid-cols-2 xl:grid-cols-4">
        <CpuCard />
        <RamCard />
        <DiskCard />
        <TrafficSplitCard />
      </section>

      <section className="grid grid-cols-1 gap-4 md:grid-cols-2">
        <InboundsActiveCard />
        <ClientsTelemetryCard />
      </section>

      <TrafficChart />

      <QuickLinks />
    </div>
  );
}
