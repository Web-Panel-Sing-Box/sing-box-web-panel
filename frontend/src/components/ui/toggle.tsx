import { m } from "framer-motion";

import { cn } from "@/lib/utils";

type ToggleSize = "sm" | "md" | "lg";

type ToggleProps = {
  checked: boolean;
  onChange: (next: boolean) => void;
  disabled?: boolean;
  label?: string;
  description?: string;
  size?: ToggleSize;
  className?: string;
};

const trackSize: Record<ToggleSize, string> = {
  sm: "h-5 w-9",
  md: "h-6 w-11",
  lg: "h-7 w-12"
};

const knobSize: Record<ToggleSize, string> = {
  sm: "h-4 w-4",
  md: "h-5 w-5",
  lg: "h-6 w-6"
};

// Horizontal travel in pixels — track width minus knob width minus 2× side inset.
const knobTravel: Record<ToggleSize, number> = {
  sm: 16,
  md: 20,
  lg: 20
};

export function Toggle({
  checked,
  onChange,
  disabled,
  label,
  description,
  size = "sm",
  className
}: ToggleProps) {
  return (
    <button
      type="button"
      role="switch"
      aria-checked={checked}
      disabled={disabled}
      onClick={() => onChange(!checked)}
      className={cn(
        "group flex items-center gap-3 text-left",
        disabled && "cursor-not-allowed opacity-50",
        className
      )}
    >
      <span
        className={cn(
          "relative inline-flex shrink-0 items-center rounded-full transition-colors duration-200",
          trackSize[size],
          checked ? "bg-brand" : "bg-[#3a3a3a]"
        )}
      >
        <m.span
          className={cn("absolute left-0.5 rounded-full bg-white shadow-sm", knobSize[size])}
          animate={{ x: checked ? knobTravel[size] : 0 }}
          transition={{ type: "spring", stiffness: 700, damping: 40 }}
        />
      </span>
      {label ? (
        <span className="flex flex-col">
          <span className="text-sm text-ink-primary">{label}</span>
          {description ? <span className="text-xs text-ink-tertiary">{description}</span> : null}
        </span>
      ) : null}
    </button>
  );
}
