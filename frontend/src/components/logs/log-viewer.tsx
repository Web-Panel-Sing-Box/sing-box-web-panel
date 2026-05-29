
import { useEffect, useRef } from "react";
import { AnimatePresence, m } from "framer-motion";

import { useLogs } from "@/lib/mock/store";
import { useLogFilter } from "@/hooks/useLogFilter";
import { formatTime } from "@/lib/format";
import { cn } from "@/lib/utils";
import { useI18n } from "@/lib/i18n";

import type { LogFilter } from "./log-filter-bar";

const levelColor: Record<"info" | "warn" | "error", string> = {
  info: "text-ink-secondary",
  warn: "text-amber",
  error: "text-danger"
};

export function LogViewer({ filter }: { filter: LogFilter }) {
  const logs = useLogs();
  const { t } = useI18n();
  const scrollRef = useRef<HTMLDivElement>(null);

  const filtered = useLogFilter(logs, filter);

  useEffect(() => {
    if (!scrollRef.current) return;
    scrollRef.current.scrollTop = scrollRef.current.scrollHeight;
  }, [filtered.length]);

  return (
    <div className="flex min-h-0 flex-1 overflow-hidden rounded-xl border border-subtle bg-canvas">
      <div
        ref={scrollRef}
        className="min-h-0 w-full flex-1 overflow-y-auto px-4 py-4 font-mono text-[12.5px] leading-relaxed"
      >
        <AnimatePresence initial={false}>
          {filtered.map((l) => (
            <m.div
              key={l.id}
              initial={{ opacity: 0 }}
              animate={{ opacity: 1 }}
              transition={{ duration: 0.05 }}
              className="flex gap-3 py-0.5"
            >
              <span className="shrink-0 text-ink-tertiary">{formatTime(l.t)}</span>
              <span className={cn("w-12 shrink-0 uppercase", levelColor[l.level])}>{l.level}</span>
              <span className={cn(l.level === "error" ? "text-danger" : "text-ink-primary/85")}>{l.message}</span>
            </m.div>
          ))}
        </AnimatePresence>
        {filtered.length === 0 ? (
          <div className="py-10 text-center text-ink-tertiary">{t("logs.noMatch")}</div>
        ) : null}
      </div>
    </div>
  );
}
