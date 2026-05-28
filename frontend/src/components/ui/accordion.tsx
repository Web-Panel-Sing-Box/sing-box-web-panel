
import { useState, type ReactNode } from "react";
import { AnimatePresence, motion } from "framer-motion";
import { ChevronRight } from "lucide-react";

import { accordionVariants } from "@/lib/motion";
import { cn } from "@/lib/utils";

type AccordionProps = {
  title: ReactNode;
  description?: ReactNode;
  defaultOpen?: boolean;
  children: ReactNode;
  className?: string;
};

export function Accordion({ title, description, defaultOpen = true, children, className }: AccordionProps) {
  const [open, setOpen] = useState(defaultOpen);
  return (
    <div className={cn("rounded-xl border border-subtle bg-surface", className)}>
      <button
        type="button"
        onClick={() => setOpen((v) => !v)}
        className="flex w-full items-center justify-between gap-3 px-4 py-3 text-left transition-colors duration-150 hover:bg-hover"
      >
        <div className="min-w-0">
          <div className="text-sm font-medium text-ink-primary">{title}</div>
          {description ? <div className="mt-0.5 text-xs text-ink-tertiary">{description}</div> : null}
        </div>
        <ChevronRight
          size={16}
          className={cn(
            "shrink-0 text-ink-secondary transition-transform duration-200",
            open && "rotate-90"
          )}
        />
      </button>
      <AnimatePresence initial={false}>
        {open ? (
          <motion.div
            initial="initial"
            animate="animate"
            exit="exit"
            variants={accordionVariants}
            className="overflow-hidden"
          >
            <div className="border-t border-subtle px-4 py-4">{children}</div>
          </motion.div>
        ) : null}
      </AnimatePresence>
    </div>
  );
}
