import { QRCodeSVG } from "qrcode.react";

import { cn } from "@/lib/utils";

type QrCodeProps = {
  /** Data encoded into the QR code (e.g. an otpauth:// URI). */
  payload: string;
  /** Pixel size of the rendered QR square. */
  size?: number;
  className?: string;
};

/**
 * Real, scannable QR code rendered as SVG. Wraps the payload in the same white
 * rounded card framing used across the panel.
 */
export function QrCode({ payload, size = 180, className }: QrCodeProps) {
  return (
    <div className={cn("rounded-2xl bg-white p-4", className)}>
      <QRCodeSVG value={payload || " "} size={size} level="M" marginSize={4} />
    </div>
  );
}
