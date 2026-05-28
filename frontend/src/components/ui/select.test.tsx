import { expect, test } from "vitest";
import { screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { useState } from "react";

import { Select } from "@/components/ui/select";
import { renderWithProviders } from "@/test/test-utils";

function Probe() {
  const [value, setValue] = useState("one");
  return (
    <div className="overflow-hidden">
      <Select
        value={value}
        onChange={setValue}
        options={[
          { value: "one", label: "One" },
          { value: "two", label: "Two" }
        ]}
      />
    </div>
  );
}

test("renders dropdown options in a high z-index portal", async () => {
  const user = userEvent.setup();
  renderWithProviders(<Probe />, { withStore: false });

  await user.click(screen.getByRole("button", { name: /one/i }));

  const listbox = screen.getByRole("listbox");
  expect(listbox).toBeInTheDocument();
  expect(listbox.parentElement).toBe(document.body);
  expect(listbox).toHaveClass("z-[120]");

  await user.click(screen.getByRole("option", { name: "Two" }));
  expect(screen.getByRole("button", { name: /two/i })).toBeInTheDocument();
});
