import { useCallback, useEffect, useRef, useState } from "react";

import { useToast } from "@/components/ui/toast";
import { useI18n } from "@/lib/i18n";
import { useStoreActions } from "@/lib/mock/store";
import type { Inbound, Protocol, TlsMode, Transmission } from "@/lib/mock/inbounds";
import { makeUuid, randomHex, randomPort } from "@/lib/random";

export type InboundFormMode = "create" | "edit" | "clone";

type Params = {
  open: boolean;
  mode: InboundFormMode;
  inbound: Inbound | null | undefined;
  onClose: () => void;
};

export function useInboundForm({ open, mode, inbound, onClose }: Params) {
  const { addInbound, updateInbound, removeInbound } = useStoreActions();
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

  const randomizePort = useCallback(() => {
    setDiceSpin((v) => v + 360);
    setPort(randomPort());
  }, []);

  const regenerateUuid = useCallback(() => {
    setUuid(makeUuid());
  }, []);

  const generateKeypair = useCallback(() => {
    setPrivateKey(randomHex(64));
    setPublicKey(randomHex(64));
  }, []);

  const handleSave = useCallback(async () => {
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
  }, [remark, protocol, port, transmission, tls, sni, dest, mode, inbound, push, t, addInbound, updateInbound, onClose]);

  const handleDelete = useCallback(() => {
    if (!inbound) return;
    removeInbound(inbound.id);
    push(t("inbounds.deleted", { remark: inbound.remark }), "success");
    setConfirmDelete(false);
    onClose();
  }, [inbound, removeInbound, push, t, onClose]);

  const openConfirmDelete = useCallback(() => setConfirmDelete(true), []);
  const closeConfirmDelete = useCallback(() => setConfirmDelete(false), []);

  return {
    remarkRef,
    confirmDelete,
    openConfirmDelete,
    closeConfirmDelete,
    remark, setRemark,
    protocol, setProtocol,
    port, setPort,
    trafficReset, setTrafficReset,
    transmission, setTransmission,
    sniffing, setSniffing,
    snifHttp, setSnifHttp,
    snifTls, setSnifTls,
    snifQuic, setSnifQuic,
    snifFakedns, setSnifFakedns,
    metadataOnly, setMetadataOnly,
    routeOnly, setRouteOnly,
    ipsExcluded, setIpsExcluded,
    domainsExcluded, setDomainsExcluded,
    tls, setTls,
    dest, setDest,
    sni, setSni,
    shortIds, setShortIds,
    privateKey, publicKey, generateKeypair,
    userId, setUserId,
    uuid, regenerateUuid,
    subscription, setSubscription,
    totalFlowGb, setTotalFlowGb,
    expiry, setExpiry,
    startAfterFirstUse, setStartAfterFirstUse,
    diceSpin, randomizePort,
    saving, handleSave, handleDelete
  };
}
