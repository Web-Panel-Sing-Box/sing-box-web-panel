
import { useEffect, useState } from "react";
import { Copy, QrCode, RefreshCw } from "lucide-react";

import { Button } from "@/components/ui/button";
import { Input, Label } from "@/components/ui/input";
import { Modal, ModalBody, ModalFooter, ModalHeader } from "@/components/ui/modal";
import { Select } from "@/components/ui/select";
import { Toggle } from "@/components/ui/toggle";
import { useToast } from "@/components/ui/toast";
import type { Client, ClientStatus } from "@/lib/mock/clients";
import { useStore } from "@/lib/mock/store";

import { QrModal } from "./qr-modal";

const GB = 1024 ** 3;

type Props = {
  client: Client | null;
  onClose: () => void;
};

export function ClientDetailModal({ client, onClose }: Props) {
  const { inbounds, updateClient } = useStore();
  const { push } = useToast();

  const [draft, setDraft] = useState<Client | null>(client);
  const [qrOpen, setQrOpen] = useState(false);
  const [copied, setCopied] = useState(false);

  useEffect(() => {
    setDraft(client);
    setCopied(false);
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
    push("Link copied to clipboard");
    window.setTimeout(() => setCopied(false), 1500);
  }

  function save() {
    if (!draft) return;
    updateClient(draft.id, draft);
    push("Client updated");
    onClose();
  }

  return (
    <>
      <Modal open={!!client} onClose={onClose} width="max-w-[640px]">
        <ModalHeader title={draft.name} subtitle="Edit credentials, quota, and lifecycle" onClose={onClose} />
        <ModalBody className="space-y-3">
          <div className="grid grid-cols-1 gap-3 sm:grid-cols-2">
            <div>
              <Label>User name</Label>
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
            <Label>Inbound</Label>
            <Select
              value={draft.inboundId}
              options={inbounds.map((i) => ({ value: i.id, label: i.remark }))}
              onChange={(v) => update("inboundId", v)}
            />
          </div>
          <div>
            <Label>Subscription</Label>
            <Input value={draft.subscription} mono onChange={(e) => update("subscription", e.target.value)} />
          </div>
          <div className="grid grid-cols-1 gap-3 sm:grid-cols-2">
            <div>
              <Label>Total flow (GB)</Label>
              <Input
                type="number"
                value={Math.round(draft.totalQuota / GB)}
                onChange={(e) => update("totalQuota", Number(e.target.value) * GB)}
                mono
              />
            </div>
            <div>
              <Label>Expiry date</Label>
              <Input
                type="date"
                value={draft.expiry.slice(0, 10)}
                onChange={(e) => update("expiry", new Date(e.target.value).toISOString())}
              />
            </div>
          </div>
          <div className="grid grid-cols-1 gap-3 sm:grid-cols-2">
            <div>
              <Label>Status</Label>
              <Select<ClientStatus>
                value={draft.status}
                options={[
                  { value: "active", label: "Active" },
                  { value: "disabled", label: "Disabled" },
                  { value: "expired", label: "Expired" }
                ]}
                onChange={(v) => update("status", v)}
              />
            </div>
            <div className="flex items-end">
              <Toggle
                checked={draft.startAfterFirstUse}
                onChange={(v) => update("startAfterFirstUse", v)}
                label="Start after first use"
              />
            </div>
          </div>
        </ModalBody>
        <ModalFooter accent="cyan">
          <Button variant="secondary" onClick={() => setQrOpen(true)}>
            <QrCode size={14} />
            Get QR
          </Button>
          <Button variant="secondary" onClick={copyLink}>
            <Copy size={14} />
            {copied ? "Copied!" : "Copy link"}
          </Button>
          <div className="flex-1" />
          <Button variant="danger" onClick={onClose}>
            Cancel
          </Button>
          <Button variant="primary" onClick={save}>
            Save
          </Button>
        </ModalFooter>
      </Modal>
      <QrModal open={qrOpen} onClose={() => setQrOpen(false)} payload={draft.subscription} />
    </>
  );
}
