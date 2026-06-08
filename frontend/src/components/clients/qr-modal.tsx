
import { Modal, ModalBody, ModalHeader } from "@/components/ui/modal";
import { Button } from "@/components/ui/button";
import { QrCode } from "@/components/ui/qr-code";

type QrModalProps = {
  open: boolean;
  onClose: () => void;
  payload: string;
  label?: string;
};

export function QrModal({ open, onClose, payload, label = "Subscription" }: QrModalProps) {
  return (
    <Modal open={open} onClose={onClose} width="max-w-[380px]">
      <ModalHeader title={label} onClose={onClose} />
      <ModalBody className="flex flex-col items-center gap-4 pb-6">
        <QrCode payload={payload} size={220} />
        <p className="break-all text-center font-mono text-[11px] text-ink-tertiary">{payload}</p>
        <Button variant="secondary" onClick={onClose} className="w-full">
          Close
        </Button>
      </ModalBody>
    </Modal>
  );
}
