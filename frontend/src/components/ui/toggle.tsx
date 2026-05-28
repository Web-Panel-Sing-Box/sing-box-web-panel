
import { motion } from "framer-motion";

import { cn } from "@/lib/utils";

type ToggleProps = {
  checked: boolean;
  onChange: (next: boolean) => void;
  disabled?: boolean;
  label?: string;
  description?: string;
  className?: string;
};

export function Toggle({ checked, onChange, disabled, label, description, className }: ToggleProps) {
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
          "relative inline-flex h-5 w-9 shrink-0 items-center rounded-full transition-colors duration-200",
          checked ? "bg-brand" : "bg-[#3a3a3a]"
        )}
      >
        <motion.span
          className="absolute left-0.5 h-4 w-4 rounded-full bg-white shadow-sm"
          animate={{ x: checked ? 16 : 0 }}
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
