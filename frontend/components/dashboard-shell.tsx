"use client";

import { useEffect, useMemo, useState } from "react";
import { AnimatePresence, motion } from "framer-motion";
import {
  Activity,
  BadgeCheck,
  Ban,
  Cpu,
  Database,
  KeyRound,
  Play,
  QrCode,
  RefreshCw,
  RotateCcw,
  Server,
  Square,
  Zap
} from "lucide-react";
import {
  Area,
  AreaChart,
  CartesianGrid,
  ResponsiveContainer,
  Tooltip,
  XAxis,
  YAxis
} from "recharts";

import { api, Inbound, Metrics, Setting, User } from "@/lib/api";
import { cn, formatBytes, formatSpeed } from "@/lib/utils";

type AuthState = "checking" | "authenticated" | "anonymous";

export function DashboardShell({
  authState,
  onAuthState
}: {
  authState: AuthState;
  onAuthState: (state: AuthState) => void;
}) {
  const [metrics, setMetrics] = useState<Metrics | null>(null);
  const [users, setUsers] = useState<User[]>([]);
  const [inbounds, setInbounds] = useState<Inbound[]>([]);
  const [settings, setSettings] = useState<Setting[]>([]);
  const [logs, setLogs] = useState<string[]>([]);
  const [selectedUser, setSelectedUser] = useState<User | null>(null);
  const [links, setLinks] = useState<string[]>([]);
  const [history, setHistory] = useState<{ tick: string; up: number; down: number }[]>([]);

  async function load() {
    const [nextMetrics, nextUsers, nextInbounds, nextLogs, nextSettings] = await Promise.all([
      api.metrics(),
      api.users(),
      api.inbounds(),
      api.logs(),
      api.settings()
    ]);
    setMetrics(nextMetrics);
    setUsers(nextUsers);
    setInbounds(nextInbounds);
    setLogs(nextLogs);
    setSettings(nextSettings);
    setHistory((prev) =>
      [
        ...prev,
        {
          tick: new Date().toLocaleTimeString([], { minute: "2-digit", second: "2-digit" }),
          up: nextMetrics.upload_bps,
          down: nextMetrics.download_bps
        }
      ].slice(-18)
    );
  }

  useEffect(() => {
    if (authState !== "authenticated") return;
    load().catch(() => onAuthState("anonymous"));
    const timer = window.setInterval(() => load().catch(() => undefined), 5000);
    return () => window.clearInterval(timer);
  }, [authState]);

  async function openLinks(user: User) {
    setSelectedUser(user);
    const response = await api.links(user.id);
    setLinks(response.subscription_url ? [...response.links, response.subscription_url] : response.links);
  }

  const inboundById = useMemo(() => new Map(inbounds.map((item) => [item.id, item])), [inbounds]);

  return (
    <main className="min-h-screen bg-void text-zinc-100">
      <div className="fixed inset-0 -z-10 grid-glow opacity-70" />
      <div className="mx-auto flex min-h-screen w-full max-w-[1480px] flex-col px-5 py-5 lg:px-8">
        <Header metrics={metrics} onRefresh={() => load().catch(() => undefined)} />

        <section className="grid gap-4 lg:grid-cols-[1fr_420px]">
          <div className="space-y-4">
            <MetricStrip metrics={metrics} />
            <TrafficChart data={history} />
            <ClientsTable
              users={users}
              inboundById={inboundById}
              onOpenLinks={openLinks}
              onReset={async (user) => {
                await api.resetTraffic(user.id);
                await load();
              }}
              onDisable={async (user) => {
                await api.disableUser(user.id);
                await load();
              }}
            />
          </div>
          <div className="space-y-4">
            <CorePanel metrics={metrics} onAction={async (action) => {
              await api.core(action);
              await load();
            }} />
            <InboundPanel inbounds={inbounds} users={users} />
            <SettingsPanel settings={settings} />
            <LogPanel logs={logs} />
          </div>
        </section>
      </div>
      <ConnectionModal user={selectedUser} links={links} onClose={() => setSelectedUser(null)} />
    </main>
  );
}

