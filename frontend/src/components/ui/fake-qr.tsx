import { useMemo } from "react";

import { cn } from "@/lib/utils";

/**
 * Deterministic stand-in for a real QR code: hashes the payload into a stable
 * pseudo-random module grid with the three position markers. Decorative only —
 * not scannable. Real QR generation arrives with the backend integration.
 */
function deterministicQr(text: string, size = 25): number[][] {
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

type FakeQrCodeProps = {
  payload: string;
  className?: string;
  /** Pixel size of a single module. */
  unit?: number;
};

export function FakeQrCode({ payload, className, unit = 8 }: FakeQrCodeProps) {
  const grid = useMemo(() => deterministicQr(payload || "empty", 25), [payload]);
  const size = grid.length;
  const padding = 16;
  const svgSize = size * unit + padding * 2;
  return (
    <div className={cn("rounded-2xl bg-white p-4", className)}>
      <svg width={svgSize} height={svgSize} viewBox={`0 0 ${svgSize} ${svgSize}`}>
        <rect width={svgSize} height={svgSize} fill="#ffffff" />
        {grid.flatMap((row, y) =>
          row.map((v, x) =>
            v ? (
              <rect
                key={`${x}-${y}`}
                x={padding + x * unit}
                y={padding + y * unit}
                width={unit}
                height={unit}
                fill="#171717"
              />
            ) : null
          )
        )}
      </svg>
    </div>
  );
}
