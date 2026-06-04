import { useEffect, useState } from "react";
import { Download, ExternalLink, RefreshCw } from "lucide-react";

import { ApiError, getToken } from "@/api/client";
import { getPanelVersion, startPanelUpdate, type PanelVersionDTO } from "@/api/panel";
import { Button } from "@/components/ui/button";
import { Card, CardHeader, CardLabel } from "@/components/ui/card";
import { useToast } from "@/components/ui/toast";
import { useI18n } from "@/lib/i18n";
import { cn } from "@/lib/utils";

function apiMessage(err: unknown, fallback: string): string {
  if (
    err instanceof ApiError &&
    err.body &&
    typeof err.body === "object" &&
    "error" in err.body
  ) {
    return String((err.body as { error: unknown }).error);
  }
  return fallback;
}

export function PanelVersionCard() {
  const { t } = useI18n();
  const { push } = useToast();
  const [version, setVersion] = useState<PanelVersionDTO | null>(null);
  const [busy, setBusy] = useState(false);

  const load = async () => {
    if (!getToken()) return;
    setBusy(true);
    try {
      setVersion(await getPanelVersion());
    } catch (err) {
      push(apiMessage(err, t("panel.versionLoadFailed")), "error");
    } finally {
      setBusy(false);
    }
  };

  useEffect(() => {
    load();
  }, []);

  const runUpdate = async () => {
    setBusy(true);
    try {
      const next = await startPanelUpdate();
      setVersion(next);
      push(t("panel.updateStarted"), "success");
    } catch (err) {
      push(apiMessage(err, t("panel.updateFailed")), "error");
    } finally {
      setBusy(false);
    }
  };

  const status = version?.status ?? "check_failed";
  const canUpdate = Boolean(version?.updateAvailable) && status !== "running";

  return (
    <Card>
      <CardHeader>
        <CardLabel>{t("panel.title")}</CardLabel>
        <Button
          type="button"
          variant="ghost"
          size="sm"
          loading={busy && status !== "running"}
          onClick={load}
          title={t("panel.refresh")}
          aria-label={t("panel.refresh")}
          className="size-8 px-0"
        >
          <RefreshCw size={15} />
        </Button>
      </CardHeader>

      <div className="space-y-3">
        <div>
          <div className="text-xs text-ink-tertiary">{t("panel.current")}</div>
          <div className="truncate font-mono text-lg text-ink-primary">{version?.currentVersion || "dev"}</div>
        </div>
        <div className="grid grid-cols-[minmax(0,1fr)_auto] items-end gap-3">
          <div className="min-w-0">
            <div className="text-xs text-ink-tertiary">{t("panel.latest")}</div>
            <div className="truncate font-mono text-sm text-ink-secondary">{version?.latestVersion || t("panel.unknown")}</div>
          </div>
          {version?.releaseURL ? (
            <a
              href={version.releaseURL}
              target="_blank"
              rel="noreferrer"
              className="grid size-8 place-items-center rounded-lg border border-subtle text-ink-tertiary transition-colors duration-200 hover:bg-hover hover:text-ink-primary"
              title={t("panel.release")}
              aria-label={t("panel.release")}
            >
              <ExternalLink size={14} />
            </a>
          ) : null}
        </div>
        <div className={cn("text-xs", version?.updateAvailable ? "text-amber" : "text-ink-tertiary")}>
          {statusText(status, t)}
        </div>
        <Button
          type="button"
          variant={canUpdate ? "white" : "secondary"}
          size="sm"
          className="w-full"
          loading={busy && status === "running"}
          disabled={!canUpdate || busy}
          onClick={runUpdate}
        >
          <Download size={14} />
          {t("panel.update")}
        </Button>
      </div>
    </Card>
  );
}

function statusText(status: PanelVersionDTO["status"], t: ReturnType<typeof useI18n>["t"]) {
  switch (status) {
    case "update_available":
      return t("panel.status.update_available");
    case "running":
      return t("panel.status.running");
    case "failed":
      return t("panel.status.failed");
    case "updated":
      return t("panel.status.updated");
    case "development":
      return t("panel.status.development");
    case "check_failed":
      return t("panel.status.check_failed");
    case "not_configured":
      return t("panel.status.not_configured");
    default:
      return t("panel.status.up_to_date");
  }
}
