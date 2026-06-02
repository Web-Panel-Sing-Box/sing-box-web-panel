import { afterEach, expect, test } from "vitest";
import { screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";

import { useI18n } from "@/lib/i18n";
import { renderWithProviders } from "@/test/test-utils";

afterEach(() => {
  window.localStorage.clear();
});

function Probe() {
  const { language, setLanguage, t } = useI18n();
  return (
    <div>
      <p>{language}</p>
      <p>{t("dashboard.total", { count: 6 })}</p>
      <button type="button" onClick={() => setLanguage("ru")}>
        switch
      </button>
    </div>
  );
}

test("defaults to English and persists Russian selection", async () => {
  const user = userEvent.setup();
  renderWithProviders(<Probe />);

  expect(screen.getByText("en")).toBeInTheDocument();
  expect(screen.getByText("6 total")).toBeInTheDocument();

  await user.click(screen.getByRole("button", { name: "switch" }));

  expect(screen.getByText("ru")).toBeInTheDocument();
  expect(window.localStorage.getItem("shilka:language")).toBe("ru");
  expect(screen.getByText("всего 6")).toBeInTheDocument();
});
