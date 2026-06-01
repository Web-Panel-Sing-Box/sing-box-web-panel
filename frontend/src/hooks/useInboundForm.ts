import { useCallback, useEffect, useRef, useState } from "react";

import { useToast } from "@/components/ui/toast";
import { useI18n } from "@/lib/i18n";
import { useStoreActions } from "@/lib/mock/store";
import {
  DEFAULT_NETWORK,
  DEFAULT_QUIC_CC,
  DEFAULT_VLESS_TRANSPORT,
  type Flow,
  type Inbound,
  type Network,
  type ObfsType,
  type Protocol,
  type QuicCc,
  type TlsMode,
  type VlessTransport
} from "@/lib/mock/inbounds";
import { makeUuid, randomHex, randomPort } from "@/lib/random";

export type InboundFormMode = "create" | "edit" | "clone";

type Params = {
  open: boolean;
  mode: InboundFormMode;
  inbound: Inbound | null | undefined;
  onClose: () => void;
};

// naive + hysteria2 always run over TLS; only vless can be plain / reality.
function tlsForProtocol(protocol: Protocol, current: TlsMode): TlsMode {
  if (protocol === "naive" || protocol === "hysteria2") return "tls";
  return current;
}

export function useInboundForm({ open, mode, inbound, onClose }: Params) {
  const { addInbound, updateInbound, removeInbound } = useStoreActions();
  const { push } = useToast();
  const { t } = useI18n();
  const remarkRef = useRef<HTMLInputElement>(null);

  const [confirmDelete, setConfirmDelete] = useState(false);
  const [remark, setRemark] = useState("");
  const [protocol, setProtocolState] = useState<Protocol>("naive");
  const [port, setPort] = useState<number | string>(() => randomPort());
  const [trafficReset, setTrafficReset] = useState("never");

  // vless transport
  const [transport, setTransport] = useState<VlessTransport>(DEFAULT_VLESS_TRANSPORT);
  const [flow, setFlow] = useState<Flow>("");

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

  // naive + hysteria2 auth
  const [username, setUsername] = useState("user-001");
  const [password, setPassword] = useState("");

  // naive transport-level
  const [network, setNetwork] = useState<Network>(DEFAULT_NETWORK);
  const [quicCc, setQuicCc] = useState<QuicCc>(DEFAULT_QUIC_CC);

  // hysteria2
  const [obfsType, setObfsType] = useState<ObfsType>("none");
  const [obfsPassword, setObfsPassword] = useState("");
  const [upMbps, setUpMbps] = useState<string>("100");
  const [downMbps, setDownMbps] = useState<string>("100");

  // vless user template
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
    const nextProtocol = inbound?.protocol ?? "naive";
    setRemark(mode === "clone" && inbound ? `${inbound.remark}-copy` : inbound?.remark ?? "");
    setProtocolState(nextProtocol);
    setPort(mode === "clone" ? randomPort() : inbound?.port ?? randomPort());
    setTrafficReset("never");
    setTransport(inbound?.transport ?? DEFAULT_VLESS_TRANSPORT);
    setFlow(inbound?.flow ?? "");
    setTls(tlsForProtocol(nextProtocol, inbound?.tls ?? "none"));
    setSni(inbound?.sni ?? "www.cloudflare.com");
    setDest(inbound?.dest ?? "www.cloudflare.com:443");
    setUsername(inbound?.username ?? "user-001");
    setPassword(inbound?.password ?? randomHex(12));
    setNetwork(inbound?.network ?? DEFAULT_NETWORK);
    setQuicCc(inbound?.quicCc ?? DEFAULT_QUIC_CC);
    setObfsType(inbound?.obfsType ?? "none");
    setObfsPassword(inbound?.obfsPassword ?? randomHex(12));
    setUpMbps(inbound?.upMbps != null ? String(inbound.upMbps) : "100");
    setDownMbps(inbound?.downMbps != null ? String(inbound.downMbps) : "100");
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

  // Switching protocol normalizes invalid combinations (naive/hy2 require TLS).
  const setProtocol = useCallback((next: Protocol) => {
    setProtocolState(next);
    setTls((current) => tlsForProtocol(next, current));
    if (next === "naive" || next === "hysteria2") {
      setPassword((p) => p || randomHex(12));
    }
  }, []);

  const randomizePort = useCallback(() => {
    setDiceSpin((v) => v + 360);
    setPort(randomPort());
  }, []);

  const regenerateUuid = useCallback(() => {
    setUuid(makeUuid());
  }, []);

  const regeneratePassword = useCallback(() => {
    setPassword(randomHex(12));
  }, []);

  const regenerateObfsPassword = useCallback(() => {
    setObfsPassword(randomHex(12));
  }, []);

  const generateKeypair = useCallback(() => {
    setPrivateKey(randomHex(64));
    setPublicKey(randomHex(64));
  }, []);

  const buildPayload = useCallback((): Omit<Inbound, "id" | "createdAt" | "clientCount" | "enabled"> => {
    const safePort = Number(port) || randomPort();
    const base = { remark: remark.trim(), protocol, port: safePort };
    if (protocol === "vless") {
      return {
        ...base,
        transport,
        flow: transport === "tcp" ? flow : "",
        tls,
        sni: tls === "none" ? undefined : sni,
        dest: tls === "reality" ? dest : undefined
      };
    }
    if (protocol === "naive") {
      return {
        ...base,
        tls: "tls",
        sni,
        username: username.trim(),
        password,
        network,
        quicCc
      };
    }
    // hysteria2
    return {
      ...base,
      tls: "tls",
      sni,
      username: username.trim(),
      password,
      obfsType,
      obfsPassword: obfsType === "salamander" ? obfsPassword : undefined,
      upMbps: Number(upMbps) || 0,
      downMbps: Number(downMbps) || 0
    };
  }, [
    remark,
    protocol,
    port,
    transport,
    flow,
    tls,
    sni,
    dest,
    username,
    password,
    network,
    quicCc,
    obfsType,
    obfsPassword,
    upMbps,
    downMbps
  ]);

  const handleSave = useCallback(async () => {
    if (!remark.trim()) {
      push(t("inbounds.remarkRequired"), "error");
      return;
    }
    setSaving(true);
    await new Promise((r) => setTimeout(r, 650));
    const payload = buildPayload();
    if (mode === "edit" && inbound) {
      updateInbound(inbound.id, payload);
      push(t("inbounds.updated"), "success");
    } else {
      addInbound(payload);
      push(mode === "clone" ? t("inbounds.cloned") : t("inbounds.created"), "success");
    }
    setSaving(false);
    onClose();
  }, [remark, buildPayload, mode, inbound, push, t, addInbound, updateInbound, onClose]);

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
    transport, setTransport,
    flow, setFlow,
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
    username, setUsername,
    password, setPassword, regeneratePassword,
    network, setNetwork,
    quicCc, setQuicCc,
    obfsType, setObfsType,
    obfsPassword, setObfsPassword, regenerateObfsPassword,
    upMbps, setUpMbps,
    downMbps, setDownMbps,
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
