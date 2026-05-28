
import {
  createContext,
  useCallback,
  useContext,
  useEffect,
  useMemo,
  useRef,
  useState
} from "react";

import { buildSeedClients, type Client, type ClientStatus } from "./clients";
import { SEED_INBOUNDS, type Inbound } from "./inbounds";
import { buildLog, nextLog, seedLogs, type LogEntry, type LogLevel } from "./logs";
import { initialMetrics, seedTrafficHistory, tickMetrics, type Metrics, type TrafficPoint } from "./metrics";

type StoreState = {
  metrics: Metrics;
  history: TrafficPoint[];
  inbounds: Inbound[];
  clients: Client[];
  logs: LogEntry[];
  paused: boolean;
};

type StoreActions = {
  toggleInbound: (id: string) => void;
  addInbound: (input: Omit<Inbound, "id" | "createdAt" | "clientCount" | "enabled">) => void;
  updateInbound: (id: string, patch: Partial<Omit<Inbound, "id" | "createdAt" | "clientCount">>) => void;
  removeInbound: (id: string) => void;
  cloneInbound: (id: string) => void;
  addClient: (input: { name: string; inboundId: string; totalQuota: number; expiry: string; startAfterFirstUse?: boolean }) => void;
  updateClient: (id: string, patch: Partial<Client>) => void;
  removeClient: (id: string) => void;
  resetClientTraffic: (id: string) => void;
  setClientStatus: (id: string, status: ClientStatus) => void;
  setPaused: (v: boolean) => void;
  appendLog: (level: LogLevel, message: string) => void;
  setCoreRunning: (v: boolean) => void;
};

const StoreContext = createContext<(StoreState & StoreActions) | null>(null);

