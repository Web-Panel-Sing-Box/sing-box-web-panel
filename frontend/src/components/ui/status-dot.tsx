import { cn } from "@/lib/utils";

type StatusDotProps = {
  state: "online" | "stopped" | "warning" | "neutral";
  size?: number;
  className?: string;
};

const colors: Record<StatusDotProps["state"], string> = {
  online: "bg-success shadow-[0_0_0_3px_rgba(25,195,125,0.18)]",
  stopped: "bg-danger shadow-[0_0_0_3px_rgba(239,68,68,0.18)]",
  warning: "bg-amber shadow-[0_0_0_3px_rgba(250,204,21,0.18)]",
  neutral: "bg-ink-tertiary"
};

export function StatusDot({ state, size = 8, className }: StatusDotProps) {
  return (
    <span
      className={cn("inline-block rounded-full", colors[state], className)}
      style={{ width: size, height: size }}
    />
  );
}
