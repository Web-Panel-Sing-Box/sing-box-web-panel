import { useEffect, useState } from "react";
import { Plus, Trash2 } from "lucide-react";

import { listScheduledTasks, createScheduledTask, updateScheduledTask, deleteScheduledTask } from "@/api";
import type { ScheduledTaskDTO, ScheduledTaskCreateRequest } from "@/api/scheduled-tasks";
import { Button } from "@/components/ui/button";
import { Input, Label } from "@/components/ui/input";
import { Modal, ModalBody, ModalFooter, ModalHeader } from "@/components/ui/modal";
import { Select } from "@/components/ui/select";
import { Toggle } from "@/components/ui/toggle";
import { useToast } from "@/components/ui/toast";
import { useDisclosure } from "@/hooks/useDisclosure";
import { useI18n } from "@/lib/i18n";

export function ScheduledTaskList() {
  const { t } = useI18n();
  const { push } = useToast();
  const modal = useDisclosure();
  const ACTIONS = [
    { value: "reset_traffic_all", label: t("settings.tasks.actionResetTraffic") },
    { value: "delete_expired_clients", label: t("settings.tasks.actionDeleteExpired") },
    { value: "backup_database", label: t("settings.tasks.actionBackupDb") },
    { value: "rotate_reality_keys", label: t("settings.tasks.actionRotateReality") },
  ];
  const [tasks, setTasks] = useState<ScheduledTaskDTO[]>([]);
  const [loading, setLoading] = useState(true);

  function load() {
    listScheduledTasks()
      .then(setTasks)
      .catch(() => push(t("settings.tasks.loadError"), "error"))
      .finally(() => setLoading(false));
  }

  useEffect(() => { load(); }, []);

  return (
    <div className="space-y-4">
      <div className="flex items-center justify-between">
        <p className="text-xs text-ink-tertiary">{t("settings.tasks.listHint")}</p>
        <Button variant="secondary" size="sm" onClick={modal.open}>
          <Plus size={14} />
          {t("settings.tasks.add")}
        </Button>
      </div>

      {loading ? null : tasks.length === 0 ? (
        <p className="py-4 text-center text-xs text-ink-tertiary">{t("settings.tasks.empty")}</p>
      ) : (
        <div className="space-y-2">
          {tasks.map((task) => (
            <div
              key={task.id}
              className="flex items-center justify-between gap-3 rounded-lg border border-white/10 bg-elevated px-3 py-2.5"
            >
              <div className="min-w-0 flex-1">
                <div className="flex items-center gap-2">
                  <p className="truncate text-sm text-ink-primary">{task.name}</p>
                  <span className="shrink-0 rounded bg-white/5 px-1.5 py-0.5 font-mono text-[10px] text-ink-tertiary">
                    {task.action}
                  </span>
                </div>
                <p className="mt-0.5 font-mono text-[11px] text-ink-tertiary">{task.cronExpr}</p>
              </div>
              <div className="flex shrink-0 items-center gap-2">
                <Toggle
                  checked={task.enabled}
                  onChange={(v) => {
                    updateScheduledTask(task.id, { enabled: v }).then(load);
                  }}
                />
                <button
                  type="button"
                  onClick={() => {
                    deleteScheduledTask(task.id).then(() => {
                      push(t("settings.tasks.deleted"), "success");
                      load();
                    });
                  }}
                  className="grid size-7 place-items-center rounded-md text-ink-tertiary transition-colors hover:bg-hover hover:text-danger"
                  title={t("common.delete")}
                >
                  <Trash2 size={14} />
                </button>
              </div>
            </div>
          ))}
        </div>
      )}

      <AddTaskModal
        open={modal.isOpen}
        onClose={modal.close}
        actions={ACTIONS}
        onCreated={() => {
          load();
          push(t("settings.tasks.created"), "success");
          modal.close();
        }}
      />
    </div>
  );
}

type AddTaskModalProps = {
  open: boolean;
  onClose: () => void;
  actions: { value: string; label: string }[];
  onCreated: () => void;
};

function AddTaskModal({ open, onClose, actions, onCreated }: AddTaskModalProps) {
  const { t } = useI18n();
  const { push } = useToast();
  const [name, setName] = useState("");
  const [cronExpr, setCronExpr] = useState("");
  const [action, setAction] = useState("reset_traffic_all");

  function save() {
    const body: ScheduledTaskCreateRequest = { name, cronExpr, action };
    createScheduledTask(body)
      .then(onCreated)
      .catch((e) => push(e?.body?.error ?? t("settings.tasks.createError"), "error"));
  }

  function close() {
    setName("");
    setCronExpr("");
    setAction("reset_traffic_all");
    onClose();
  }

  return (
    <Modal open={open} onClose={close} width="max-w-[440px]">
      <ModalHeader title={t("settings.tasks.addTitle")} onClose={close} />
      <ModalBody className="space-y-3">
        <div>
          <Label>{t("settings.tasks.name")}</Label>
          <Input value={name} onChange={(e) => setName(e.target.value)} />
        </div>
        <div>
          <Label>{t("settings.tasks.cronExpr")}</Label>
          <Input value={cronExpr} onChange={(e) => setCronExpr(e.target.value)} mono placeholder="0 0 1 * * *" />
        </div>
        <div>
          <Label>{t("settings.tasks.action")}</Label>
          <Select value={action} options={actions} onChange={setAction} />
        </div>
      </ModalBody>
      <ModalFooter>
        <Button variant="primary" onClick={save} disabled={!name || !cronExpr}>
          {t("common.save")}
        </Button>
      </ModalFooter>
    </Modal>
  );
}
