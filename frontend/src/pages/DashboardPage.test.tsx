import { expect, test } from "vitest";
import { screen } from "@testing-library/react";

import { DashboardPage } from "@/pages/DashboardPage";
import { renderWithProviders } from "@/test/test-utils";
import type { InboundDTO } from "@/lib/store";

const INBOUNDS: InboundDTO[] = [
  { id: "1", remark: "edge", protocol: "vless", port: 44321, transmission: "tcp", tls: "reality", enabled: true, clientCount: 5, createdAt: "2026-01-01T00:00:00Z" },
];

test("removes quick links and exposes protocol links to filtered inbounds", () => {
  renderWithProviders(<DashboardPage />, { seed: { inbounds: INBOUNDS } });

  expect(screen.queryByText("Manage inbounds")).not.toBeInTheDocument();
  expect(screen.getAllByText("Traffic").length).toBeGreaterThan(0);
  expect(screen.getByText("Inbounds active")).toBeInTheDocument();
  expect(screen.getByRole("link", { name: "vless" })).toHaveAttribute(
    "href",
    "/inbounds?protocol=vless",
  );
});
