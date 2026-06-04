import { afterEach, expect, test, vi } from "vitest";
import { screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";

import { clearToken, setToken } from "@/api/client";
import { PanelVersionCard } from "@/components/dashboard/panel-version-card";
import { renderWithProviders } from "@/test/test-utils";

const { getPanelVersion, startPanelUpdate } = vi.hoisted(() => ({
  getPanelVersion: vi.fn(),
  startPanelUpdate: vi.fn(),
}));

vi.mock("@/api/panel", () => ({
  getPanelVersion,
  startPanelUpdate,
}));

afterEach(() => {
  clearToken();
  vi.clearAllMocks();
});

test("renders panel versions and starts update", async () => {
  setToken("test-token");
  getPanelVersion.mockResolvedValue({
    currentVersion: "1.0.0",
    latestVersion: "1.1.0",
    updateAvailable: true,
    releaseURL: "https://example.test/release",
    checkedAt: "2026-06-04T00:00:00Z",
    status: "update_available",
  });
  startPanelUpdate.mockResolvedValue({
    currentVersion: "1.0.0",
    latestVersion: "1.1.0",
    updateAvailable: true,
    releaseURL: "https://example.test/release",
    checkedAt: "2026-06-04T00:00:00Z",
    status: "running",
  });

  const user = userEvent.setup();
  renderWithProviders(<PanelVersionCard />);

  expect(await screen.findByText("1.0.0")).toBeInTheDocument();
  expect(screen.getByText("1.1.0")).toBeInTheDocument();
  expect(screen.getByText("Update available")).toBeInTheDocument();

  await user.click(screen.getByRole("button", { name: "Update panel" }));

  expect(startPanelUpdate).toHaveBeenCalledTimes(1);
  expect(await screen.findByText("Panel update started")).toBeInTheDocument();
});
