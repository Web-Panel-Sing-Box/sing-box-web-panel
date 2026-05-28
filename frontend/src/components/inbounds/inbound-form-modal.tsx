
import { useEffect, useMemo, useRef, useState } from "react";
import { motion } from "framer-motion";
import { Copy, Dices, RefreshCw, Trash2 } from "lucide-react";

import { Accordion } from "@/components/ui/accordion";
import { Button } from "@/components/ui/button";
import { DateInput, Input, Label, NumberInput, Textarea } from "@/components/ui/input";
import { Modal, ModalBody, ModalFooter, ModalHeader } from "@/components/ui/modal";
import { Segmented } from "@/components/ui/tabs";
import { Select } from "@/components/ui/select";
import { Toggle } from "@/components/ui/toggle";
import { useToast } from "@/components/ui/toast";
import { useStore } from "@/lib/mock/store";
import {
  PROTOCOL_OPTIONS,
  TRAFFIC_RESET_OPTIONS,
  TRANSMISSION_OPTIONS,
  type Inbound,
  type Protocol,
  type TlsMode,
  type Transmission
} from "@/lib/mock/inbounds";
import { useI18n } from "@/lib/i18n";

export type InboundFormMode = "create" | "edit" | "clone";

type InboundFormModalProps = {
  open: boolean;
  mode?: InboundFormMode;
  inbound?: Inbound | null;
  onClose: () => void;
  onClone?: (inbound: Inbound) => void;
};

function randomPort() {
  return Math.floor(10_000 + Math.random() * 50_000);
}

function randomHex(length: number) {
  if (typeof crypto !== "undefined" && "getRandomValues" in crypto) {
    const bytes = new Uint8Array(length / 2);
    crypto.getRandomValues(bytes);
    return Array.from(bytes).map((b) => b.toString(16).padStart(2, "0")).join("");
  }
  let s = "";
  for (let i = 0; i < length; i++) s += Math.floor(Math.random() * 16).toString(16);
  return s;
}

function makeUuid() {
  if (typeof crypto !== "undefined" && "randomUUID" in crypto) return crypto.randomUUID();
  return "00000000-0000-4000-8000-000000000000";
}

