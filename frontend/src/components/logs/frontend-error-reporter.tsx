import { useEffect } from "react";

import { postFrontendLog } from "@/api";
import { getToken } from "@/api/client";

export function FrontendErrorReporter() {
  useEffect(() => {
    const send = (message: string, fields: Record<string, string> = {}) => {
      if (!getToken()) return;
      postFrontendLog({
        level: "error",
        message,
        fields: {
          path: window.location.hash || window.location.pathname,
          user_agent: window.navigator.userAgent,
          ...fields,
        },
      }).catch(() => {
        // Logging must never create another user-visible failure.
      });
    };

    const onError = (event: ErrorEvent) => {
      send(event.message || "Frontend runtime error", {
        filename: event.filename,
        line: String(event.lineno || ""),
        column: String(event.colno || ""),
        stack: event.error?.stack ? String(event.error.stack) : "",
      });
    };

    const onRejection = (event: PromiseRejectionEvent) => {
      const reason = event.reason;
      send(reason instanceof Error ? reason.message : "Unhandled promise rejection", {
        stack: reason instanceof Error && reason.stack ? reason.stack : String(reason ?? ""),
      });
    };

    window.addEventListener("error", onError);
    window.addEventListener("unhandledrejection", onRejection);
    return () => {
      window.removeEventListener("error", onError);
      window.removeEventListener("unhandledrejection", onRejection);
    };
  }, []);

  return null;
}
