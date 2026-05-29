
import { m } from "framer-motion";

import { cn } from "@/lib/utils";

type Option<T extends string> = { value: T; label: string };

type SegmentedProps<T extends string> = {
  value: T;
  onChange: (v: T) => void;
  options: Option<T>[];
  layoutId?: string;
  className?: string;
};

export function Segmented<T extends string>({
  value,
  onChange,
  options,
  layoutId = "segmented-active",
  className
}: SegmentedProps<T>) {
  return (
    <div className={cn("inline-flex rounded-full bg-canvas p-1", className)}>
      {options.map((opt) => {
        const active = opt.value === value;
        return (
          <button
            key={opt.value}
            type="button"
            onClick={() => onChange(opt.value)}
            className={cn(
              "relative isolate h-8 rounded-full px-4 text-xs font-medium transition-colors duration-150",
              active ? "text-ink-primary" : "text-ink-secondary hover:text-ink-primary"
            )}
          >
            {active ? (
              <m.span
                layoutId={layoutId}
                className="absolute inset-0 -z-10 rounded-full bg-elevated shadow-card"
                transition={{ type: "spring", stiffness: 500, damping: 38 }}
              />
            ) : null}
            <span>{opt.label}</span>
          </button>
        );
      })}
    </div>
  );
}
