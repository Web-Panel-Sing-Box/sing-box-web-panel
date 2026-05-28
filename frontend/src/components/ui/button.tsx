
import { forwardRef, type ReactNode } from "react";
import { motion, type HTMLMotionProps } from "framer-motion";
import { Loader2 } from "lucide-react";

import { cn } from "@/lib/utils";

type Variant = "primary" | "secondary" | "ghost" | "danger" | "white";
type Size = "sm" | "md";

const base =
  "inline-flex items-center justify-center gap-2 rounded-lg font-medium transition-colors duration-200 disabled:opacity-50 disabled:cursor-not-allowed select-none whitespace-nowrap";

const sizeMap: Record<Size, string> = {
  sm: "h-8 px-3 text-xs",
  md: "h-10 px-4 text-sm"
};

const variantMap: Record<Variant, string> = {
  primary: "bg-brand text-white hover:bg-[#0e8e6e]",
  secondary: "bg-transparent border border-subtle text-ink-primary hover:bg-hover",
  ghost: "bg-transparent text-ink-secondary hover:bg-hover hover:text-ink-primary",
  danger: "bg-transparent text-ink-secondary hover:text-danger",
  white: "bg-white text-canvas hover:bg-white/90"
};

type ButtonProps = Omit<HTMLMotionProps<"button">, "children"> & {
  variant?: Variant;
  size?: Size;
  loading?: boolean;
  children?: ReactNode;
};

export const Button = forwardRef<HTMLButtonElement, ButtonProps>(function Button(
  { variant = "secondary", size = "md", loading, className, children, disabled, ...props },
  ref
) {
  return (
    <motion.button
      ref={ref}
      whileTap={{ scale: 0.98 }}
      transition={{ type: "tween", duration: 0.08, ease: "easeOut" }}
      className={cn(base, sizeMap[size], variantMap[variant], className)}
      disabled={disabled || loading}
      {...props}
    >
      {loading ? <Loader2 className="size-4 animate-spin" /> : null}
      {children}
    </motion.button>
  );
});