function Header({ metrics, onRefresh }: { metrics: Metrics | null; onRefresh: () => void }) {
  return (
    <header className="mb-5 flex min-h-14 items-center justify-between border-b border-line pb-4">
      <div className="flex items-center gap-3">
        <span className="grid size-10 place-items-center border border-glow/40 bg-glow/10 text-glow shadow-neon">
          <Server size={18} />
        </span>
        <div>
          <h1 className="text-lg font-semibold text-white">Sing Grok</h1>
          <p className="text-xs text-zinc-500">CORE {metrics?.core.running ? "ONLINE" : "OFFLINE"}</p>
        </div>
      </div>
      <button
        onClick={onRefresh}
        className="grid size-10 place-items-center border border-line bg-panel text-zinc-300 transition hover:border-glow hover:text-glow"
        title="Refresh"
      >
        <RefreshCw size={17} />
      </button>
    </header>
  );
}

function MetricStrip({ metrics }: { metrics: Metrics | null }) {
  const items = [
    { label: "CPU", value: `${metrics?.cpu_percent.toFixed(1) ?? "0.0"}%`, icon: Cpu },
    { label: "RAM", value: `${metrics?.memory_percent.toFixed(1) ?? "0.0"}%`, icon: Database },
    { label: "UP", value: formatSpeed(metrics?.upload_bps ?? 0), icon: Activity },
    { label: "DOWN", value: formatSpeed(metrics?.download_bps ?? 0), icon: Zap },
    { label: "ACTIVE", value: `${metrics?.active_users ?? 0}`, icon: BadgeCheck }
  ];
  return (
    <div className="grid gap-3 sm:grid-cols-2 xl:grid-cols-5">
      {items.map((item) => (
        <motion.div
          key={item.label}
          layout
          className="min-h-24 border border-line bg-panel p-4 shadow-[inset_0_1px_0_rgba(255,255,255,.04)]"
        >
          <div className="mb-5 flex items-center justify-between text-zinc-500">
            <span className="text-xs">{item.label}</span>
            <item.icon size={16} />
          </div>
          <div className="truncate text-xl font-semibold text-white">{item.value}</div>
        </motion.div>
      ))}
    </div>
  );
}

function TrafficChart({ data }: { data: { tick: string; up: number; down: number }[] }) {
  return (
    <section className="h-[310px] border border-line bg-panel p-4 shadow-neon">
      <div className="mb-4 flex items-center justify-between">
        <h2 className="text-sm font-semibold text-white">TRAFFIC</h2>
        <span className="text-xs text-zinc-500">LIVE</span>
      </div>
      <ResponsiveContainer width="100%" height="85%">
        <AreaChart data={data}>
          <defs>
            <linearGradient id="up" x1="0" x2="0" y1="0" y2="1">
              <stop offset="5%" stopColor="#00e5ff" stopOpacity={0.5} />
              <stop offset="95%" stopColor="#00e5ff" stopOpacity={0} />
            </linearGradient>
            <linearGradient id="down" x1="0" x2="0" y1="0" y2="1">
              <stop offset="5%" stopColor="#b7ff35" stopOpacity={0.42} />
              <stop offset="95%" stopColor="#b7ff35" stopOpacity={0} />
            </linearGradient>
          </defs>
          <CartesianGrid stroke="#171717" vertical={false} />
          <XAxis dataKey="tick" stroke="#52525b" tickLine={false} axisLine={false} fontSize={11} />
          <YAxis stroke="#52525b" tickLine={false} axisLine={false} fontSize={11} tickFormatter={formatBytes} />
          <Tooltip
            contentStyle={{ background: "#050505", border: "1px solid #171717", color: "#fff" }}
            formatter={(value) => formatSpeed(Number(value))}
          />
          <Area type="monotone" dataKey="up" stroke="#00e5ff" fill="url(#up)" strokeWidth={2} />
          <Area type="monotone" dataKey="down" stroke="#b7ff35" fill="url(#down)" strokeWidth={2} />
        </AreaChart>
      </ResponsiveContainer>
    </section>
  );
}

