import { useCallback, useEffect, useMemo, useState } from "react";
import {
  Copy,
  KeyRound,
  Plus,
  RefreshCw,
  RotateCw,
  ServerCog,
  Trash2,
} from "lucide-react";

import {
  createAPIToken,
  createNode,
  deleteAPIToken,
  deleteNode,
  listAPITokens,
  listNodes,
  probeNode,
  setAPITokenEnabled,
  syncNode,
  toggleNode,
  type APITokenDTO,
  type NodeDTO,
} from "@/api";
import { Button } from "@/components/ui/button";
import { Input, Label } from "@/components/ui/input";
import { cn } from "@/lib/utils";

type NodeForm = {
  name: string;
  scheme: "http" | "https";
  address: string;
  port: string;
  basePath: string;
  apiToken: string;
  allowPrivateAddress: boolean;
};

const emptyForm: NodeForm = {
  name: "",
  scheme: "https",
  address: "",
  port: "443",
  basePath: "",
  apiToken: "",
  allowPrivateAddress: false,
};

export function NodesPage() {
  const [nodes, setNodes] = useState<NodeDTO[]>([]);
  const [tokens, setTokens] = useState<APITokenDTO[]>([]);
  const [form, setForm] = useState<NodeForm>(emptyForm);
  const [tokenName, setTokenName] = useState("node");
  const [createdToken, setCreatedToken] = useState("");
  const [busy, setBusy] = useState<string | null>(null);
  const [error, setError] = useState("");

  const load = useCallback(async () => {
    const [nodeList, tokenList] = await Promise.all([
      listNodes(),
      listAPITokens(),
    ]);
    setNodes(nodeList);
    setTokens(tokenList);
  }, []);

  useEffect(() => {
    void load().catch(() => undefined);
  }, [load]);

  const counts = useMemo(
    () => ({
      online: nodes.filter((node) => node.status === "online").length,
      enabled: nodes.filter((node) => node.enabled).length,
      tokens: tokens.filter((token) => token.enabled).length,
    }),
    [nodes, tokens],
  );

  const submitNode = async () => {
    setError("");
    setBusy("create-node");
    try {
      await createNode({
        name: form.name,
        scheme: form.scheme,
        address: form.address,
        port: Number(form.port || "443"),
        basePath: form.basePath,
        apiToken: form.apiToken,
        enabled: true,
        allowPrivateAddress: form.allowPrivateAddress,
      });
      setForm(emptyForm);
      await load();
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to create node");
    } finally {
      setBusy(null);
    }
  };

  const submitToken = async () => {
    setError("");
    setBusy("create-token");
    try {
      const token = await createAPIToken({ name: tokenName, scopes: "node" });
      setCreatedToken(token.token);
      await load();
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to create token");
    } finally {
      setBusy(null);
    }
  };

  const runNodeAction = async (
    id: string,
    action: "probe" | "sync" | "toggle" | "delete",
  ) => {
    setBusy(`${action}-${id}`);
    setError("");
    try {
      if (action === "probe") await probeNode(id);
      if (action === "sync") await syncNode(id);
      if (action === "toggle") await toggleNode(id);
      if (action === "delete") await deleteNode(id);
      await load();
    } catch (err) {
      setError(err instanceof Error ? err.message : "Node action failed");
    } finally {
      setBusy(null);
    }
  };

  const runTokenAction = async (
    id: string,
    action: "enable" | "disable" | "delete",
  ) => {
    setBusy(`${action}-token-${id}`);
    setError("");
    try {
      if (action === "delete") await deleteAPIToken(id);
      if (action === "enable") await setAPITokenEnabled(id, true);
      if (action === "disable") await setAPITokenEnabled(id, false);
      await load();
    } catch (err) {
      setError(err instanceof Error ? err.message : "Token action failed");
    } finally {
      setBusy(null);
    }
  };

  return (
    <div className="mx-auto flex max-w-[1240px] flex-col gap-6">
      <div className="flex flex-wrap items-end justify-between gap-4">
        <div>
          <h2 className="text-2xl font-semibold text-ink-primary">Nodes</h2>
          <p className="mt-1 text-sm text-ink-tertiary">
            Remote Shilka panels, heartbeat, API tokens, and sync cache.
          </p>
        </div>
        <div className="flex gap-2 text-xs text-ink-tertiary">
          <span className="rounded-md border border-subtle px-2 py-1">
            {counts.online}/{nodes.length} online
          </span>
          <span className="rounded-md border border-subtle px-2 py-1">
            {counts.enabled} enabled
          </span>
          <span className="rounded-md border border-subtle px-2 py-1">
            {counts.tokens} tokens
          </span>
        </div>
      </div>

      {error ? (
        <div className="rounded-lg border border-danger/40 bg-danger/10 px-3 py-2 text-sm text-danger">
          {error}
        </div>
      ) : null}

      <section className="grid gap-4 lg:grid-cols-[1.15fr_0.85fr]">
        <div className="rounded-lg border border-subtle bg-surface p-4">
          <div className="mb-4 flex items-center gap-2 text-sm font-medium text-ink-primary">
            <ServerCog size={16} />
            Register node
          </div>
          <div className="grid gap-3 sm:grid-cols-2">
            <div>
              <Label>Name</Label>
              <Input
                value={form.name}
                onChange={(event) =>
                  setForm((prev) => ({ ...prev, name: event.target.value }))
                }
                placeholder="tokyo-1"
              />
            </div>
            <div>
              <Label>Address</Label>
              <Input
                value={form.address}
                onChange={(event) =>
                  setForm((prev) => ({ ...prev, address: event.target.value }))
                }
                placeholder="panel.example.com"
              />
            </div>
            <div>
              <Label>Scheme</Label>
              <select
                value={form.scheme}
                onChange={(event) =>
                  setForm((prev) => ({
                    ...prev,
                    scheme: event.target.value as "http" | "https",
                  }))
                }
                className="h-10 w-full rounded-lg border border-white/10 bg-elevated px-3 text-sm text-ink-primary outline-none"
              >
                <option value="https">HTTPS</option>
                <option value="http">HTTP</option>
              </select>
            </div>
            <div>
              <Label>Port</Label>
              <Input
                value={form.port}
                onChange={(event) =>
                  setForm((prev) => ({ ...prev, port: event.target.value }))
                }
                inputMode="numeric"
                placeholder="443"
              />
            </div>
            <div>
              <Label>Base path</Label>
              <Input
                value={form.basePath}
                onChange={(event) =>
                  setForm((prev) => ({ ...prev, basePath: event.target.value }))
                }
                placeholder="/panel"
              />
            </div>
            <div>
              <Label>API token</Label>
              <Input
                value={form.apiToken}
                onChange={(event) =>
                  setForm((prev) => ({ ...prev, apiToken: event.target.value }))
                }
                mono
                placeholder="remote bearer token"
              />
            </div>
          </div>
          <label className="mt-3 flex items-center gap-2 text-xs text-ink-secondary">
            <input
              type="checkbox"
              checked={form.allowPrivateAddress}
              onChange={(event) =>
                setForm((prev) => ({
                  ...prev,
                  allowPrivateAddress: event.target.checked,
                }))
              }
              className="size-4"
            />
            Allow private or loopback address
          </label>
          <div className="mt-4 flex justify-end">
            <Button
              variant="white"
              onClick={submitNode}
              loading={busy === "create-node"}
              disabled={!form.name || !form.address || !form.apiToken}
            >
              <Plus size={16} />
              Add node
            </Button>
          </div>
        </div>

        <div className="rounded-lg border border-subtle bg-surface p-4">
          <div className="mb-4 flex items-center gap-2 text-sm font-medium text-ink-primary">
            <KeyRound size={16} />
            Local API tokens
          </div>
          <div className="flex gap-2">
            <Input
              value={tokenName}
              onChange={(event) => setTokenName(event.target.value)}
              placeholder="node"
            />
            <Button
              variant="white"
              onClick={submitToken}
              loading={busy === "create-token"}
            >
              Create
            </Button>
          </div>
          {createdToken ? (
            <div className="mt-3 rounded-lg border border-brand/30 bg-brand/10 p-3">
              <div className="text-xs text-ink-tertiary">Token shown once</div>
              <div className="mt-1 flex items-center gap-2">
                <code className="min-w-0 flex-1 truncate font-mono text-xs text-ink-primary">
                  {createdToken}
                </code>
                <Button
                  size="sm"
                  variant="ghost"
                  onClick={() => navigator.clipboard?.writeText(createdToken)}
                >
                  <Copy size={14} />
                </Button>
              </div>
            </div>
          ) : null}
          <div className="mt-4 max-h-64 overflow-auto">
            {tokens.map((token) => (
              <div
                key={token.id}
                className="flex items-center justify-between border-t border-subtle py-2 text-sm"
              >
                <div className="min-w-0">
                  <div className="truncate text-ink-primary">{token.name}</div>
                  <div className="font-mono text-xs text-ink-tertiary">
                    {token.tokenPrefix}... · {token.scopes}
                  </div>
                </div>
                <div className="flex items-center gap-1">
                  <Button
                    size="sm"
                    variant="ghost"
                    onClick={() =>
                      runTokenAction(
                        token.id,
                        token.enabled ? "disable" : "enable",
                      )
                    }
                  >
                    {token.enabled ? "On" : "Off"}
                  </Button>
                  <Button
                    size="sm"
                    variant="danger"
                    onClick={() => runTokenAction(token.id, "delete")}
                  >
                    <Trash2 size={14} />
                  </Button>
                </div>
              </div>
            ))}
          </div>
        </div>
      </section>

      <section className="overflow-x-auto rounded-lg border border-subtle bg-surface">
        <div className="grid min-w-[920px] grid-cols-[1.5fr_0.8fr_0.8fr_1fr_0.9fr] gap-3 border-b border-subtle px-4 py-3 text-xs uppercase tracking-[0.18em] text-ink-tertiary">
          <span>Node</span>
          <span>Status</span>
          <span>Latency</span>
          <span>Version</span>
          <span className="text-right">Actions</span>
        </div>
        <div className="max-h-[calc(100dvh-420px)] min-h-40 min-w-[920px] overflow-auto">
          {nodes.map((node) => (
            <div
              key={node.id}
              className="grid grid-cols-[1.5fr_0.8fr_0.8fr_1fr_0.9fr] items-center gap-3 border-b border-subtle px-4 py-3 text-sm"
            >
              <div className="min-w-0">
                <div className="truncate text-ink-primary">{node.name}</div>
                <div className="truncate font-mono text-xs text-ink-tertiary">
                  {node.scheme}://{node.address}:{node.port}
                  {node.basePath}
                </div>
              </div>
              <span
                className={cn(
                  "w-fit rounded-md px-2 py-1 text-xs",
                  node.status === "online" && "bg-brand/10 text-brand",
                  node.status === "offline" && "bg-danger/10 text-danger",
                  node.status === "unknown" && "bg-white/5 text-ink-tertiary",
                )}
              >
                {node.status}
              </span>
              <span className="font-mono text-xs text-ink-secondary">
                {node.latencyMs ? `${node.latencyMs} ms` : "—"}
              </span>
              <span className="truncate text-xs text-ink-secondary">
                {node.coreVersion || node.panelVersion || "—"}
              </span>
              <div className="flex justify-end gap-1">
                <Button
                  size="sm"
                  variant="ghost"
                  title="Probe"
                  onClick={() => runNodeAction(node.id, "probe")}
                >
                  <RefreshCw size={14} />
                </Button>
                <Button
                  size="sm"
                  variant="ghost"
                  title="Sync"
                  onClick={() => runNodeAction(node.id, "sync")}
                >
                  <RotateCw size={14} />
                </Button>
                <Button
                  size="sm"
                  variant="ghost"
                  onClick={() => runNodeAction(node.id, "toggle")}
                >
                  {node.enabled ? "On" : "Off"}
                </Button>
                <Button
                  size="sm"
                  variant="danger"
                  title="Delete"
                  onClick={() => runNodeAction(node.id, "delete")}
                >
                  <Trash2 size={14} />
                </Button>
              </div>
            </div>
          ))}
          {nodes.length === 0 ? (
            <div className="grid min-h-40 place-items-center text-sm text-ink-tertiary">
              No nodes registered.
            </div>
          ) : null}
        </div>
      </section>
    </div>
  );
}
