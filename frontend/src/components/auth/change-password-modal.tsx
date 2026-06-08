import { useEffect, useState } from "react";

import * as api from "@/api";
import { ApiError } from "@/api/client";
import { Button } from "@/components/ui/button";
import { Input, Label } from "@/components/ui/input";
import { Modal, ModalBody, ModalFooter, ModalHeader } from "@/components/ui/modal";
import { useToast } from "@/components/ui/toast";
import { useI18n } from "@/lib/i18n";

type ChangePasswordModalProps = {
  open: boolean;
  onClose: () => void;
};

export function ChangePasswordModal({ open, onClose }: ChangePasswordModalProps) {
  const { t } = useI18n();
  const { push } = useToast();
  const [current, setCurrent] = useState("");
  const [next, setNext] = useState("");
  const [confirm, setConfirm] = useState("");
  const [submitting, setSubmitting] = useState(false);

  useEffect(() => {
    if (!open) {
      setCurrent("");
      setNext("");
      setConfirm("");
      setSubmitting(false);
    }
  }, [open]);

  const save = async () => {
    if (submitting) return;
    if (!current || !next) {
      push(t("settings.passwordRequired"), "error");
      return;
    }
    if (next !== confirm) {
      push(t("settings.passwordMismatch"), "error");
      return;
    }
    setSubmitting(true);
    try {
      await api.changePassword({ current_password: current, new_password: next });
      push(t("settings.passwordChanged"), "success");
      onClose();
    } catch (err) {
      const body = err instanceof ApiError ? err.body : null;
      const message =
        body && typeof body === "object" && body !== null && "error" in body
          ? String((body as { error: unknown }).error)
          : t("settings.passwordChangeFailed");
      push(message, "error");
    } finally {
      setSubmitting(false);
    }
  };

  return (
    <Modal open={open} onClose={onClose} width="max-w-[440px]">
      <ModalHeader title={t("settings.changePassword")} onClose={onClose} />
      <ModalBody className="space-y-3">
        <div>
          <Label>{t("settings.currentPassword")}</Label>
          <Input type="password" value={current} onChange={(e) => setCurrent(e.target.value)} autoComplete="current-password" />
        </div>
        <div>
          <Label>{t("settings.newPassword")}</Label>
          <Input type="password" value={next} onChange={(e) => setNext(e.target.value)} autoComplete="new-password" />
        </div>
        <div>
          <Label>{t("settings.confirmPassword")}</Label>
          <Input type="password" value={confirm} onChange={(e) => setConfirm(e.target.value)} autoComplete="new-password" />
        </div>
      </ModalBody>
      <ModalFooter>
        <Button variant="danger" onClick={onClose}>
          {t("common.cancel")}
        </Button>
        <Button variant="primary" onClick={save} loading={submitting}>
          {t("common.save")}
        </Button>
      </ModalFooter>
    </Modal>
  );
}
