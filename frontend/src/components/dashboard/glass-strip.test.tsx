import { expect, test } from "vitest";
import { screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";

import { GlassStrip } from "@/components/dashboard/glass-strip";
import { renderWithProviders } from "@/test/test-utils";

test("asks for confirmation only when stopping an active core", async () => {
  const user = userEvent.setup();
  renderWithProviders(<GlassStrip />, { seed: { metrics: { coreRunning: true } } });

  await user.click(screen.getByRole("button", { name: /active/i }));
  expect(screen.getByText("Are you sure you want to stop the core?")).toBeInTheDocument();

  await user.click(screen.getByRole("button", { name: "Stop core" }));
  expect(screen.getByText("Core stopped")).toBeInTheDocument();
  await waitFor(() => expect(screen.queryByText("Are you sure you want to stop the core?")).not.toBeInTheDocument());
  await waitFor(() => expect(screen.getByRole("button", { name: /stopped/i })).toBeInTheDocument());

  await user.click(screen.getByRole("button", { name: /stopped/i }));
  expect(screen.queryByText("Are you sure you want to stop the core?")).not.toBeInTheDocument();
  expect(screen.getByText("Core started")).toBeInTheDocument();
});
