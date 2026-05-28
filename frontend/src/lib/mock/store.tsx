
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
  removeInbound: (id: string) => void;
  cloneInbound: (id: string) => void;
  updateClient: (id: string, patch: Partial<Client>) => void;
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

  const updateClient = useCallback<StoreActions["updateClient"]>((id, patch) => {
    setClients((prev) => prev.map((c) => (c.id === id ? { ...c, ...patch } : c)));
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
      removeInbound,
      cloneInbound,
      updateClient,
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
      removeInbound,
      cloneInbound,
      updateClient,
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
