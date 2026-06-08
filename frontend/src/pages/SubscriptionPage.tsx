import { useEffect, useMemo, useState } from "react";
import { useParams } from "react-router-dom";
import { Check, Clock, Copy, Infinity as InfinityIcon, Zap } from "lucide-react";

import { getSubscriptionMeta, type SubscriptionMeta } from "@/api";
import { QrCode } from "@/components/ui/qr-code";
import { Progress } from "@/components/ui/progress";
import { formatBytes, formatDate } from "@/lib/format";

type LoadState =
  | { status: "loading" }
  | { status: "error"; message: string }
  | { status: "ready"; meta: SubscriptionMeta };

// Import deep links understood by popular clients, built from the sub URL.
const IMPORTERS: { label: string; href: (url: string) => string }[] = [
  { label: "v2rayNG", href: (u) => `v2rayng://install-config?url=${encodeURIComponent(u)}` },
  { label: "sing-box", href: (u) => `sing-box://import-remote-profile?url=${encodeURIComponent(u)}` },
  { label: "Streisand", href: (u) => `streisand://import/${u}` },
  { label: "Hiddify", href: (u) => `hiddify://import/${u}` },
  { label: "Clash", href: (u) => `clash://install-config?url=${encodeURIComponent(u)}` },
];

function daysLeft(expiry: string): { label: string; tone: string } | null {
  if (!expiry) return null;
  const diff = new Date(expiry).getTime() - Date.now();
  if (Number.isNaN(diff)) return null;
  if (diff <= 0) return { label: "Expired", tone: "text-danger" };
  const days = Math.floor(diff / 86_400_000);
  if (days >= 1) return { label: `${days}d left`, tone: days <= 3 ? "text-amber" : "text-ink-secondary" };
  const hours = Math.max(1, Math.floor(diff / 3_600_000));
  return { label: `${hours}h left`, tone: "text-amber" };
}

export function SubscriptionPage() {
  const { token = "" } = useParams();
  const [state, setState] = useState<LoadState>({ status: "loading" });

  useEffect(() => {
    let cancelled = false;
    getSubscriptionMeta(token)
      .then((meta) => {
        if (!cancelled) setState({ status: "ready", meta });
      })
      .catch(() => {
        if (!cancelled) setState({ status: "error", message: "Subscription not found or disabled." });
      });
    return () => {
      cancelled = true;
    };
  }, [token]);

  return (
    <div className="min-h-screen w-full bg-canvas px-4 py-10 text-ink-primary">
      <div className="mx-auto w-full max-w-[460px]">
        {state.status === "loading" ? (
          <div className="rounded-2xl border border-subtle bg-surface p-8 text-center text-sm text-ink-tertiary">
            Loading…
          </div>
        ) : null}
        {state.status === "error" ? (
          <div className="rounded-2xl border border-subtle bg-surface p-8 text-center">
            <img src="/logo.png" alt="Shilka" className="mx-auto mb-4 size-12 rounded-xl" />
            <p className="text-sm text-danger">{state.message}</p>
          </div>
        ) : null}
        {state.status === "ready" ? <Ready meta={state.meta} /> : null}
      </div>
    </div>
  );
}

