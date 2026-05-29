
import { useState } from "react";

import { useMetrics, useStoreActions } from "@/lib/mock/store";
import { StatusDot } from "@/components/ui/status-dot";
import { formatBytes, formatUptime } from "@/lib/format";
import { Modal, ModalBody, ModalFooter, ModalHeader } from "@/components/ui/modal";
import { Button } from "@/components/ui/button";
import { useToast } from "@/components/ui/toast";
import { useI18n } from "@/lib/i18n";

export function GlassStrip() {
  const { metrics } = useMetrics();
  const { setCoreRunning } = useStoreActions();
  const { push } = useToast();
  const { t } = useI18n();
  const [confirmStop, setConfirmStop] = useState(false);
  const running = metrics.coreRunning;

  function handleCoreClick() {
    if (running) {
      setConfirmStop(true);
      return;
    }
    setCoreRunning(true);
    push(t("core.started"), "success");
  }

  function stopCore() {
    setCoreRunning(false);
    setConfirmStop(false);
    push(t("core.stopped"), "success");
  }

  return (
    <>
      <div className="glass grid gap-4 rounded-2xl px-5 py-4 sm:grid-cols-2 lg:grid-cols-[minmax(260px,1.4fr)_minmax(150px,1fr)_minmax(170px,1fr)_minmax(170px,1fr)] lg:items-center">
        <div className="flex min-w-0 items-center gap-3">
          <button
            type="button"
            onClick={handleCoreClick}
            className="flex h-9 shrink-0 items-center gap-2 rounded-full border border-subtle bg-canvas/70 px-3 text-xs transition-colors duration-200 hover:bg-hover"
            title={running ? t("core.clickToStop") : t("core.clickToStart")}
          >
            <StatusDot state={running ? "online" : "stopped"} />
            <span className="text-ink-primary">{running ? t("common.active") : t("common.stopped")}</span>
          </button>
          <span className="truncate font-mono text-xs text-ink-tertiary">{metrics.coreVersion}</span>
        </div>

        <Metric label={t("dashboard.uptime")} value={formatUptime(metrics.uptimeSeconds)} />
        <Metric label={t("dashboard.totalSent")} value={formatBytes(metrics.totalSent)} />
        <Metric label={t("dashboard.totalReceived")} value={formatBytes(metrics.totalReceived)} />
      </div>

      <Modal open={confirmStop} onClose={() => setConfirmStop(false)} width="max-w-[420px]">
        <ModalHeader title={t("core.stopQuestion")} subtitle={t("core.stopBody")} onClose={() => setConfirmStop(false)} />
        <ModalBody className="py-4">
          <div className="flex items-center gap-3 rounded-xl border border-subtle bg-canvas/70 p-3 text-sm text-ink-secondary">
            <StatusDot state="online" />
            <span>{metrics.coreVersion}</span>
          </div>
        </ModalBody>
        <ModalFooter>
          <Button variant="secondary" onClick={() => setConfirmStop(false)}>
            {t("common.keepRunning")}
          </Button>
          <Button variant="danger" onClick={stopCore}>
            {t("core.stopConfirm")}
          </Button>
        </ModalFooter>
      </Modal>
    </>
  );
}

function Metric({ label, value }: { label: string; value: string }) {
  return (
    <div className="min-w-0 rounded-xl border border-white/8 bg-canvas/30 px-3 py-2">
      <span className="block truncate text-xs text-ink-secondary">{label}</span>
      <span className="block truncate font-mono text-sm text-ink-primary">{value}</span>
    </div>
  );
}