function ClientsTable({
  users,
  inboundById,
  onOpenLinks,
  onReset,
  onDisable
}: {
  users: User[];
  inboundById: Map<number, Inbound>;
  onOpenLinks: (user: User) => void;
  onReset: (user: User) => void;
  onDisable: (user: User) => void;
}) {
  const [expanded, setExpanded] = useState<number | null>(null);
  return (
    <section className="border border-line bg-panel">
      <div className="flex h-12 items-center justify-between border-b border-line px-4">
        <h2 className="text-sm font-semibold text-white">CLIENTS</h2>
        <span className="text-xs text-zinc-500">{users.length} ROWS</span>
      </div>
      <div className="overflow-x-auto">
        <table className="w-full min-w-[760px] border-collapse text-left text-sm">
          <thead className="text-xs text-zinc-500">
            <tr className="border-b border-line">
              <th className="px-4 py-3 font-medium">USER</th>
              <th className="px-4 py-3 font-medium">INBOUND</th>
              <th className="px-4 py-3 font-medium">TRAFFIC</th>
              <th className="px-4 py-3 font-medium">LIMIT</th>
              <th className="px-4 py-3 font-medium">STATUS</th>
              <th className="px-4 py-3 text-right font-medium">ACTIONS</th>
            </tr>
          </thead>
          <tbody>
            {users.map((user) => {
              const inbound = inboundById.get(user.inbound_id);
              const open = expanded === user.id;
              return (
                <tr
                  key={user.id}
                  className={cn("border-b border-line/70 align-top transition", open && "bg-panel2")}
                >
                  <td className="px-4 py-3">
                    <button onClick={() => setExpanded(open ? null : user.id)} className="text-white">
                      {user.username}
                    </button>
                    <AnimatePresence>
                      {open ? (
                        <motion.div
                          layoutId={`client-${user.id}`}
                          initial={{ opacity: 0, height: 0 }}
                          animate={{ opacity: 1, height: "auto" }}
                          exit={{ opacity: 0, height: 0 }}
                          className="mt-3 overflow-hidden text-xs text-zinc-500"
                        >
                          <div className="truncate">UUID {user.uuid}</div>
                          <div>IP LIMIT {user.ip_limit || "NONE"}</div>
                        </motion.div>
                      ) : null}
                    </AnimatePresence>
                  </td>
                  <td className="px-4 py-3 text-zinc-300">{inbound?.tag ?? "missing"}</td>
                  <td className="px-4 py-3 text-zinc-300">{formatBytes(user.used_traffic)}</td>
                  <td className="px-4 py-3 text-zinc-300">{user.total_traffic ? formatBytes(user.total_traffic) : "UNLIM"}</td>
                  <td className="px-4 py-3">
                    <span className={cn("text-xs", user.status === "active" ? "text-pulse" : "text-zinc-500")}>
                      {user.status.toUpperCase()}
                    </span>
                  </td>
                  <td className="px-4 py-3">
                    <div className="flex justify-end gap-2">
                      <IconButton title="Links" onClick={() => onOpenLinks(user)} icon={<QrCode size={15} />} />
                      <IconButton title="Reset traffic" onClick={() => onReset(user)} icon={<RotateCcw size={15} />} />
                      <IconButton title="Disable" onClick={() => onDisable(user)} icon={<Ban size={15} />} />
                    </div>
                  </td>
                </tr>
              );
            })}
          </tbody>
        </table>
      </div>
    </section>
  );
}

function CorePanel({
  metrics,
  onAction
}: {
  metrics: Metrics | null;
  onAction: (action: "start" | "stop" | "restart" | "reload") => void;
}) {
  return (
    <section className="border border-line bg-panel p-4">
      <div className="mb-4 flex items-center justify-between">
        <h2 className="text-sm font-semibold text-white">CORE</h2>
        <span className={cn("text-xs", metrics?.core.running ? "text-pulse" : "text-zinc-500")}>
          {metrics?.core.running ? "RUNNING" : "STOPPED"}
        </span>
      </div>
      <div className="grid grid-cols-4 gap-2">
        <IconButton title="Start" onClick={() => onAction("start")} icon={<Play size={15} />} />
        <IconButton title="Stop" onClick={() => onAction("stop")} icon={<Square size={15} />} />
        <IconButton title="Restart" onClick={() => onAction("restart")} icon={<RefreshCw size={15} />} />
        <IconButton title="Reload" onClick={() => onAction("reload")} icon={<KeyRound size={15} />} />
      </div>
      <p className="mt-4 truncate text-xs text-zinc-500">{metrics?.core.detail || metrics?.core.mode || "unknown"}</p>
    </section>
  );
}

