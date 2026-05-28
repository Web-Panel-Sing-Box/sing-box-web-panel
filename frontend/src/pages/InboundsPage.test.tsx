import { expect, test } from "vitest";
import { screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";

import { InboundsPage } from "@/pages/InboundsPage";
import { renderWithProviders } from "@/test/test-utils";

test("opens inbound rows directly in edit mode and starts clone mode from the modal", async () => {
  const user = userEvent.setup();
  renderWithProviders(<InboundsPage />);

  await user.click(screen.getByText("berlin-edge-01"));
  expect(screen.getByText("Edit inbound connection")).toBeInTheDocument();

  await user.click(screen.getByRole("button", { name: "Clone" }));
  expect(screen.getByText("Clone inbound connection")).toBeInTheDocument();
  expect(screen.getByDisplayValue("berlin-edge-01-copy")).toBeInTheDocument();
});

test("shows a confirmation modal before deleting an inbound", async () => {
  const user = userEvent.setup();
  renderWithProviders(<InboundsPage />);

  await user.click(screen.getByText("berlin-edge-01"));
  await user.click(screen.getByRole("button", { name: "Delete" }));

  expect(screen.getByText("Delete this inbound?")).toBeInTheDocument();
  expect(screen.getByText("berlin-edge-01 will be removed from the mock list.")).toBeInTheDocument();
});

test("honors protocol filters from dashboard links", () => {
  renderWithProviders(<InboundsPage />, { route: "/inbounds?protocol=naive" });

  expect(screen.getByText("Filtered by naive")).toBeInTheDocument();
  expect(screen.getByText("amsterdam-naive-01")).toBeInTheDocument();
  expect(screen.queryByText("berlin-edge-01")).not.toBeInTheDocument();
});
