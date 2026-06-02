import { expect, test } from "vitest";
import { screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";

import { InboundsPage } from "@/pages/InboundsPage";
import { renderWithProviders } from "@/test/test-utils";
import type { InboundDTO } from "@/lib/store";

const SEED_INBOUNDS: InboundDTO[] = [
  { id: "1", remark: "berlin-edge-01", protocol: "vless", port: 44321, transmission: "tcp", tls: "reality", sni: "www.cloudflare.com", dest: "www.cloudflare.com:443", enabled: true, clientCount: 18, createdAt: "2026-03-12T14:11:00Z" },
  { id: "2", remark: "amsterdam-naive-01", protocol: "naive", port: 38119, transmission: "tcp", tls: "tls", enabled: true, clientCount: 31, createdAt: "2026-04-02T18:00:00Z" },
];

test("opens inbound rows directly in edit mode and starts clone mode from the modal", async () => {
  const user = userEvent.setup();
  renderWithProviders(<InboundsPage />, { seed: { inbounds: SEED_INBOUNDS } });

  await user.click(screen.getByText("berlin-edge-01"));
  expect(
    await screen.findByText("Edit inbound connection"),
  ).toBeInTheDocument();

  await user.click(await screen.findByRole("button", { name: "Clone" }));
  expect(
    await screen.findByText("Clone inbound connection"),
  ).toBeInTheDocument();
  expect(screen.getByDisplayValue("berlin-edge-01-copy")).toBeInTheDocument();
});

test("shows a confirmation modal before deleting an inbound", async () => {
  const user = userEvent.setup();
  renderWithProviders(<InboundsPage />, { seed: { inbounds: SEED_INBOUNDS } });

  await user.click(screen.getByText("berlin-edge-01"));
  await user.click(await screen.findByRole("button", { name: "Delete" }));

  expect(await screen.findByText("Delete this inbound?")).toBeInTheDocument();
});

test("honors protocol filters from dashboard links", () => {
  renderWithProviders(<InboundsPage />, { seed: { inbounds: SEED_INBOUNDS }, route: "/inbounds?protocol=naive" });

  expect(screen.getByText("Filtered by naive")).toBeInTheDocument();
  expect(screen.getByText("amsterdam-naive-01")).toBeInTheDocument();
  expect(screen.queryByText("berlin-edge-01")).not.toBeInTheDocument();
});
