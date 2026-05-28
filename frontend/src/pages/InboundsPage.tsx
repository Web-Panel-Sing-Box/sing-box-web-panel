
import { useState } from "react";
import { Plus } from "lucide-react";

import { Button } from "@/components/ui/button";
import { InboundsList } from "@/components/inbounds/inbounds-list";
import { InboundFormModal } from "@/components/inbounds/inbound-form-modal";

export function InboundsPage() {
  const [open, setOpen] = useState(false);
  return (
    <div className="mx-auto flex max-w-[1240px] flex-col gap-6">
      <div className="flex flex-wrap items-center justify-between gap-3">
        <div>
          <h2 className="text-2xl font-semibold text-ink-primary">Inbounds</h2>
          <p className="mt-1 text-sm text-ink-tertiary">Listeners exposed by the local sing-box core</p>
        </div>
        <Button variant="white" onClick={() => setOpen(true)} className="rounded-full">
          <Plus size={16} />
          New configuration
        </Button>
      </div>

      <InboundsList />

      <InboundFormModal open={open} onClose={() => setOpen(false)} />
    </div>
  );
}
