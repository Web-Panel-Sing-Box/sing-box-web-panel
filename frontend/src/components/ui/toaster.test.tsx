import { expect, test } from "vitest";
import { screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";

import { useToast } from "@/components/ui/toast";
import { renderWithProviders } from "@/test/test-utils";

function Probe() {
  const { push } = useToast();
  return (
    <button type="button" onClick={() => push("Settings saved", "success")}>
      notify
    </button>
  );
}

test("renders a green textual check mark for success toasts", async () => {
  const user = userEvent.setup();
  renderWithProviders(<Probe />, { withStore: false });

  await user.click(screen.getByRole("button", { name: "notify" }));

  expect(screen.getByText("Settings saved")).toBeInTheDocument();
  expect(screen.getByText("✓")).toBeInTheDocument();
});
