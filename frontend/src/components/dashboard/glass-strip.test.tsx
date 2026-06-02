import { expect, test, vi } from "vitest";
import { screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";

import { GlassStrip } from "@/components/dashboard/glass-strip";
import { renderWithProviders } from "@/test/test-utils";

const { startCore, stopCore } = vi.hoisted(() => ({
  startCore: vi.fn(() => Promise.resolve()),
  stopCore: vi.fn(() => Promise.resolve()),
}));

vi.mock("@/lib/store", () => ({
  useMetrics: () => ({
    metrics: {
      coreRunning: true,
      coreVersion: "1.11.0",
      uptimeSeconds: 0,
      totalSent: 0,
      totalReceived: 0,
    },
    history: [],
  }),
  useStoreActions: () => ({ startCore, stopCore }),
}));

test("confirms before stopping a running core, then calls the real stop action", async () => {
  const user = userEvent.setup();
  renderWithProviders(<GlassStrip />);

  // No confirmation appears just from opening the dashboard.
  expect(
    screen.queryByText("Are you sure you want to stop the core?"),
  ).not.toBeInTheDocument();

  await user.click(screen.getByRole("button", { name: /active/i }));
  expect(
    screen.getByText("Are you sure you want to stop the core?"),
  ).toBeInTheDocument();

  await user.click(screen.getByRole("button", { name: "Stop core" }));
  expect(stopCore).toHaveBeenCalledTimes(1);
  expect(await screen.findByText("Core stopped")).toBeInTheDocument();
  await waitFor(() =>
    expect(
      screen.queryByText("Are you sure you want to stop the core?"),
    ).not.toBeInTheDocument(),
  );
});
