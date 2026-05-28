
import { useEffect, useMemo, useRef } from "react";
import { AnimatePresence, motion } from "framer-motion";

import { useStore } from "@/lib/mock/store";
import { formatTime } from "@/lib/format";
import { cn } from "@/lib/utils";

import type { LogFilter } from "./log-filter-bar";

const levelColor: Record<"info" | "warn" | "error", string> = {
  info: "text-ink-secondary",
  warn: "text-amber",
  error: "text-danger"
};

export function LogViewer({ filter }: { filter: LogFilter }) {
  const { logs } = useStore();
  const scrollRef = useRef<HTMLDivElement>(null);

  const filtered = useMemo(() => {
    const q = filter.query.trim().toLowerCase();
    return logs.filter((l) => {
      if (filter.level !== "all" && l.level !== filter.level) return false;
      if (!q) return true;
      return l.message.toLowerCase().includes(q);
    });
  }, [logs, filter]);

  useEffect(() => {
    if (!scrollRef.current) return;
    scrollRef.current.scrollTop = scrollRef.current.scrollHeight;
  }, [filtered.length]);

  return (
    <div className="overflow-hidden rounded-xl border border-subtle bg-canvas">
      <div
        ref={scrollRef}
        className="h-[calc(100vh-260px)] min-h-[420px] overflow-y-auto px-4 py-4 font-mono text-[12.5px] leading-relaxed"
      >
        <AnimatePresence initial={false}>
          {filtered.map((l) => (
            <motion.div
              key={l.id}
              initial={{ opacity: 0 }}
              animate={{ opacity: 1 }}
              transition={{ duration: 0.05 }}
              className="flex gap-3 py-0.5"
            >
              <span className="shrink-0 text-ink-tertiary">{formatTime(l.t)}</span>
              <span className={cn("w-12 shrink-0 uppercase", levelColor[l.level])}>{l.level}</span>
              <span className={cn(l.level === "error" ? "text-danger" : "text-ink-primary/85")}>{l.message}</span>
            </motion.div>
          ))}
        </AnimatePresence>
        {filtered.length === 0 ? (
          <div className="py-10 text-center text-ink-tertiary">No log lines match the current filter.</div>
        ) : null}
      </div>
    </div>
  );
}
