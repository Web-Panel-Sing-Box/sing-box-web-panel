import { useCallback, useState } from "react";

type CopyResult = {
  copied: boolean;
  copy: (text: string) => Promise<boolean>;
  reset: () => void;
};

export function useCopyToClipboard(resetMs = 1500): CopyResult {
  const [copied, setCopied] = useState(false);

  const copy = useCallback(
    async (text: string) => {
      if (typeof navigator === "undefined" || !navigator.clipboard) return false;
      try {
        await navigator.clipboard.writeText(text);
        setCopied(true);
        window.setTimeout(() => setCopied(false), resetMs);
        return true;
      } catch {
        return false;
      }
    },
    [resetMs]
  );

  const reset = useCallback(() => setCopied(false), []);

  return { copied, copy, reset };
}
