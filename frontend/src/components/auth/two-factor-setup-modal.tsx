import { useState } from "react";
import { Check, Copy } from "lucide-react";

import { Button } from "@/components/ui/button";
import { FakeQrCode } from "@/components/ui/fake-qr";
import { Input, Label } from "@/components/ui/input";
import { Modal, ModalBody, ModalFooter, ModalHeader } from "@/components/ui/modal";
import { useToast } from "@/components/ui/toast";
import { useCopyToClipboard } from "@/hooks/useCopyToClipboard";
import { buildOtpAuthUri, TWO_FACTOR_SECRET } from "@/lib/auth";
import { useI18n } from "@/lib/i18n";

type TwoFactorSetupModalProps = {
  open: boolean;
  onClose: () => void;
  onConfirmed: () => void;
};

const VALID_CODE = "123456";

export function TwoFactorSetupModal({ open, onClose, onConfirmed }: TwoFactorSetupModalProps) {
  const { t } = useI18n();
  const { push } = useToast();
  const { copied, copy } = useCopyToClipboard();
  const [code, setCode] = useState("");

  const close = () => {
    setCode("");
    onClose();
  };

  const confirm = () => {
    if (code !== VALID_CODE) {
      push(t("settings.twoFactorInvalidCode"), "error");
      return;
    }
    onConfirmed();
    push(t("settings.twoFactorEnabled"), "success");
    close();
  };

  return (
    <Modal open={open} onClose={close} width="max-w-[420px]">
      <ModalHeader title={t("settings.twoFactorSetupTitle")} onClose={close} />
      <ModalBody className="flex flex-col items-center gap-5">
        <FakeQrCode payload={buildOtpAuthUri()} unit={6} />

        <div className="w-full space-y-1.5">
          <p className="text-xs text-ink-secondary">{t("settings.twoFactorSecret")}</p>
          <div className="flex items-center gap-2 rounded-lg border border-white/10 bg-elevated px-3 py-2">
            <code className="min-w-0 flex-1 break-all font-mono text-xs text-ink-primary">
              {TWO_FACTOR_SECRET}
            </code>
            <button
              type="button"
              onClick={() => copy(TWO_FACTOR_SECRET)}
              className="shrink-0 rounded-md p-1.5 text-ink-secondary transition-colors duration-150 hover:bg-hover hover:text-ink-primary"
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
        <Button variant="primary" onClick={confirm} disabled={code.length !== 6}>
          {t("settings.twoFactorEnable")}
        </Button>
      </ModalFooter>
    </Modal>
  );
}
