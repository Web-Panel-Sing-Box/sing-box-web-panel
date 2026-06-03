import { useEffect, useState } from "react";

import * as api from "@/api";
import { Button } from "@/components/ui/button";
import { Input, Label } from "@/components/ui/input";
import { Modal, ModalBody, ModalFooter, ModalHeader } from "@/components/ui/modal";
import { useToast } from "@/components/ui/toast";
import { useI18n } from "@/lib/i18n";

type TwoFactorDisableModalProps = {
  open: boolean;
  onClose: () => void;
  onDisabled: () => void;
};

export function TwoFactorDisableModal({ open, onClose, onDisabled }: TwoFactorDisableModalProps) {
  const { t } = useI18n();
  const { push } = useToast();
  const [code, setCode] = useState("");
  const [submitting, setSubmitting] = useState(false);

  useEffect(() => {
    if (!open) {
      setCode("");
      setSubmitting(false);
    }
  }, [open]);

  const confirm = async () => {
    if (!code || submitting) return;
    setSubmitting(true);
    try {
      await api.disableTOTP({ code });
      onDisabled();
      push(t("settings.twoFactorDisabled"), "success");
      onClose();
    } catch {
      push(t("settings.twoFactorInvalidCode"), "error");
    } finally {
      setSubmitting(false);
    }
  };

  return (
    <Modal open={open} onClose={onClose} width="max-w-[420px]">
      <ModalHeader
        title={t("settings.twoFactorDisableTitle")}
        subtitle={t("settings.twoFactorDisableHint")}
        onClose={onClose}
      />
      <ModalBody>
        <Label htmlFor="twofa-disable-code">{t("settings.twoFactorEnterCode")}</Label>
        <Input
          id="twofa-disable-code"
          value={code}
          onChange={(e) => setCode(e.target.value.replace(/[^0-9A-Za-z-]/g, "").slice(0, 9))}
          onKeyDown={(e) => {
            if (e.key === "Enter") confirm();
          }}
          autoComplete="one-time-code"
          mono
          autoFocus
        />
      </ModalBody>
      <ModalFooter>
        <Button variant="secondary" onClick={onClose}>
          {t("common.cancel")}
        </Button>
        <Button variant="danger" onClick={confirm} loading={submitting} disabled={!code || submitting}>
          {t("settings.twoFactorDisableConfirm")}
        </Button>
      </ModalFooter>
    </Modal>
  );
}
