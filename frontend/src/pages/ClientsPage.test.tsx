import { expect, test } from "vitest";
import { screen, waitFor } from "@testing-library/react";

import { ClientsPage } from "@/pages/ClientsPage";
import { renderWithProviders } from "@/test/test-utils";

test("initializes the inbound filter from the URL", async () => {
  renderWithProviders(<ClientsPage />, { route: "/clients?inbound=ib_04" });

  expect(
    await screen.findByRole("button", { name: "frankfurt-ws-01" }),
  ).toBeInTheDocument();
  expect(screen.getByText("miyu")).toBeInTheDocument();
  await waitFor(() =>
    expect(screen.queryByText("alex_kim")).not.toBeInTheDocument(),
  );
});