function InboundPanel({ inbounds, users }: { inbounds: Inbound[]; users: User[] }) {
  return (
    <section className="border border-line bg-panel">
      <div className="flex h-12 items-center justify-between border-b border-line px-4">
        <h2 className="text-sm font-semibold text-white">INBOUNDS</h2>
        <span className="text-xs text-zinc-500">{inbounds.length}</span>
      </div>
      <div className="divide-y divide-line">
        {inbounds.map((inbound) => (
          <div key={inbound.id} className="grid grid-cols-[1fr_auto] gap-3 px-4 py-3 text-sm">
            <div className="min-w-0">
              <div className="truncate text-white">{inbound.tag}</div>
              <div className="text-xs text-zinc-500">
                {inbound.protocol.toUpperCase()} :{inbound.port} / {users.filter((user) => user.inbound_id === inbound.id).length} USERS
              </div>
            </div>
            <span className={cn("text-xs", inbound.status === "active" ? "text-pulse" : "text-zinc-500")}>
              {inbound.status.toUpperCase()}
            </span>
          </div>
        ))}
      </div>
    </section>
  );
}

function LogPanel({ logs }: { logs: string[] }) {
  return (
    <section className="h-[360px] border border-line bg-black p-4">
      <div className="mb-3 flex items-center justify-between">
        <h2 className="text-sm font-semibold text-white">LOGS</h2>
        <span className="text-xs text-zinc-500">TAIL</span>
      </div>
      <pre className="h-[294px] overflow-hidden whitespace-pre-wrap text-xs leading-relaxed text-zinc-500">
        {logs.length ? logs.join("\n") : "$ waiting for sing-box output"}
      </pre>
    </section>
  );
}

function SettingsPanel({ settings }: { settings: Setting[] }) {
  const visible = settings.filter((setting) =>
    ["public_host", "clash_api_port", "v2ray_api_port", "log_level"].includes(setting.key)
  );
  return (
    <section className="border border-line bg-panel">
      <div className="flex h-12 items-center justify-between border-b border-line px-4">
        <h2 className="text-sm font-semibold text-white">SETTINGS</h2>
        <span className="text-xs text-zinc-500">{visible.length}</span>
      </div>
      <div className="divide-y divide-line">
        {visible.map((setting) => (
          <div key={setting.key} className="grid grid-cols-[150px_1fr] gap-3 px-4 py-3 text-xs">
            <span className="truncate text-zinc-500">{setting.key}</span>
            <span className="truncate text-zinc-300">{setting.value || "unset"}</span>
          </div>
        ))}
      </div>
    </section>
  );
}

function ConnectionModal({
  user,
  links,
  onClose
}: {
  user: User | null;
  links: string[];
  onClose: () => void;
}) {
  return (
    <AnimatePresence>
      {user ? (
        <motion.div
          className="fixed inset-0 z-50 grid place-items-center bg-black/72 p-5"
          initial={{ opacity: 0 }}
          animate={{ opacity: 1 }}
          exit={{ opacity: 0 }}
        >
          <motion.div
            layoutId={`client-${user.id}`}
            className="w-full max-w-[560px] border border-line bg-panel p-5 shadow-neon"
            initial={{ scale: 0.98 }}
            animate={{ scale: 1 }}
            exit={{ scale: 0.98 }}
          >
            <div className="mb-4 flex items-center justify-between">
              <h2 className="text-sm font-semibold text-white">{user.username}</h2>
              <button onClick={onClose} className="text-xs text-zinc-500 hover:text-white">
                CLOSE
              </button>
            </div>
            <div className="mb-4 grid place-items-center border border-line bg-white p-4">
              <img
                alt=""
                src={`/api/users/${user.id}/qr`}
                className="size-56"
              />
            </div>
            <div className="space-y-2">
              {links.map((link) => (
                <button
                  key={link}
                  onClick={() => navigator.clipboard.writeText(link)}
                  className="w-full truncate border border-line bg-black px-3 py-2 text-left text-xs text-zinc-300 transition hover:border-glow hover:text-glow"
                >
                  {link}
                </button>
              ))}
            </div>
          </motion.div>
        </motion.div>
      ) : null}
    </AnimatePresence>
  );
}

function IconButton({
  title,
  onClick,
  icon
}: {
  title: string;
  onClick: () => void;
  icon: React.ReactNode;
}) {
  return (
    <button
      title={title}
      onClick={onClick}
      className="grid size-9 place-items-center border border-line bg-black text-zinc-400 transition hover:border-glow hover:text-glow"
    >
      {icon}
    </button>
  );
}
