
import { useEffect, useRef, useState } from "react";
import { AnimatePresence, motion } from "framer-motion";
import { Check, ChevronDown } from "lucide-react";

import { dropdownVariants } from "@/lib/motion";
import { cn } from "@/lib/utils";

export type SelectOption<T extends string = string> = {
  value: T;
  label: string;
  description?: string;
};

type SelectProps<T extends string> = {
  value: T;
  options: SelectOption<T>[];
  onChange: (value: T) => void;
  placeholder?: string;
  className?: string;
  disabled?: boolean;
};

export function Select<T extends string>({
  value,
  options,
  onChange,
  placeholder = "Select…",
  className,
  disabled
}: SelectProps<T>) {
  const [open, setOpen] = useState(false);
  const ref = useRef<HTMLDivElement>(null);

  useEffect(() => {
    if (!open) return;
    function onClick(e: MouseEvent) {
      if (!ref.current?.contains(e.target as Node)) setOpen(false);
    }
    function onKey(e: KeyboardEvent) {
      if (e.key === "Escape") setOpen(false);
    }
    window.addEventListener("mousedown", onClick);
    window.addEventListener("keydown", onKey);
    return () => {
      window.removeEventListener("mousedown", onClick);
      window.removeEventListener("keydown", onKey);
    };
  }, [open]);

  const selected = options.find((o) => o.value === value);

  return (
    <div ref={ref} className={cn("relative", className)}>
      <button
        type="button"
        disabled={disabled}
        onClick={() => setOpen((v) => !v)}
        className={cn(
          "flex h-10 w-full items-center justify-between gap-2 rounded-lg border border-white/10 bg-elevated px-3 text-left text-sm text-ink-primary transition-colors duration-150",
          open ? "border-white" : "hover:border-white/20",
          disabled && "cursor-not-allowed opacity-50"
        )}
      >
        <span className={cn("truncate", !selected && "text-ink-tertiary")}>
          {selected?.label ?? placeholder}
        </span>
        <ChevronDown
          size={16}
          className={cn("shrink-0 text-ink-secondary transition-transform duration-150", open && "rotate-180")}
        />
      </button>

      <AnimatePresence>
        {open ? (
          <motion.div
            initial="initial"
            animate="animate"
            exit="exit"
            variants={dropdownVariants}
            className="absolute left-0 right-0 top-[calc(100%+6px)] z-30 overflow-hidden rounded-lg border border-subtle bg-elevated shadow-pop"
          >
            <ul className="max-h-72 overflow-y-auto py-1">
              {options.map((opt) => {
                const active = opt.value === value;
                return (
                  <li key={opt.value}>
                    <button
                      type="button"
                      onClick={() => {
                        onChange(opt.value);
                        setOpen(false);
                      }}
                      className={cn(
                        "flex w-full items-start justify-between gap-3 px-3 py-2 text-left text-sm transition-colors duration-150 hover:bg-hover",
                        active ? "text-ink-primary" : "text-ink-secondary"
                      )}
                    >
                      <div className="min-w-0 flex-1">
                        <div className="truncate text-ink-primary">{opt.label}</div>
                        {opt.description ? (
                          <div className="truncate text-xs text-ink-tertiary">{opt.description}</div>
                        ) : null}
                      </div>
                      {active ? <Check size={14} className="mt-1 text-brand" /> : null}
                    </button>
                  </li>
                );
              })}
            </ul>
          </motion.div>
        ) : null}
      </AnimatePresence>
    </div>
  );
}
