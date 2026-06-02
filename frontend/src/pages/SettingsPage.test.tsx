import { afterEach, expect, test, vi } from "vitest";
import { screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";

import { SettingsPage } from "@/pages/SettingsPage";
import { renderWithProviders } from "@/test/test-utils";

vi.mock("@/api/settings", () => ({
  getSettings: vi.fn().mockResolvedValue({}),
  saveSettings: vi.fn().mockResolvedValue({ ok: "saved" }),
}));

afterEach(() => {
  window.localStorage.clear();
});

test("uses a single page-level save button and switches to Russian", async () => {
  const user = userEvent.setup();
  renderWithProviders(<SettingsPage />);

  expect(screen.getAllByRole("button", { name: "Save" })).toHaveLength(1);

  await user.click(screen.getByRole("button", { name: "English" }));
  await user.click(screen.getByRole("option", { name: "Русский" }));

  expect(
    screen.getByRole("heading", { name: "Настройки" }),
  ).toBeInTheDocument();
  expect(screen.getAllByRole("button", { name: "Сохранить" })).toHaveLength(1);
});
