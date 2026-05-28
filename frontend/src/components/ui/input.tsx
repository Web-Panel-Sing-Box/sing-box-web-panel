
import { forwardRef, type InputHTMLAttributes, type ReactNode } from "react";

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
