import { expect, test } from "@playwright/test";

test("dashboard core confirmation and removed quick links", async ({ page }) => {
  await page.goto("/");

  await expect(page.getByText("Inbounds active")).toBeVisible();
  await expect(page.getByText("Manage inbounds")).toHaveCount(0);

  await page.getByRole("button", { name: /active/i }).click();
  await expect(page.getByText("Are you sure you want to stop the core?")).toBeVisible();
  await page.getByRole("button", { name: "Keep running" }).evaluate((el: HTMLElement) => el.click());
  await expect(page.getByText("Are you sure you want to stop the core?")).toHaveCount(0);
});

test("inbound rows open edit modal directly and confirm delete", async ({ page }) => {
  await page.goto("/inbounds");

  await page.getByText("berlin-edge-01").click();
  await expect(page.getByText("Edit inbound connection")).toBeVisible();

  await page.getByRole("button", { name: "Delete" }).evaluate((el: HTMLElement) => el.click());
  await expect(page.getByText("Delete this inbound?")).toBeVisible();
});

test("dropdowns render above modal content", async ({ page }) => {
  await page.goto("/inbounds");

  await page.getByRole("button", { name: "New configuration" }).click();
  await expect(page.getByText("New inbound connection")).toBeVisible();
  await page.getByRole("button", { name: "Naive Proxy" }).first().evaluate((el: HTMLElement) => el.click());

  const listbox = page.getByRole("listbox");
  await expect(listbox).toBeVisible();
  await expect(listbox).toHaveCSS("z-index", "120");
});

test("clients page accepts inbound filter from URL", async ({ page }) => {
  await page.goto("/clients?inbound=ib_04");

  await expect(page.getByRole("button", { name: "frankfurt-ws-01" }).first()).toBeVisible();
  await expect(page.getByText("miyu")).toBeVisible();
  await expect(page.getByText("alex_kim")).toHaveCount(0);
});

test("settings can switch the UI to Russian", async ({ page }) => {
  await page.goto("/settings");

  await page.getByRole("button", { name: "English" }).click();
  await page.getByRole("option", { name: "Русский" }).click();

  await expect(page.getByRole("heading", { name: "Настройки" })).toBeVisible();
});
