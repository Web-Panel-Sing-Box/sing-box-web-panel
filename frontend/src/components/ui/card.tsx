import { forwardRef, type HTMLAttributes } from "react";

import { cn } from "@/lib/utils";

type CardProps = HTMLAttributes<HTMLDivElement> & {
  elevated?: boolean;
  padded?: boolean;
};

export const Card = forwardRef<HTMLDivElement, CardProps>(function Card(
  { className, elevated, padded = true, ...props },
  ref
) {
  return (
    <div
      ref={ref}
      className={cn(
        "rounded-xl border border-subtle shadow-card",
        elevated ? "bg-elevated" : "bg-surface",
        padded && "p-5",
        className
      )}
      {...props}
    />
  );
});

export function CardHeader({ className, ...props }: HTMLAttributes<HTMLDivElement>) {
  return <div className={cn("mb-4 flex items-center justify-between gap-3", className)} {...props} />;
}

export function CardTitle({ className, ...props }: HTMLAttributes<HTMLHeadingElement>) {
  return <h3 className={cn("text-sm font-medium text-ink-primary", className)} {...props} />;
}

export function CardLabel({ className, ...props }: HTMLAttributes<HTMLSpanElement>) {
  return <span className={cn("text-xs text-ink-secondary", className)} {...props} />;
}
