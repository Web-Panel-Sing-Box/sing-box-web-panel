"use client";

import { FormEvent, useState } from "react";
import { motion } from "framer-motion";
import { LockKeyhole, Terminal } from "lucide-react";

import { api } from "@/lib/api";

export function LoginPanel({ onAuthenticated }: { onAuthenticated: () => void }) {
  const [username, setUsername] = useState("admin");
  const [password, setPassword] = useState("");
  const [error, setError] = useState("");
  const [pending, setPending] = useState(false);

  async function submit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    setPending(true);
    setError("");
    try {
      await api.login(username, password);
      onAuthenticated();
    } catch (err) {
      setError(err instanceof Error ? err.message : "Login failed");
    } finally {
      setPending(false);
    }
  }

  return (
    <main className="grid min-h-screen place-items-center bg-void grid-glow px-6">
      <motion.form
        initial={{ opacity: 0, y: 18 }}
        animate={{ opacity: 1, y: 0 }}
        transition={{ duration: 0.38, ease: [0.22, 1, 0.36, 1] }}
        onSubmit={submit}
        className="w-full max-w-[420px] border border-line bg-panel/95 p-6 shadow-neon"
      >
        <div className="mb-8 flex items-center gap-3">
          <span className="grid size-10 place-items-center border border-glow/40 bg-glow/10 text-glow">
            <Terminal size={19} />
          </span>
          <div>
            <h1 className="text-lg font-semibold tracking-normal text-white">Sing Grok</h1>
            <p className="text-xs text-zinc-500">LOCALHOST CONTROL PLANE</p>
          </div>
        </div>

        <label className="mb-2 block text-xs text-zinc-500" htmlFor="username">
          USER
        </label>
        <input
          id="username"
          value={username}
          onChange={(event) => setUsername(event.target.value)}
          className="mb-4 h-11 w-full border border-line bg-black px-3 text-sm text-white outline-none transition focus:border-glow"
        />

        <label className="mb-2 block text-xs text-zinc-500" htmlFor="password">
          SECRET
        </label>
        <div className="mb-4 flex h-11 items-center border border-line bg-black px-3 focus-within:border-glow">
          <LockKeyhole size={16} className="mr-2 text-zinc-600" />
          <input
            id="password"
            value={password}
            onChange={(event) => setPassword(event.target.value)}
            type="password"
            className="h-full min-w-0 flex-1 bg-transparent text-sm text-white outline-none"
          />
        </div>

        {error ? <p className="mb-4 text-xs text-red-400">{error}</p> : null}

        <button
          disabled={pending}
          className="h-11 w-full border border-glow/50 bg-glow/10 text-sm font-semibold text-glow transition hover:bg-glow/15 disabled:cursor-wait disabled:opacity-60"
        >
          {pending ? "AUTH..." : "ENTER"}
        </button>
      </motion.form>
    </main>
  );
}
