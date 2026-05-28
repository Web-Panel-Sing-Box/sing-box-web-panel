
import { AnimatePresence, motion } from "framer-motion";

import { ToastProvider, useToast } from "./toast";
import { cn } from "@/lib/utils";

export function Toaster({ children }: { children: React.ReactNode }) {
  return (
    <ToastProvider>
      {children}
      <ToasterViewport />
    </ToastProvider>
  );
}

function ToasterViewport() {
  const { toasts } = useToast();
  return (
    <div className="pointer-events-none fixed inset-x-0 top-4 z-[100] flex flex-col items-center gap-2 px-4">
      <AnimatePresence>
        {toasts.map((t) => (
          <motion.div
            key={t.id}
            initial={{ opacity: 0, y: -16, scale: 0.96 }}
            animate={{ opacity: 1, y: 0, scale: 1 }}
            exit={{ opacity: 0, y: -8, scale: 0.98 }}
            transition={{ duration: 0.2, ease: "easeOut" }}
            className={cn(
              "pointer-events-auto flex w-full max-w-md items-start gap-3 rounded-xl border border-subtle bg-elevated px-4 py-3 text-sm text-ink-primary shadow-pop",
              t.variant === "success" && "border-l-4 border-l-success",
              t.variant === "error" && "border-l-4 border-l-danger"
            )}
          >
            <span className="flex-1">
              {t.message}
              {t.variant === "success" ? <span className="ml-2 font-medium text-success">✓</span> : null}
            </span>
          </motion.div>
        ))}
      </AnimatePresence>
    </div>
  );
}
