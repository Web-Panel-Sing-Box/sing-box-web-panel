import { useEffect, useState } from "react";
import { Check, Copy } from "lucide-react";

import * as api from "@/api";
import { Button } from "@/components/ui/button";
import { Input, Label } from "@/components/ui/input";
import { Modal, ModalBody, ModalFooter, ModalHeader } from "@/components/ui/modal";
import { QrCode } from "@/components/ui/qr-code";
import { useToast } from "@/components/ui/toast";
import { useCopyToClipboard } from "@/hooks/useCopyToClipboard";
import { useI18n } from "@/lib/i18n";

type TwoFactorSetupModalProps = {
  open: boolean;
  onClose: () => void;
  onConfirmed: () => void;
};

type Step = "scan" | "recovery";

function parseSecret(qrUri: string): string {
  try {
    return new URLSearchParams(new URL(qrUri).search).get("secret") ?? "";
  } catch {
    return "";
  }
}

export function TwoFactorSetupModal({
  open,
  onClose,
  onConfirmed,
}: TwoFactorSetupModalProps) {
  const { t } = useI18n();
  const { push } = useToast();
  const { copied, copy } = useCopyToClipboard();
  const { copied: codesCopied, copy: copyCodes } = useCopyToClipboard();

  const [step, setStep] = useState<Step>("scan");
  const [qrUri, setQrUri] = useState("");
  const [secret, setSecret] = useState("");
  const [code, setCode] = useState("");
  const [recoveryCodes, setRecoveryCodes] = useState<string[]>([]);
  const [loading, setLoading] = useState(false);
  const [submitting, setSubmitting] = useState(false);

  useEffect(() => {
    if (!open) {
      setStep("scan");
      setQrUri("");
      setSecret("");
      setCode("");
      setRecoveryCodes([]);
      setLoading(false);
      setSubmitting(false);
      return;
    }

    let cancelled = false;
    setLoading(true);
    api
      .setupTOTP()
      .then((res) => {
        if (cancelled) return;
        setQrUri(res.qr_uri);
        setSecret(parseSecret(res.qr_uri));
      })
      .catch(() => {
        if (cancelled) return;
        push(t("settings.twoFactorSetupFailed"), "error");
        onClose();
      })
      .finally(() => {
        if (!cancelled) setLoading(false);
      });

    return () => {
      cancelled = true;
    };
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [open]);

  const confirm = async () => {
    if (code.length !== 6 || submitting) return;
    setSubmitting(true);
    try {
      const res = await api.confirmTOTP({ code });
      setRecoveryCodes(res.recovery_codes);
      setStep("recovery");
    } catch {
      push(t("settings.twoFactorInvalidCode"), "error");
    } finally {
      setSubmitting(false);
    }
  };

  const finish = () => {
    onConfirmed();
    push(t("settings.twoFactorEnabled"), "success");
    onClose();
  };

  return (
    <Modal open={open} onClose={onClose} width="max-w-[420px]">
      {step === "scan" ? (
        <>
          <ModalHeader
            title={t("settings.twoFactorSetupTitle")}
            subtitle={t("settings.twoFactorScanHint")}
            onClose={onClose}
          />
          <ModalBody className="flex flex-col items-center gap-5">
            {loading || !qrUri ? (
              <div className="grid h-[212px] w-[212px] place-items-center rounded-2xl bg-white/5 text-xs text-ink-tertiary">
                ...
              </div>
            ) : (
              <QrCode payload={qrUri} />
            )}

            <div className="w-full space-y-1.5">
              <p className="text-xs text-ink-secondary">{t("settings.twoFactorSecret")}</p>
              <div className="flex items-center gap-2 rounded-lg border border-white/10 bg-elevated px-3 py-2">
                <code className="min-w-0 flex-1 break-all font-mono text-xs text-ink-primary">
                  {secret || "..."}
                </code>
                <button
                  type="button"
                  onClick={() => copy(secret)}
                  disabled={!secret}
                  className="shrink-0 rounded-md p-1.5 text-ink-secondary transition-colors duration-150 hover:bg-hover hover:text-ink-primary disabled:opacity-40"
                  aria-label="Copy secret"
                >
                  {copied ? <Check size={14} /> : <Copy size={14} />}
                </button>
              </div>
            </div>

            <div className="w-full">
              <Label htmlFor="twofa-code">{t("settings.twoFactorEnterCode")}</Label>
              <Input
                id="twofa-code"
                value={code}
                onChange={(e) => setCode(e.target.value.replace(/[^0-9]/g, "").slice(0, 6))}
                onKeyDown={(e) => {
                  if (e.key === "Enter") confirm();
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
            <Button
              variant="primary"
              onClick={confirm}
              loading={submitting}
              disabled={code.length !== 6 || loading || submitting}
            >
              {t("settings.twoFactorEnable")}
            </Button>
          </ModalFooter>
        </>
      ) : (
        <>
          <ModalHeader
            title={t("settings.twoFactorRecoveryTitle")}
            subtitle={t("settings.twoFactorRecoveryHint")}
            onClose={onClose}
          />
          <ModalBody className="flex flex-col gap-4">
            <div className="grid grid-cols-2 gap-2 rounded-lg border border-white/10 bg-elevated p-3">
              {recoveryCodes.map((rc) => (
                <code key={rc} className="text-center font-mono text-sm text-ink-primary">
                  {rc}
                </code>
              ))}
            </div>
            <Button
              variant="secondary"
              onClick={() => copyCodes(recoveryCodes.join("\n"))}
              className="w-full"
            >
              {codesCopied ? <Check size={14} /> : <Copy size={14} />}
              {t("settings.twoFactorCopyCodes")}
            </Button>
          </ModalBody>
          <ModalFooter>
            <Button variant="primary" onClick={finish}>
              {t("settings.twoFactorRecoveryDone")}
            </Button>
          </ModalFooter>
        </>
      )}
    </Modal>
  );
}
