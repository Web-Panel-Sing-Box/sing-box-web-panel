import { expect, test } from "vitest";
import { screen } from "@testing-library/react";

import { DashboardPage } from "@/pages/DashboardPage";
import { renderWithProviders } from "@/test/test-utils";

test("removes quick links and exposes protocol links to filtered inbounds", () => {
  renderWithProviders(<DashboardPage />);

  expect(screen.queryByText("Manage inbounds")).not.toBeInTheDocument();
  expect(screen.getAllByText("Traffic").length).toBeGreaterThan(0);
  expect(screen.getByText("Inbounds active")).toBeInTheDocument();
  expect(screen.getByRole("link", { name: "vless" })).toHaveAttribute(
    "href",
    "/inbounds?protocol=vless",
  );
});
