
import { useEffect, useState } from "react";
import { Copy, QrCode, RefreshCw } from "lucide-react";

import { Button } from "@/components/ui/button";
import { DateInput, Input, Label, NumberInput } from "@/components/ui/input";
import { Modal, ModalBody, ModalFooter, ModalHeader } from "@/components/ui/modal";
import { Select } from "@/components/ui/select";
import { Toggle } from "@/components/ui/toggle";
import { useToast } from "@/components/ui/toast";
import type { Client, ClientStatus } from "@/lib/mock/clients";
import { useStore } from "@/lib/mock/store";
import { useI18n } from "@/lib/i18n";

import { QrModal } from "./qr-modal";

const GB = 1024 ** 3;

type Props = {
  client: Client | null;
  onClose: () => void;
};

export function ClientDetailModal({ client, onClose }: Props) {
  const { inbounds, updateClient } = useStore();
  const { push } = useToast();
  const { t } = useI18n();

  const [draft, setDraft] = useState<Client | null>(client);
  const [qrOpen, setQrOpen] = useState(false);
  const [copied, setCopied] = useState(false);
  const [totalFlowGb, setTotalFlowGb] = useState("");

  useEffect(() => {
    setDraft(client);
    setCopied(false);
    if (client) {
      const gb = Math.round(client.totalQuota / GB);
      setTotalFlowGb(gb > 0 ? String(gb) : "");
    }
  }, [client]);

  if (!draft || !client) {
    return <Modal open={false} onClose={onClose}>{null}</Modal>;
  }

  function update<K extends keyof Client>(key: K, value: Client[K]) {
    setDraft((prev) => (prev ? { ...prev, [key]: value } : prev));
  }

  async function copyLink() {
    if (!draft) return;
    try {
      await navigator.clipboard.writeText(draft.subscription);
    } catch {
      // ignore
    }
    setCopied(true);
    push(t("clients.linkCopied"), "success");
    window.setTimeout(() => setCopied(false), 1500);
  }

  function save() {
    if (!draft) return;
    const parsedFlow = totalFlowGb === "" ? 0 : Number(totalFlowGb);
    updateClient(draft.id, { ...draft, totalQuota: parsedFlow * GB });
    push(t("clients.updated"), "success");
    onClose();
  }

  return (
    <>
      <Modal open={!!client} onClose={onClose} width="max-w-[640px]">
        <ModalHeader title={draft.name} onClose={onClose} />
        <ModalBody className="space-y-3">
          <div className="grid grid-cols-1 gap-3 sm:grid-cols-2">
            <div>
              <Label>{t("clients.userName")}</Label>
              <Input value={draft.name} onChange={(e) => update("name", e.target.value)} />
            </div>
            <div>
              <Label>UUID</Label>
              <Input
                value={draft.uuid}
                mono
                readOnly
                trailing={
                  <button
                    type="button"
                    onClick={() => update("uuid", crypto.randomUUID())}
                    className="grid size-7 place-items-center rounded-md text-ink-secondary transition-colors duration-150 hover:bg-hover hover:text-ink-primary"
                    title="Regenerate"
                  >
                    <RefreshCw size={14} />
                  </button>
                }
              />
            </div>
          </div>
          <div>
            <Label>{t("clients.inbound")}</Label>
            <Select
              value={draft.inboundId}
              options={inbounds.map((i) => ({ value: i.id, label: i.remark }))}
              onChange={(v) => update("inboundId", v)}
            />
          </div>
          <div>
            <Label>{t("inbounds.subscription")}</Label>
            <Input value={draft.subscription} mono onChange={(e) => update("subscription", e.target.value)} />
          </div>
          <div className="grid grid-cols-1 gap-3 sm:grid-cols-2">
            <div>
              <Label>{t("inbounds.totalFlow")}</Label>
              <NumberInput value={totalFlowGb} onChange={setTotalFlowGb} min={0} mono placeholder="0" />
            </div>
            <div>
              <Label>{t("inbounds.expiryDate")}</Label>
              <DateInput
                value={draft.expiry.slice(0, 10)}
                onChange={(v) => update("expiry", v ? new Date(v).toISOString() : draft.expiry)}
              />
            </div>
          </div>
          <div className="grid grid-cols-1 gap-3 sm:grid-cols-2">
            <div>
              <Label>{t("common.status")}</Label>
              <Select<ClientStatus>
                value={draft.status}
                options={[
                  { value: "active", label: t("common.active") },
                  { value: "disabled", label: t("common.disabled") },
                  { value: "expired", label: t("common.expired") }
                ]}
                onChange={(v) => update("status", v)}
              />
            </div>
            <div className="flex min-h-[66px] items-center justify-center rounded-lg border border-subtle bg-canvas/40 px-3">
              <Toggle
                checked={draft.startAfterFirstUse}
                onChange={(v) => update("startAfterFirstUse", v)}
                label={t("inbounds.startAfterFirstUse")}
              />
            </div>
          </div>
        </ModalBody>
        <ModalFooter>
          <Button variant="secondary" onClick={() => setQrOpen(true)}>
            <QrCode size={14} />
            {t("clients.getQr")}
          </Button>
          <Button variant="secondary" onClick={copyLink}>
            <Copy size={14} />
            {copied ? t("clients.copied") : t("clients.copyLink")}
          </Button>
          <div className="flex-1" />
          <Button variant="danger" onClick={onClose}>
            {t("common.cancel")}
          </Button>
          <Button variant="primary" onClick={save}>
            {t("common.save")}
          </Button>
        </ModalFooter>
      </Modal>
      <QrModal open={qrOpen} onClose={() => setQrOpen(false)} payload={draft.subscription} />
    </>
  );
}