function Ready({ meta }: { meta: SubscriptionMeta }) {
  const absoluteSub = useMemo(() => {
    if (/^https?:\/\//i.test(meta.subscriptionUrl)) return meta.subscriptionUrl;
    return `${window.location.origin}${meta.subscriptionUrl}`;
  }, [meta.subscriptionUrl]);

  const unlimited = meta.total <= 0;
  const pct = unlimited ? 0 : Math.min(100, Math.max(0, (meta.used / meta.total) * 100));
  const expiry = daysLeft(meta.expiry);

  return (
    <div className="space-y-4">
      <header className="flex items-center gap-3">
        <img src="/logo.png" alt="Shilka" className="size-11 rounded-xl" />
        <div className="min-w-0">
          <div className="text-xs uppercase tracking-wider text-ink-tertiary">Shilka</div>
          <div className="truncate text-lg font-semibold">{meta.name}</div>
        </div>
        <span className="ml-auto inline-flex items-center gap-1.5 rounded-full border border-subtle bg-surface px-2.5 py-1 text-[11px]">
          <span className={`size-1.5 rounded-full ${meta.online ? "bg-success" : "bg-ink-tertiary"}`} />
          {meta.online ? "Online" : "Offline"}
        </span>
      </header>

      <section className="rounded-2xl border border-subtle bg-surface p-5">
        <div className="flex items-baseline justify-between">
          <div className="font-mono text-2xl font-semibold">
            {formatBytes(meta.used)}
            <span className="mx-1.5 text-ink-tertiary">/</span>
            <span className="text-ink-secondary">
              {unlimited ? <InfinityIcon className="inline size-5" /> : formatBytes(meta.total)}
            </span>
          </div>
          <div className="flex items-center gap-2 text-xs">
            {unlimited ? (
              <span className="inline-flex items-center gap-1 rounded-full bg-violet/15 px-2 py-1 text-violet">
                <Zap size={12} /> Unlimited
              </span>
            ) : null}
            {expiry ? (
              <span className={`inline-flex items-center gap-1 rounded-full bg-white/5 px-2 py-1 ${expiry.tone}`}>
                <Clock size={12} /> {expiry.label}
              </span>
            ) : null}
          </div>
        </div>
        {!unlimited ? (
          <div className="mt-3">
            <Progress value={pct} height={8} />
            <div className="mt-1 text-right font-mono text-[11px] text-ink-tertiary">{pct.toFixed(1)}%</div>
          </div>
        ) : null}
        <div className="mt-3 text-xs text-ink-tertiary">
          {meta.expiry ? `Expires ${formatDate(meta.expiry)}` : "No expiry date"}
        </div>
      </section>

      <section className="flex flex-col items-center gap-4 rounded-2xl border border-subtle bg-surface p-5">
        <QrCode payload={absoluteSub} size={196} />
        <CopyRow value={absoluteSub} />
      </section>

      <section className="rounded-2xl border border-subtle bg-surface p-5">
        <div className="mb-3 text-xs uppercase tracking-wider text-ink-tertiary">Import to app</div>
        <div className="grid grid-cols-2 gap-2 sm:grid-cols-3">
          {IMPORTERS.map((imp) => (
            <a
              key={imp.label}
              href={imp.href(absoluteSub)}
              className="rounded-lg border border-subtle bg-canvas px-3 py-2 text-center text-sm text-ink-secondary transition-colors duration-150 hover:border-white/20 hover:text-ink-primary"
            >
              {imp.label}
            </a>
          ))}
        </div>
      </section>

      {meta.links.length > 0 ? (
        <section className="rounded-2xl border border-subtle bg-surface p-5">
          <div className="mb-3 text-xs uppercase tracking-wider text-ink-tertiary">Configs</div>
          <div className="space-y-2">
            {meta.links.map((link) => (
              <div key={link.url} className="rounded-lg border border-subtle bg-canvas p-3">
                <div className="mb-1 flex items-center gap-2 text-xs text-ink-tertiary">
                  <span className="rounded-full bg-white/5 px-2 py-0.5 font-mono uppercase">{link.protocol}</span>
                  <span className="truncate">{link.label}</span>
                </div>
                <CopyRow value={link.url} />
              </div>
            ))}
          </div>
        </section>
      ) : null}
    </div>
  );
}

function CopyRow({ value }: { value: string }) {
  const [copied, setCopied] = useState(false);
  const copy = async () => {
    try {
      await navigator.clipboard.writeText(value);
      setCopied(true);
      window.setTimeout(() => setCopied(false), 1500);
    } catch {
      // clipboard unavailable
    }
  };
  return (
    <div className="flex w-full items-center gap-2">
      <code className="min-w-0 flex-1 truncate rounded-md bg-canvas px-2 py-1.5 font-mono text-[11px] text-ink-secondary">
        {value}
      </code>
      <button
        type="button"
        onClick={copy}
        className="grid size-8 shrink-0 place-items-center rounded-md border border-subtle text-ink-secondary transition-colors duration-150 hover:bg-hover hover:text-ink-primary"
        aria-label="Copy"
      >
        {copied ? <Check size={14} className="text-success" /> : <Copy size={14} />}
      </button>
    </div>
  );
}
