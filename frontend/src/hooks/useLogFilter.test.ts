import { expect, test } from "vitest";
import { renderHook } from "@testing-library/react";

import { useLogFilter } from "@/hooks/useLogFilter";
import type { LogEntry } from "@/lib/store";

const logs: LogEntry[] = [
  { id: "1", t: 1, level: "info", source: "panel", message: "server started" },
  {
    id: "2",
    t: 2,
    level: "error",
    source: "frontend",
    message: "render failed",
    requestId: "req_2",
    fields: { component: "Dashboard" },
  },
];

test("filters logs by level, source, and fields", () => {
  const { result } = renderHook(() =>
    useLogFilter(logs, { level: "error", source: "frontend", query: "dashboard" }),
  );

  expect(result.current).toHaveLength(1);
  expect(result.current[0].id).toBe("2");
});
