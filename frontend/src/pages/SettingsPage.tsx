
import { useState } from "react";

import { Button } from "@/components/ui/button";
import { Card, CardHeader, CardTitle } from "@/components/ui/card";
import { Input, Label } from "@/components/ui/input";
import { Select } from "@/components/ui/select";
import { Toggle } from "@/components/ui/toggle";
import { useToast } from "@/components/ui/toast";

const LANGUAGE_OPTIONS = [
  { value: "en", label: "English" },
  { value: "ru", label: "Russian" },
  { value: "ja", label: "Japanese" }
];

const LOG_LEVELS = [
  { value: "debug", label: "Debug" },
  { value: "info", label: "Info" },
  { value: "warn", label: "Warn" },
  { value: "error", label: "Error" }
];

export function SettingsPage() {
  const { push } = useToast();
  const [panelName, setPanelName] = useState("Sing Grok");
  const [language, setLanguage] = useState("en");
  const [twoFactor, setTwoFactor] = useState(false);
  const [binaryPath, setBinaryPath] = useState("/usr/local/bin/sing-box");
  const [logLevel, setLogLevel] = useState("info");
  const [publicHost, setPublicHost] = useState("panel.example");
  const [ttl, setTtl] = useState("72h");

  const save = (section: string) => () => {
    push(`${section} settings saved`);
  };

  return (
    <div className="mx-auto flex max-w-[920px] flex-col gap-6">
      <div>
        <h2 className="text-2xl font-semibold text-ink-primary">Settings</h2>
        <p className="mt-1 text-sm text-ink-tertiary">All values below are local to this mock build</p>
      </div>

      <Section title="General" onSave={save("General")}>
        <Row label="Panel name" hint="Shown in the browser tab and login screen">
          <Input value={panelName} onChange={(e) => setPanelName(e.target.value)} />
        </Row>
        <Row label="Language" hint="UI language for new sessions">
          <Select value={language} options={LANGUAGE_OPTIONS} onChange={setLanguage} />
        </Row>
      </Section>

      <Section title="Security" onSave={save("Security")}>
        <Row label="Admin username" hint="Read-only in this build">
          <Input value="admin" readOnly mono />
        </Row>
        <Row label="Two-factor auth" hint="TOTP via authenticator app">
          <Toggle checked={twoFactor} onChange={setTwoFactor} />
        </Row>
        <Row label="Change password" hint="Mock-disabled in this build">
          <Button disabled variant="secondary">
            Change password
          </Button>
        </Row>
      </Section>

      <Section title="Sing-box" onSave={save("Sing-box")}>
        <Row label="Binary path" hint="Path used by ProcessManager">
          <Input value={binaryPath} onChange={(e) => setBinaryPath(e.target.value)} mono />
        </Row>
        <Row label="Log level" hint="Controls verbosity of sing-box stdout">
          <Select value={logLevel} options={LOG_LEVELS} onChange={setLogLevel} />
        </Row>
        <Row label="Clash API port" hint="Bound to 127.0.0.1">
          <Input value="9090" readOnly mono />
        </Row>
        <Row label="V2Ray API port" hint="Bound to 127.0.0.1">
          <Input value="9091" readOnly mono />
        </Row>
      </Section>

      <Section title="Subscriptions" onSave={save("Subscriptions")}>
        <Row label="Public base URL" hint="Origin used when generating client links">
          <Input value={publicHost} onChange={(e) => setPublicHost(e.target.value)} mono />
        </Row>
        <Row label="Token TTL" hint="Lifetime of subscription tokens">
          <Input value={ttl} onChange={(e) => setTtl(e.target.value)} mono />
        </Row>
      </Section>
    </div>
  );
}

function Section({ title, onSave, children }: { title: string; onSave: () => void; children: React.ReactNode }) {
  return (
    <Card>
      <CardHeader>
        <CardTitle>{title}</CardTitle>
        <Button variant="primary" size="sm" onClick={onSave}>
          Save
        </Button>
      </CardHeader>
      <div className="space-y-4">{children}</div>
    </Card>
  );
}

function Row({ label, hint, children }: { label: string; hint?: string; children: React.ReactNode }) {
  return (
    <div className="grid grid-cols-1 items-start gap-3 sm:grid-cols-[1fr_320px]">
      <div className="space-y-0.5">
        <Label>{label}</Label>
        {hint ? <p className="text-xs text-ink-tertiary">{hint}</p> : null}
      </div>
      <div>{children}</div>
    </div>
  );
}
