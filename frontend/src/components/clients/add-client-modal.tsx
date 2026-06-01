import { useEffect, useState } from "react";

import { Button } from "@/components/ui/button";
import { DateInput, Input, Label, NumberInput } from "@/components/ui/input";
import { Modal, ModalBody, ModalFooter, ModalHeader } from "@/components/ui/modal";
import { Select } from "@/components/ui/select";
import { Toggle } from "@/components/ui/toggle";
import { useToast } from "@/components/ui/toast";
import { useInbounds, useStoreActions } from "@/lib/store";
import { useI18n } from "@/lib/i18n";

const GB = 1024 ** 3;

function defaultExpiryIso() {
  const d = new Date();
  d.setMonth(d.getMonth() + 1);
  return d.toISOString();
}

type Props = {
  open: boolean;
  onClose: () => void;
  /** Optional inbound pre-selection — e.g. when navigated from the filter bar. */
  defaultInboundId?: string;
};

export function AddClientModal({ open, onClose, defaultInboundId }: Props) {
  const inbounds = useInbounds();
  const { addClient } = useStoreActions();
  const { push } = useToast();
  const { t } = useI18n();

  const [name, setName] = useState("");
  const [inboundId, setInboundId] = useState("");
  const [totalFlowGb, setTotalFlowGb] = useState("100");
  const [expiry, setExpiry] = useState("");
  const [startAfterFirstUse, setStartAfterFirstUse] = useState(false);
  const [saving, setSaving] = useState(false);
  const [nameError, setNameError] = useState<string | undefined>(undefined);

  useEffect(() => {
    if (!open) return;
    setName("");
    setNameError(undefined);
    setInboundId(defaultInboundId ?? inbounds[0]?.id ?? "");
    setTotalFlowGb("100");
    setExpiry(defaultExpiryIso().slice(0, 10));
    setStartAfterFirstUse(false);
  }, [open, defaultInboundId, inbounds]);

  async function handleSave() {
    if (!name.trim()) {
      setNameError(t("clients.nameRequired"));
      return;
    }
    if (!inboundId) {
      push(t("clients.inboundRequired"), "error");
      return;
    }
    setSaving(true);
    await new Promise((r) => setTimeout(r, 500));
    addClient({
      name: name.trim(),
      inboundId,
      totalQuota: (Number(totalFlowGb) || 0) * GB,
      expiry: expiry ? new Date(expiry).toISOString() : defaultExpiryIso(),
      startAfterFirstUse
    });
    setSaving(false);
    push(t("clients.created"), "success");
    onClose();
  }

  return (
    <Modal open={open} onClose={onClose} width="max-w-[560px]">
      <ModalHeader title={t("clients.addTitle")} onClose={onClose} />
      <ModalBody className="space-y-3">
        <div>
          <Label>{t("clients.userName")}</Label>
          <Input
            value={name}
            onChange={(e) => {
              setName(e.target.value);
              if (nameError) setNameError(undefined);
            }}
            placeholder="e.g. vadim_denisych#0001"
            error={nameError}
          />
        </div>
        <div>
          <Label>{t("clients.inbound")}</Label>
          <Select
            value={inboundId}
            options={inbounds.map((i) => ({ value: i.id, label: i.remark }))}
            onChange={setInboundId}
          />
        </div>
        <div className="grid grid-cols-1 gap-3 sm:grid-cols-2">
          <div>
            <Label>{t("inbounds.totalFlow")}</Label>
            <NumberInput value={totalFlowGb} onChange={setTotalFlowGb} min={0} mono placeholder="0" />
          </div>
          <div>
            <Label>{t("inbounds.expiryDate")}</Label>
            <DateInput value={expiry} onChange={setExpiry} />
          </div>
        </div>
        <div className="flex min-h-[72px] items-center justify-center rounded-lg rounded-lg border border-subtle bg-canvas/40 px-3 py-2">
          <Toggle
            size="lg"
            checked={startAfterFirstUse}
            onChange={setStartAfterFirstUse}
            label={t("inbounds.startAfterFirstUse")}
          />
        </div>
      </ModalBody>
      <ModalFooter>
        <Button variant="danger" onClick={onClose}>
          {t("common.cancel")}
        </Button>
        <Button variant="primary" onClick={handleSave} loading={saving}>
          {t("common.save")}
        </Button>
      </ModalFooter>
    </Modal>
  );
}
