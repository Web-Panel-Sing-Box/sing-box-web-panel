import { useEffect, useMemo, useRef, useState } from "react";

import { ApiError } from "@/api/client";
import { listNodes, type NodeDTO } from "@/api";
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
  defaultNodeId?: string;
};

export function AddClientModal({ open, onClose, defaultInboundId, defaultNodeId }: Props) {
  const inbounds = useInbounds();
  // Keep the latest inbounds in a ref so the init effect can read them without
  // re-running every time polling delivers a new array (SIN-35).
  const inboundsRef = useRef(inbounds);
  inboundsRef.current = inbounds;
  const { addClient } = useStoreActions();
  const { push } = useToast();
  const { t } = useI18n();

  const [name, setName] = useState("");
  const [nodes, setNodes] = useState<NodeDTO[]>([]);
  const [nodeId, setNodeId] = useState("local");
  const [inboundId, setInboundId] = useState("");
  const [totalFlowGb, setTotalFlowGb] = useState("100");
  const [expiry, setExpiry] = useState("");
  const [startAfterFirstUse, setStartAfterFirstUse] = useState(false);
  const [saving, setSaving] = useState(false);
  const [nameError, setNameError] = useState<string | undefined>(undefined);
  const nodeOptions = useMemo(
    () => [
      { value: "local", label: t("common.local") },
      ...nodes
        .filter((node) => node.enabled && node.hasApiToken)
        .map((node) => ({ value: node.id, label: node.name })),
    ],
    [nodes, t],
  );
  const inboundOptions = useMemo(
    () =>
      inbounds
        .filter((inbound) => (nodeId === "local" ? !inbound.nodeId : inbound.nodeId === nodeId))
        .map((inbound) => ({ value: inbound.id, label: inbound.remark })),
    [inbounds, nodeId],
  );

  useEffect(() => {
    if (!open) return;
    void listNodes()
      .then(setNodes)
      .catch(() => setNodes([]));
    setName("");
    setNameError(undefined);
    const currentInbounds = inboundsRef.current;
    const defaultInbound = defaultInboundId
      ? currentInbounds.find((inbound) => inbound.id === defaultInboundId)
      : undefined;
    const nextNodeId = defaultInbound?.nodeId ?? (defaultNodeId && defaultNodeId !== "all" ? defaultNodeId : "local");
    const nextInbound =
      defaultInbound && (nextNodeId === "local" ? !defaultInbound.nodeId : defaultInbound.nodeId === nextNodeId)
        ? defaultInbound.id
        : (currentInbounds.find((inbound) => (nextNodeId === "local" ? !inbound.nodeId : inbound.nodeId === nextNodeId))?.id ?? "");
    setNodeId(nextNodeId);
    setInboundId(nextInbound);
    setTotalFlowGb("100");
    setExpiry(defaultExpiryIso().slice(0, 10));
    setStartAfterFirstUse(false);
  }, [open, defaultInboundId, defaultNodeId]);

  useEffect(() => {
    if (!open) return;
    if (inboundOptions.some((option) => option.value === inboundId)) return;
    setInboundId(inboundOptions[0]?.value ?? "");
  }, [open, inboundId, inboundOptions]);

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
    try {
      await addClient({
        ...(nodeId !== "local" ? { nodeId } : {}),
        name: name.trim(),
        inboundId,
        totalQuota: (Number(totalFlowGb) || 0) * GB,
        expiry: expiry ? new Date(expiry).toISOString() : defaultExpiryIso(),
        startAfterFirstUse,
      });
      push(t("clients.created"), "success");
      onClose();
    } catch (err) {
      const body = err instanceof ApiError ? err.body : null;
      const message =
        body && typeof body === "object" && body !== null && "error" in body
          ? String((body as { error: unknown }).error)
          : t("clients.createFailed");
      push(message, "error");
    } finally {
      setSaving(false);
    }
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
          <Label>{t("common.node")}</Label>
          <Select
            value={nodeId}
            options={nodeOptions}
            onChange={setNodeId}
          />
        </div>
        <div>
          <Label>{t("clients.inbound")}</Label>
          <Select
            value={inboundId}
            options={inboundOptions}
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