export function InboundFormModal({ open, mode = "create", inbound, onClose, onClone }: InboundFormModalProps) {
  const { addInbound, updateInbound, removeInbound } = useStore();
  const { push } = useToast();
  const { t } = useI18n();
  const remarkRef = useRef<HTMLInputElement>(null);
  const [confirmDelete, setConfirmDelete] = useState(false);

  const [remark, setRemark] = useState("");
  const [protocol, setProtocol] = useState<Protocol>("naive");
  const [port, setPort] = useState<number | string>(() => randomPort());
  const [trafficReset, setTrafficReset] = useState("never");
  const [transmission, setTransmission] = useState<Transmission>("tcp");

  const [sniffing, setSniffing] = useState(true);
  const [snifHttp, setSnifHttp] = useState(true);
  const [snifTls, setSnifTls] = useState(true);
  const [snifQuic, setSnifQuic] = useState(false);
  const [snifFakedns, setSnifFakedns] = useState(false);
  const [metadataOnly, setMetadataOnly] = useState(false);
  const [routeOnly, setRouteOnly] = useState(true);
  const [ipsExcluded, setIpsExcluded] = useState("");
  const [domainsExcluded, setDomainsExcluded] = useState("");

  const [tls, setTls] = useState<TlsMode>("none");
  const [dest, setDest] = useState("www.cloudflare.com:443");
  const [sni, setSni] = useState("www.cloudflare.com");
  const [shortIds, setShortIds] = useState("");
  const [privateKey, setPrivateKey] = useState("");
  const [publicKey, setPublicKey] = useState("");

  const [userId, setUserId] = useState("user-001");
  const [uuid, setUuid] = useState(makeUuid());
  const [subscription, setSubscription] = useState("");
  const [totalFlowGb, setTotalFlowGb] = useState("100");
  const [expiry, setExpiry] = useState("");
  const [startAfterFirstUse, setStartAfterFirstUse] = useState(false);

  const [diceSpin, setDiceSpin] = useState(0);
  const [saving, setSaving] = useState(false);

  // Reset on close
  useEffect(() => {
    if (!open) return;
    setRemark(mode === "clone" && inbound ? `${inbound.remark}-copy` : inbound?.remark ?? "");
    setProtocol(inbound?.protocol ?? "naive");
    setPort(mode === "clone" ? randomPort() : inbound?.port ?? randomPort());
    setTrafficReset("never");
    setTransmission(inbound?.transmission ?? "tcp");
    setTls(inbound?.tls ?? "none");
    setSni(inbound?.sni ?? "www.cloudflare.com");
    setDest(inbound?.dest ?? "www.cloudflare.com:443");
    setUuid(makeUuid());
    setUserId("user-001");
    setSubscription("");
    setTotalFlowGb("100");
    setExpiry("");
    setPrivateKey("");
    setPublicKey("");
    setShortIds("");
    setStartAfterFirstUse(false);
    setConfirmDelete(false);
    if (mode === "clone") {
      window.setTimeout(() => {
        remarkRef.current?.focus();
        remarkRef.current?.select();
      }, 50);
    }
  }, [open, mode, inbound]);

  const transportFields = useMemo(() => {
    switch (transmission) {
      case "ws":
        return (
          <div className="grid grid-cols-1 gap-3 sm:grid-cols-2">
            <div>
              <Label>Path</Label>
              <Input placeholder="/ws" mono />
            </div>
            <div>
              <Label>Host</Label>
              <Input placeholder="panel.example" mono />
            </div>
          </div>
        );
      case "grpc":
        return (
          <div>
            <Label>Service name</Label>
            <Input placeholder="grpc-svc" mono />
          </div>
        );
      case "mkcp":
        return (
          <div>
            <Label>Header type</Label>
            <Select
              value="none"
              onChange={() => undefined}
              options={[
                { value: "none", label: "none" },
                { value: "srtp", label: "srtp" },
                { value: "wechat-video", label: "wechat-video" },
                { value: "wireguard", label: "wireguard" }
              ]}
            />
          </div>
        );
      case "xhttp":
      case "httpupgrade":
        return (
          <div className="grid grid-cols-1 gap-3 sm:grid-cols-2">
            <div>
              <Label>Path</Label>
              <Input placeholder="/up" mono />
            </div>
            <div>
              <Label>Host</Label>
              <Input placeholder="panel.example" mono />
            </div>
          </div>
        );
      default:
        return null;
    }
  }, [transmission]);

  async function handleSave() {
    if (!remark.trim()) {
      push(t("inbounds.remarkRequired"), "error");
      return;
    }
    setSaving(true);
    await new Promise((r) => setTimeout(r, 650));
    const payload = {
      remark: remark.trim(),
      protocol,
      port: Number(port) || randomPort(),
      transmission,
      tls,
      sni: tls === "none" ? undefined : sni,
      dest: tls === "reality" ? dest : undefined
    };
    if (mode === "edit" && inbound) {
      updateInbound(inbound.id, payload);
      push(t("inbounds.updated"), "success");
    } else {
      addInbound(payload);
      push(mode === "clone" ? t("inbounds.cloned") : t("inbounds.created"), "success");
    }
    setSaving(false);
    onClose();
  }

  function handleDelete() {
    if (!inbound) return;
    removeInbound(inbound.id);
    push(t("inbounds.deleted", { remark: inbound.remark }), "success");
    setConfirmDelete(false);
    onClose();
  }

  const title =
    mode === "edit"
      ? t("inbounds.modalEdit")
      : mode === "clone"
        ? t("inbounds.modalClone")
        : t("inbounds.modalCreate");

  return (
    <>
    <Modal open={open} onClose={onClose} width="max-w-[760px]">
      <ModalHeader title={title} onClose={onClose} />
      <ModalBody className="space-y-3">
        <Accordion title={t("inbounds.basics")}>
          <div className="space-y-3">
            <div>
              <Label>{t("common.remark")}</Label>
              <Input
                ref={remarkRef}
                value={remark}
                onChange={(e) => setRemark(e.target.value)}
                placeholder="e.g. vadim-vless#0001"
                className={mode === "clone" ? "selection:bg-brand/40" : undefined}
              />
            </div>
            <div className="grid grid-cols-1 gap-3 sm:grid-cols-2">
              <div>
                <Label>{t("common.protocol")}</Label>
                <Select value={protocol} options={PROTOCOL_OPTIONS} onChange={(v) => setProtocol(v)} />
              </div>
              <div>
                <Label>{t("common.port")}</Label>
                <Input
                  type="number"
                  value={port}
                  onChange={(e) => setPort(e.target.value)}
                  mono
                  trailing={
                    <motion.button
                      type="button"
                      animate={{ rotate: diceSpin }}
                      transition={{ duration: 0.4, ease: "easeOut" }}
                      onClick={() => {
                        setDiceSpin((v) => v + 360);
                        setPort(randomPort());
                      }}
                      className="grid size-7 place-items-center rounded-md text-ink-secondary transition-colors duration-150 hover:bg-hover hover:text-ink-primary"
                      title="Randomize"
                    >
                      <Dices size={14} />
                    </motion.button>
                  }
                />
              </div>
            </div>
            <div className="grid grid-cols-1 gap-3 sm:grid-cols-2">
              <div>
                <Label>{t("inbounds.trafficReset")}</Label>
                <Select value={trafficReset} options={TRAFFIC_RESET_OPTIONS} onChange={setTrafficReset} />
              </div>
              <div>
                <Label>{t("inbounds.transmission")}</Label>
                <Select value={transmission} options={TRANSMISSION_OPTIONS} onChange={(v) => setTransmission(v)} />
              </div>
            </div>
            {transportFields ? <div className="rounded-lg border border-subtle bg-canvas/60 p-3">{transportFields}</div> : null}
          </div>
        </Accordion>

        <Accordion title={t("inbounds.transportSecurity")}>
          <div className="space-y-5">
            <div className="rounded-lg border border-subtle bg-canvas/60 p-3">
              <Toggle checked={sniffing} onChange={setSniffing} label={t("inbounds.enableSniffing")} description={t("inbounds.sniffingDesc")} />
              {sniffing ? (
                <div className="mt-4 space-y-3">
                  <div className="flex flex-wrap gap-3 text-xs text-ink-secondary">
                    <Checkbox label="HTTP" checked={snifHttp} onChange={setSnifHttp} />
                    <Checkbox label="TLS" checked={snifTls} onChange={setSnifTls} />
                    <Checkbox label="QUIC" checked={snifQuic} onChange={setSnifQuic} />
                    <Checkbox label="FAKEDNS" checked={snifFakedns} onChange={setSnifFakedns} />
                  </div>
                  <div className="flex flex-wrap gap-6">
                    <Toggle checked={metadataOnly} onChange={setMetadataOnly} label={t("inbounds.metadataOnly")} />
                    <Toggle checked={routeOnly} onChange={setRouteOnly} label={t("inbounds.routeOnly")} />
                  </div>
                  <div className="grid grid-cols-1 gap-3 sm:grid-cols-2">
                    <div>
                      <Label>{t("inbounds.ipsExcluded")}</Label>
                      <Textarea rows={2} value={ipsExcluded} onChange={(e) => setIpsExcluded(e.target.value)} mono placeholder="10.0.0.0/8" />
                    </div>
                    <div>
                      <Label>{t("inbounds.domainsExcluded")}</Label>
                      <Textarea rows={2} value={domainsExcluded} onChange={(e) => setDomainsExcluded(e.target.value)} mono placeholder="local.lan" />
                    </div>
                  </div>
                </div>
              ) : null}
            </div>

            <div>
              <Label>TLS</Label>
              <Segmented<TlsMode>
                value={tls}
                onChange={setTls}
                options={[
                  { value: "none", label: "None" },
                  { value: "tls", label: "TLS" },
                  { value: "reality", label: "Reality" }
                ]}
              />
            </div>

            {tls === "reality" ? (
              <div className="grid grid-cols-1 gap-3">
                <div className="grid grid-cols-1 gap-3 sm:grid-cols-2">
                  <div>
                    <Label>{t("inbounds.destination")}</Label>
                    <Input value={dest} onChange={(e) => setDest(e.target.value)} mono />
                  </div>
                  <div>
                    <Label>SNI</Label>
                    <Input value={sni} onChange={(e) => setSni(e.target.value)} mono />
                  </div>
                </div>
                <div>
                  <Label>{t("inbounds.shortIds")}</Label>
                  <Textarea rows={2} value={shortIds} onChange={(e) => setShortIds(e.target.value)} mono placeholder="One short id per line" />
                </div>
                <div className="space-y-2 rounded-lg border border-subtle bg-canvas/60 p-3">
                  <div className="flex items-center justify-between">
                    <span className="text-xs text-ink-secondary">{t("inbounds.keypair")}</span>
                    <Button
                      size="sm"
                      onClick={() => {
                        setPrivateKey(randomHex(64));
                        setPublicKey(randomHex(64));
                      }}
                    >
                      <RefreshCw size={14} />
                      {t("inbounds.generateKeypair")}
                    </Button>
                  </div>
                  <div>
                    <Label>{t("inbounds.privateKey")}</Label>
                    <Input value={privateKey} mono readOnly placeholder="—" />
                  </div>
                  <div>
                    <Label>{t("inbounds.publicKey")}</Label>
                    <Input value={publicKey} mono readOnly placeholder="—" />
                  </div>
                </div>
              </div>
            ) : tls === "tls" ? (
              <div className="grid grid-cols-1 gap-3 sm:grid-cols-2">
                <div>
                  <Label>SNI</Label>
                  <Input value={sni} onChange={(e) => setSni(e.target.value)} mono />
                </div>
              </div>
            ) : null}
          </div>
        </Accordion>

        <Accordion title={t("inbounds.userTemplate")}>
          <div className="space-y-3">
            <div className="grid grid-cols-1 gap-3 sm:grid-cols-2">
              <div>
                <Label>{t("inbounds.userId")}</Label>
                <Input value={userId} onChange={(e) => setUserId(e.target.value)} />
              </div>
              <div>
                <Label>UUID</Label>
                <Input
                  value={uuid}
                  mono
                  readOnly
                  trailing={
                    <button
                      type="button"
                      onClick={() => setUuid(makeUuid())}
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
              <Label>{t("inbounds.subscription")}</Label>
              <Input value={subscription} onChange={(e) => setSubscription(e.target.value)} placeholder="https://panel.example/sub/your-key" mono />
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
            <div className="flex min-h-[72px] items-center justify-center rounded-lg border border-subtle bg-canvas/40 px-3 py-2">
              <Toggle
                size="lg"
                checked={startAfterFirstUse}
                onChange={setStartAfterFirstUse}
                label={t("inbounds.startAfterFirstUse")}
                description={t("inbounds.startAfterFirstUseDesc")}
              />
            </div>
          </div>
        </Accordion>
      </ModalBody>
      <ModalFooter>
        {mode === "edit" && inbound ? (
          <>
            <IconActionButton title={t("common.clone")} onClick={() => onClone?.(inbound)}>
              <Copy size={16} />
            </IconActionButton>
            <IconActionButton title={t("common.delete")} danger onClick={() => setConfirmDelete(true)}>
              <Trash2 size={16} />
            </IconActionButton>
            <div className="flex-1" />
          </>
        ) : null}
        <Button variant="danger" onClick={onClose}>
          {t("common.cancel")}
        </Button>
        <Button variant="primary" onClick={handleSave} loading={saving}>
          {t("common.save")}
        </Button>
      </ModalFooter>
    </Modal>
    <Modal open={confirmDelete} onClose={() => setConfirmDelete(false)} width="max-w-[420px]">
      <ModalHeader title={t("inbounds.deleteQuestion")} subtitle={inbound ? t("inbounds.deleteBody", { remark: inbound.remark }) : undefined} onClose={() => setConfirmDelete(false)} />
      <ModalFooter>
        <Button variant="secondary" onClick={() => setConfirmDelete(false)}>
          {t("common.cancel")}
        </Button>
        <Button variant="danger" onClick={handleDelete}>
          {t("common.delete")}
        </Button>
      </ModalFooter>
    </Modal>
    </>
  );
}

type CheckboxProps = {
  label: string;
  checked: boolean;
  onChange: (v: boolean) => void;
};

function IconActionButton({
  title,
  onClick,
  danger,
  children
}: {
  title: string;
  onClick: () => void;
  danger?: boolean;
  children: React.ReactNode;
}) {
  return (
    <button
      type="button"
      title={title}
      aria-label={title}
      onClick={onClick}
      className={`grid size-9 place-items-center rounded-lg border border-subtle bg-canvas text-ink-secondary transition-colors duration-200 ${danger ? "hover:border-danger/40 hover:text-danger" : "hover:border-white/20 hover:text-ink-primary"}`}
    >
      {children}
    </button>
  );
}

function Checkbox(props: CheckboxProps) {
  const { label, checked, onChange } = props;

  return (
    <button
      type="button"
      onClick={() => onChange(!checked)}
      className="inline-flex items-center gap-2"
    >
      <span
        className={`grid size-4 place-items-center rounded border ${checked ? "border-brand bg-brand text-white" : "border-white/15 bg-canvas"}`}
      >
        {checked ? (
          <svg viewBox="0 0 12 12" width="10" height="10" fill="none">
            <path d="m2 6 3 3 5-6" stroke="currentColor" strokeWidth="1.6" strokeLinecap="round" strokeLinejoin="round" />
          </svg>
        ) : null}
      </span>
      <span className="text-xs text-ink-secondary">{label}</span>
    </button>
  );
}
