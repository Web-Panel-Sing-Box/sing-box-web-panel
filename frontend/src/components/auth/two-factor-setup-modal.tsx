import { useEffect, useRef, useState } from "react";
import { Check, Copy, Loader2 } from "lucide-react";

import * as api from "@/api";
import { Button } from "@/components/ui/button";
import { Input, Label } from "@/components/ui/input";
import { Modal, ModalBody, ModalFooter, ModalHeader } from "@/components/ui/modal";
import { useToast } from "@/components/ui/toast";
import { useCopyToClipboard } from "@/hooks/useCopyToClipboard";
import { useI18n } from "@/lib/i18n";

type TwoFactorSetupModalProps = {
  open: boolean;
  onClose: () => void;
  onConfirmed: () => void;
};

export function TwoFactorSetupModal({ open, onClose, onConfirmed }: TwoFactorSetupModalProps) {
  const { t } = useI18n();
  const { push } = useToast();
  const { copied, copy } = useCopyToClipboard();
  const [code, setCode] = useState("");
  const [secret, setSecret] = useState("");
  const [loading, setLoading] = useState(false);
  const [confirming, setConfirming] = useState(false);
  const qrKeyRef = useRef(0);

  useEffect(() => {
    if (!open) return;
    setCode("");
    setLoading(true);
    qrKeyRef.current++;
    api.setupTOTP()
      .then((res) => setSecret(res.secret))
      .catch(() => push("Failed to init TOTP", "error"))
      .finally(() => setLoading(false));
  }, [open, push]);

  const close = () => {
    setCode("");
    onClose();
  };

  const confirm = async () => {
    if (code.length !== 6) return;
    setConfirming(true);
    try {
      await api.confirmTOTP({ code });
      onConfirmed();
      push(t("settings.twoFactorEnabled"), "success");
      close();
    } catch (e: any) {
      push(e?.body?.error ?? t("settings.twoFactorInvalidCode"), "error");
    } finally {
      setConfirming(false);
    }
  };

  return (
    <Modal open={open} onClose={close} width="max-w-[420px]">
      <ModalHeader title={t("settings.twoFactorSetupTitle")} onClose={close} />
      <ModalBody className="flex flex-col items-center gap-5">
        {loading ? (
          <div className="flex h-[220px] w-[220px] items-center justify-center rounded-2xl bg-white">
            <Loader2 size={32} className="animate-spin text-ink-tertiary" />
          </div>
        ) : (
          <img
            src={`api/auth/totp/qr?t=${qrKeyRef.current}`}
            alt="TOTP QR"
            className="h-[220px] w-[220px] rounded-2xl border border-white/10 bg-white"
          />
        )}

        {secret && (
          <div className="w-full space-y-1.5">
            <p className="text-xs text-ink-secondary">{t("settings.twoFactorSecret")}</p>
            <div className="flex items-center gap-2 rounded-lg border border-white/10 bg-elevated px-3 py-2">
              <code className="min-w-0 flex-1 break-all font-mono text-xs text-ink-primary">
                {secret}
              </code>
              <button
                type="button"
                onClick={() => copy(secret)}
                className="shrink-0 rounded-md p-1.5 text-ink-secondary transition-colors duration-150 hover:bg-hover hover:text-ink-primary"
                aria-label="Copy secret"
              >
                {copied ? <Check size={14} /> : <Copy size={14} />}
              </button>
            </div>
          </div>
        )}

        <div className="w-full">
          <Label htmlFor="twofa-code">{t("settings.twoFactorEnterCode")}</Label>
          <Input
            id="twofa-code"
            value={code}
            onChange={(e) => setCode(e.target.value.replace(/[^0-9]/g, "").slice(0, 6))}
            onKeyDown={(e) => {
              if (e.key === "Enter" && code.length === 6 && !confirming) confirm();
            }}
            inputMode="numeric"
            autoComplete="one-time-code"
            maxLength={6}
            mono
            autoFocus
          />
        </div>
      </ModalBody>
      <ModalFooter>
        <Button variant="primary" onClick={confirm} disabled={code.length !== 6 || confirming || loading}>
          {confirming ? <Loader2 size={14} className="mr-1.5 animate-spin" /> : null}
          {t("settings.twoFactorEnable")}
        </Button>
      </ModalFooter>
    </Modal>
  );
}
