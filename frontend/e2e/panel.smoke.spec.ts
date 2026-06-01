import { expect, test } from "@playwright/test";

test.beforeEach(async ({ page }) => {
  await page.addInitScript(() => {
    window.localStorage.setItem("shilka:auth", "1");
  });
});

test("dashboard renders with core status and traffic chart", async ({ page }) => {
  await page.goto("/dashboard");

  await expect(page.getByRole("heading", { name: "Traffic" })).toBeVisible();
  await expect(page.getByText("CPU")).toBeVisible();
});

test("inbound form opens new configuration modal", async ({ page }) => {
  await page.goto("/inbounds");

  await page.getByRole("button", { name: "New configuration" }).click();
  await expect(page.getByText("New inbound connection")).toBeVisible();

  // Protocol selector opens a dropdown
  await page.getByRole("button", { name: "Naive Proxy" }).first().evaluate((el: HTMLElement) => el.click());
  const listbox = page.getByRole("listbox");
  await expect(listbox).toBeVisible();
  await expect(listbox).toHaveCSS("z-index", "120");
});

test("settings page elements are present", async ({ page }) => {
  await page.goto("/settings");

  await expect(page.getByRole("button", { name: "Save" })).toBeVisible();
  expect(await page.title()).toBeTruthy();
});

test("settings can switch the UI to Russian", async ({ page }) => {
  await page.goto("/settings");

  await page.getByRole("button", { name: "English" }).click();
  await page.getByRole("option", { name: "Русский" }).click();

  await expect(page.getByRole("heading", { name: "Настройки" })).toBeVisible();
});

test("logs page renders the viewer", async ({ page }) => {
  await page.goto("/logs");

  await expect(page.getByText("Pause")).toBeVisible();
});