export function MockStoreProvider({ children }: { children: React.ReactNode }) {
  const [metrics, setMetrics] = useState<Metrics>(() => initialMetrics());
  const [history, setHistory] = useState<TrafficPoint[]>(() => seedTrafficHistory());
  const [inbounds, setInbounds] = useState<Inbound[]>(SEED_INBOUNDS);
  const [clients, setClients] = useState<Client[]>(() => buildSeedClients());
  const [logs, setLogs] = useState<LogEntry[]>(() => seedLogs());
  const [paused, setPaused] = useState(false);
  const tickCount = useRef(0);

  useEffect(() => {
    if (paused) return;
    const id = window.setInterval(() => {
      tickCount.current += 1;
      setMetrics((prev) => tickMetrics({ ...prev, inboundsActive: inbounds.filter((i) => i.enabled).length, totalUsers: clients.length }));
      setHistory((prev) => {
        const last = prev[prev.length - 1];
        const next: TrafficPoint = {
          t: Date.now(),
          up: Math.max(800_000, (last?.up ?? 3_000_000) + (Math.random() - 0.5) * 800_000),
          down: Math.max(1_200_000, (last?.down ?? 6_000_000) + (Math.random() - 0.5) * 1_400_000)
        };
        const trimmed = [...prev, next];
        return trimmed.length > 60 ? trimmed.slice(trimmed.length - 60) : trimmed;
      });
      if (tickCount.current % 3 === 0) {
        setLogs((prev) => {
          const out = [...prev, nextLog()];
          return out.length > 200 ? out.slice(out.length - 200) : out;
        });
      }
    }, 1000);
    return () => window.clearInterval(id);
  }, [paused, inbounds, clients.length]);

  const toggleInbound = useCallback((id: string) => {
    setInbounds((prev) => prev.map((ib) => (ib.id === id ? { ...ib, enabled: !ib.enabled } : ib)));
  }, []);

  const addInbound = useCallback<StoreActions["addInbound"]>((input) => {
    setInbounds((prev) => [
      {
        ...input,
        id: `ib_${String(prev.length + 1).padStart(2, "0")}`,
        enabled: true,
        clientCount: 0,
        createdAt: new Date().toISOString()
      },
      ...prev
    ]);
    setLogs((prev) => [
      ...prev,
      buildLog("info", `inbound ${input.remark}:${input.port} created`)
    ]);
  }, []);

  const updateInbound = useCallback<StoreActions["updateInbound"]>((id, patch) => {
    setInbounds((prev) => prev.map((i) => (i.id === id ? { ...i, ...patch } : i)));
  }, []);

  const removeInbound = useCallback((id: string) => {
    setInbounds((prev) => prev.filter((i) => i.id !== id));
  }, []);

  const cloneInbound = useCallback((id: string) => {
    setInbounds((prev) => {
      const found = prev.find((i) => i.id === id);
      if (!found) return prev;
      const clone: Inbound = {
        ...found,
        id: `ib_${String(prev.length + 1).padStart(2, "0")}`,
        remark: `${found.remark}-copy`,
        port: Math.floor(10000 + Math.random() * 50000),
        createdAt: new Date().toISOString(),
        clientCount: 0
      };
      return [clone, ...prev];
    });
  }, []);

  const addClient = useCallback<StoreActions["addClient"]>((input) => {
    setClients((prev) => {
      const id = `cl_${Math.random().toString(36).slice(2, 8)}`;
      const subToken = Math.random().toString(36).slice(2, 10);
      const client: Client = {
        id,
        name: input.name,
        uuid: typeof crypto !== "undefined" && "randomUUID" in crypto ? crypto.randomUUID() : "00000000-0000-4000-8000-000000000000",
        inboundId: input.inboundId,
        usedDown: 0,
        usedUp: 0,
        totalQuota: input.totalQuota,
        expiry: input.expiry,
        status: "active",
        subscription: `https://panel.example/sub/${id}_${subToken}`,
        startAfterFirstUse: input.startAfterFirstUse ?? false
      };
      return [client, ...prev];
    });
    setLogs((prev) => [
      ...prev,
      buildLog("info", `client ${input.name} provisioned on inbound ${input.inboundId}`)
    ]);
  }, []);

  const updateClient = useCallback<StoreActions["updateClient"]>((id, patch) => {
    setClients((prev) => prev.map((c) => (c.id === id ? { ...c, ...patch } : c)));
  }, []);

  const removeClient = useCallback((id: string) => {
    setClients((prev) => prev.filter((c) => c.id !== id));
  }, []);

  const resetClientTraffic = useCallback((id: string) => {
    setClients((prev) => prev.map((c) => (c.id === id ? { ...c, usedDown: 0, usedUp: 0 } : c)));
  }, []);

  const setClientStatus = useCallback<StoreActions["setClientStatus"]>((id, status) => {
    setClients((prev) => prev.map((c) => (c.id === id ? { ...c, status } : c)));
  }, []);

  const appendLog = useCallback<StoreActions["appendLog"]>((level, message) => {
    setLogs((prev) => [...prev, buildLog(level, message)]);
  }, []);

  const setCoreRunning = useCallback((v: boolean) => {
    setMetrics((prev) => ({ ...prev, coreRunning: v }));
  }, []);

  const value = useMemo<StoreState & StoreActions>(
    () => ({
      metrics,
      history,
      inbounds,
      clients,
      logs,
      paused,
      toggleInbound,
      addInbound,
      updateInbound,
      removeInbound,
      cloneInbound,
      addClient,
      updateClient,
      removeClient,
      resetClientTraffic,
      setClientStatus,
      setPaused,
      appendLog,
      setCoreRunning
    }),
    [
      metrics,
      history,
      inbounds,
      clients,
      logs,
      paused,
      toggleInbound,
      addInbound,
      updateInbound,
      removeInbound,
      cloneInbound,
      addClient,
      updateClient,
      removeClient,
      resetClientTraffic,
      setClientStatus,
      appendLog,
      setCoreRunning
    ]
  );

  return <StoreContext.Provider value={value}>{children}</StoreContext.Provider>;
}

export function useStore() {
  const ctx = useContext(StoreContext);
  if (!ctx) throw new Error("useStore must be used inside <MockStoreProvider />");
  return ctx;
}
