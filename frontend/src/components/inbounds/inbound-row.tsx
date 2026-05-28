
import { useState } from "react";
import { AnimatePresence, motion } from "framer-motion";
import { ChevronDown, Copy, Pencil, QrCode, Trash2 } from "lucide-react";

import { Toggle } from "@/components/ui/toggle";
import { accordionVariants } from "@/lib/motion";
import { cn } from "@/lib/utils";
import type { Inbound } from "@/lib/mock/inbounds";
import { useStore } from "@/lib/mock/store";
import { useToast } from "@/components/ui/toast";

export function InboundRow({ inbound }: { inbound: Inbound }) {
  const [open, setOpen] = useState(false);
  const { toggleInbound, removeInbound, cloneInbound } = useStore();
  const { push } = useToast();

  return (
    <div className={cn("group border-b border-subtle last:border-b-0", open && "bg-canvas/40")}>
      <button
        type="button"
        onClick={() => setOpen((v) => !v)}
        className="grid w-full grid-cols-[110px_110px_1fr_90px_auto_auto_24px] items-center gap-3 px-4 py-3 text-left transition-colors duration-200 hover:bg-hover sm:px-5"
      >
        <ProtocolChip protocol={inbound.protocol} />
        <span className="font-mono text-sm text-ink-secondary">:{inbound.port}</span>
        <span className="truncate text-sm text-ink-primary">{inbound.remark}</span>
        <span className="hidden text-xs text-ink-tertiary md:inline">{inbound.clientCount} clients</span>
        <Toggle
          checked={inbound.enabled}
          onChange={(v) => {
            toggleInbound(inbound.id);
            push(`${inbound.remark} ${v ? "enabled" : "disabled"}`);
          }}
        />
        <span className={cn("text-xs", inbound.enabled ? "text-success" : "text-ink-tertiary")}>
          {inbound.enabled ? "Active" : "Disabled"}
        </span>
        <ChevronDown
          size={16}
          className={cn("text-ink-tertiary transition-transform duration-200", open && "rotate-180")}
        />
      </button>

      <AnimatePresence initial={false}>
        {open ? (
          <motion.div variants={accordionVariants} initial="initial" animate="animate" exit="exit" className="overflow-hidden">
            <div className="flex flex-wrap items-center gap-2 border-t border-subtle px-4 py-3 sm:px-5">
              <RowAction icon={<Pencil size={14} />} label="Edit" onClick={() => push("Editing is mocked")} />
              <RowAction icon={<Copy size={14} />} label="Clone" onClick={() => { cloneInbound(inbound.id); push(`Cloned ${inbound.remark}`); }} />
              <RowAction icon={<QrCode size={14} />} label="QR code" onClick={() => push("QR is shown on the client level")} />
              <RowAction
                icon={<Trash2 size={14} />}
                label="Delete"
                danger
                onClick={() => { removeInbound(inbound.id); push(`Deleted ${inbound.remark}`, "error"); }}
              />
              <span className="ml-auto font-mono text-[11px] text-ink-tertiary">
                {inbound.transmission.toUpperCase()} · {inbound.tls.toUpperCase()}
                {inbound.sni ? ` · ${inbound.sni}` : ""}
              </span>
            </div>
          </motion.div>
        ) : null}
      </AnimatePresence>
    </div>
  );
}

function RowAction({
  icon,
  label,
  onClick,
  danger
}: {
  icon: React.ReactNode;
  label: string;
  onClick: () => void;
  danger?: boolean;
}) {
  return (
    <button
      type="button"
      onClick={onClick}
      className={cn(
        "inline-flex h-8 items-center gap-1.5 rounded-full border border-subtle bg-canvas px-3 text-xs text-ink-secondary transition-colors duration-200",
        danger ? "hover:border-danger/40 hover:text-danger" : "hover:border-white/20 hover:text-ink-primary"
      )}
    >
      {icon}
      {label}
    </button>
  );
}

function ProtocolChip({ protocol }: { protocol: Inbound["protocol"] }) {
  const palette: Record<Inbound["protocol"], string> = {
    vless: "text-cyan border-cyan/30",
    naive: "text-amber border-amber/30",
    hysteria2: "text-violet border-violet/30"
  };
  return (
    <span
      className={cn(
        "inline-flex h-7 w-fit items-center rounded-full border bg-canvas px-2.5 font-mono text-[11px] uppercase tracking-wider",
        palette[protocol]
      )}
    >
      {protocol}
    </span>
  );
}
