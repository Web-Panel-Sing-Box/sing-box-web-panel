
import { useEffect, type ReactNode } from "react";
import { createPortal } from "react-dom";
import { AnimatePresence, motion } from "framer-motion";
import { X } from "lucide-react";

import { backdropVariants, modalVariants } from "@/lib/motion";
import { cn } from "@/lib/utils";

type ModalProps = {
  open: boolean;
  onClose: () => void;
  children: ReactNode;
  className?: string;
  width?: string;
};

export function Modal({ open, onClose, children, className, width = "max-w-[720px]" }: ModalProps) {
  useEffect(() => {
    if (!open) return;
    const prev = document.body.style.overflow;
    document.body.style.overflow = "hidden";
    function onKey(e: KeyboardEvent) {
      if (e.key === "Escape") onClose();
    }
    window.addEventListener("keydown", onKey);
    return () => {
      document.body.style.overflow = prev;
      window.removeEventListener("keydown", onKey);
    };
  }, [open, onClose]);

  if (typeof document === "undefined") return null;

  return createPortal(
    <AnimatePresence>
      {open ? (
        <motion.div
          variants={backdropVariants}
          initial="initial"
          animate="animate"
          exit="exit"
          className="fixed inset-0 z-50 flex items-start justify-center bg-black/70 px-4 py-10 backdrop-blur-[8px] sm:items-center"
          onClick={onClose}
        >
          <motion.div
            variants={modalVariants}
            initial="initial"
            animate="animate"
            exit="exit"
            onClick={(e) => e.stopPropagation()}
            className={cn(
              "relative w-full overflow-hidden rounded-2xl border border-subtle bg-elevated shadow-pop",
              width,
              className
            )}
          >
            {children}
          </motion.div>
        </motion.div>
      ) : null}
    </AnimatePresence>,
    document.body
  );
}

type ModalHeaderProps = {
  title: string;
  subtitle?: string;
  onClose: () => void;
};

export function ModalHeader({ title, subtitle, onClose }: ModalHeaderProps) {
  return (
    <div className="flex items-start justify-between gap-4 px-6 pb-2 pt-5">
      <div>
        <h2 className="text-base font-semibold text-ink-primary">{title}</h2>
        {subtitle ? <p className="mt-0.5 text-xs text-ink-tertiary">{subtitle}</p> : null}
      </div>
      <button
        type="button"
        onClick={onClose}
        className="-mr-2 rounded-md p-2 text-ink-secondary transition-colors duration-150 hover:bg-hover hover:text-ink-primary"
        aria-label="Close"
      >
        <X size={16} />
      </button>
    </div>
  );
}

export function ModalBody({ children, className }: { children: ReactNode; className?: string }) {
  return (
    <div className={cn("max-h-[60vh] overflow-y-auto px-6 py-5", className)}>{children}</div>
  );
}

export function ModalFooter({
  children,
  className
}: {
  children: ReactNode;
  /** @deprecated kept for backwards compatibility; ignored — footer is now borderless */
  accent?: "violet" | "cyan";
  className?: string;
}) {
  return (
    <div
      className={cn(
        "flex items-center justify-end gap-2 bg-elevated px-6 pb-5 pt-2",
        className
      )}
    >
      {children}
    </div>
  );
}
