
import { useMemo } from "react";

import { Modal, ModalBody, ModalHeader } from "@/components/ui/modal";
import { Button } from "@/components/ui/button";

type QrModalProps = {
  open: boolean;
  onClose: () => void;
  payload: string;
  label?: string;
};

function deterministicQr(text: string, size = 21): number[][] {
  let h = 2166136261;
  for (let i = 0; i < text.length; i++) {
    h = (h ^ text.charCodeAt(i)) >>> 0;
    h = Math.imul(h, 16777619) >>> 0;
  }
  const grid: number[][] = [];
  for (let y = 0; y < size; y++) {
    grid.push(new Array(size).fill(0));
  }
  function rand() {
    h = Math.imul(h ^ (h >>> 15), 2246822507) >>> 0;
    h = Math.imul(h ^ (h >>> 13), 3266489909) >>> 0;
    h = (h ^ (h >>> 16)) >>> 0;
    return h / 0xffffffff;
  }
  for (let y = 0; y < size; y++) {
    for (let x = 0; x < size; x++) {
      grid[y][x] = rand() > 0.5 ? 1 : 0;
    }
  }
  // Three position markers
  const drawMarker = (ox: number, oy: number) => {
    for (let y = 0; y < 7; y++) {
      for (let x = 0; x < 7; x++) {
        const edge = x === 0 || x === 6 || y === 0 || y === 6;
        const inner = x >= 2 && x <= 4 && y >= 2 && y <= 4;
        grid[oy + y][ox + x] = edge || inner ? 1 : 0;
      }
    }
  };
  drawMarker(0, 0);
  drawMarker(size - 7, 0);
  drawMarker(0, size - 7);
  return grid;
}

export function QrModal({ open, onClose, payload, label = "Subscription" }: QrModalProps) {
  const grid = useMemo(() => deterministicQr(payload || "empty", 25), [payload]);
  const size = grid.length;
  const unit = 8;
  const padding = 16;
  const svgSize = size * unit + padding * 2;
  return (
    <Modal open={open} onClose={onClose} width="max-w-[380px]">
      <ModalHeader title={label} onClose={onClose} />
      <ModalBody className="flex flex-col items-center gap-4 pb-6">
        <div className="rounded-2xl bg-white p-4">
          <svg width={svgSize} height={svgSize} viewBox={`0 0 ${svgSize} ${svgSize}`}>
            <rect width={svgSize} height={svgSize} fill="#ffffff" />
            {grid.flatMap((row, y) =>
              row.map((v, x) =>
                v ? <rect key={`${x}-${y}`} x={padding + x * unit} y={padding + y * unit} width={unit} height={unit} fill="#171717" /> : null
              )
            )}
          </svg>
        </div>
        <p className="break-all text-center font-mono text-[11px] text-ink-tertiary">{payload}</p>
        <Button variant="secondary" onClick={onClose} className="w-full">
          Close
        </Button>
      </ModalBody>
    </Modal>
  );
}
