
import { cn } from "@/lib/utils";

type ProgressProps = {
  value: number;
  max?: number;
  className?: string;
  height?: number;
  /** Color overrides */
  trackClass?: string;
  fillClass?: string;
};

export function Progress({
  value,
  max = 100,
  className,
  height = 6,
  trackClass,
  fillClass
}: ProgressProps) {
  const ratio = Math.min(1, Math.max(0, value / max));
  const critical = ratio > 0.9;
  return (
    <div
      className={cn("relative w-full overflow-hidden rounded-full bg-white/10", trackClass, className)}
      style={{ height }}
    >
      <div
        className={cn(
          "h-full rounded-full transition-[width] duration-500 ease-out",
          critical ? "bg-danger" : "bg-white/80",
          fillClass
        )}
        style={{ width: `${ratio * 100}%` }}
      />
    </div>
  );
}
