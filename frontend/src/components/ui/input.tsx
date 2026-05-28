import { forwardRef, useRef, type InputHTMLAttributes, type ReactNode } from "react";
import { Calendar } from "lucide-react";

import { cn } from "@/lib/utils";

type InputProps = InputHTMLAttributes<HTMLInputElement> & {
  accent?: "white" | "brand";
  error?: string;
  trailing?: ReactNode;
  mono?: boolean;
};

export const Input = forwardRef<HTMLInputElement, InputProps>(function Input(
  { className, accent = "white", error, trailing, mono, ...props },
  ref
) {
  return (
    <div className="space-y-1">
      <div
        className={cn(
          "group flex h-10 items-center gap-2 rounded-lg border bg-elevated px-3 transition-colors duration-150",
          "border-white/10",
          accent === "brand"
            ? "focus-within:border-brand"
            : "focus-within:border-white",
          error && "border-danger focus-within:border-danger"
        )}
      >
        <input
          ref={ref}
          className={cn(
            "h-full min-w-0 flex-1 bg-transparent text-sm text-ink-primary placeholder:text-ink-tertiary outline-none",
            mono && "font-mono",
            className
          )}
          {...props}
        />
        {trailing ? <div className="flex shrink-0 items-center text-ink-secondary">{trailing}</div> : null}
      </div>
      {error ? <p className="text-xs text-danger">{error}</p> : null}
    </div>
  );
});

export function Label({ children, htmlFor, hint }: { children: ReactNode; htmlFor?: string; hint?: string }) {
  return (
    <label htmlFor={htmlFor} className="mb-1.5 flex items-baseline justify-between text-xs text-ink-secondary">
      <span>{children}</span>
      {hint ? <span className="text-ink-tertiary">{hint}</span> : null}
    </label>
  );
}

type TextareaProps = React.TextareaHTMLAttributes<HTMLTextAreaElement> & { mono?: boolean };

export const Textarea = forwardRef<HTMLTextAreaElement, TextareaProps>(function Textarea(
  { className, mono, ...props },
  ref
) {
  return (
    <textarea
      ref={ref}
      className={cn(
        "block w-full resize-none rounded-lg border border-white/10 bg-elevated px-3 py-2 text-sm text-ink-primary placeholder:text-ink-tertiary outline-none transition-colors duration-150 focus:border-white",
        mono && "font-mono",
        className
      )}
      {...props}
    />
  );
});

type NumberInputProps = {
  value: string;
  onChange: (next: string) => void;
  min?: number;
  max?: number;
  placeholder?: string;
  mono?: boolean;
  className?: string;
  trailing?: ReactNode;
};

/**
 * Numeric input that stores its value as a string and never auto-prepends a leading "0".
 * Empty input is preserved as "" so users can clear and retype without seeing a stuck "0".
 */
export function NumberInput({ value, onChange, min, max, placeholder, mono, className, trailing }: NumberInputProps) {
  function handleChange(e: React.ChangeEvent<HTMLInputElement>) {
    const raw = e.target.value.replace(/[^0-9]/g, "");
    // Strip leading zeros (but keep a single "0" if that's all the user typed).
    const stripped = raw.replace(/^0+(?=\d)/, "");
    if (stripped === "") {
      onChange("");
      return;
    }
    if (min !== undefined && Number(stripped) < min) {
      onChange(String(min));
      return;
    }
    if (max !== undefined && Number(stripped) > max) {
      onChange(String(max));
      return;
    }
    onChange(stripped);
  }
  return (
    <Input
      type="text"
      inputMode="numeric"
      value={value}
      onChange={handleChange}
      placeholder={placeholder}
      mono={mono}
      className={className}
      trailing={trailing}
    />
  );
}

type DateInputProps = {
  value: string;
  onChange: (next: string) => void;
  className?: string;
};

/**
 * Date input that hides the default tiny calendar glyph and shows our own,
 * tappable anywhere — clicking the whole shell opens the native date picker.
 */
export function DateInput({ value, onChange, className }: DateInputProps) {
  const ref = useRef<HTMLInputElement>(null);
  function openPicker() {
    const el = ref.current;
    if (!el) return;
    try {
      // showPicker exists on Chromium-based browsers and Safari TP.
      (el as HTMLInputElement & { showPicker?: () => void }).showPicker?.();
    } catch {
      el.focus();
    }
  }
  return (
    <div
      onClick={openPicker}
      className={cn(
        "group flex h-10 cursor-pointer items-center gap-2 rounded-lg border border-white/10 bg-elevated px-3 transition-colors duration-150 hover:border-white/20 focus-within:border-white",
        className
      )}
    >
      <input
        ref={ref}
        type="date"
        value={value}
        onChange={(e) => onChange(e.target.value)}
        className="h-full min-w-0 flex-1 cursor-pointer bg-transparent text-sm text-ink-primary outline-none [&::-webkit-calendar-picker-indicator]:hidden [&::-webkit-inner-spin-button]:hidden"
      />
      <Calendar size={14} className="shrink-0 text-ink-secondary transition-colors duration-150 group-hover:text-ink-primary" />
    </div>
  );
}
