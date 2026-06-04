import { expect, test, vi } from "vitest";
import { screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";

import * as api from "@/api";
import { InboundFormModal } from "@/components/inbounds/inbound-form-modal";
import { renderWithProviders } from "@/test/test-utils";

vi.mock("@/api", () => ({
  createInbound: vi.fn().mockResolvedValue({}),
}));

test("sends allowInsecure for new TLS inbounds", async () => {
  const user = userEvent.setup();
  const onClose = vi.fn();
  renderWithProviders(<InboundFormModal open onClose={onClose} />, { seed: { inbounds: [] } });

  const toggle = screen.getByRole("switch", { name: /allow insecure/i });
  expect(toggle).toHaveAttribute("aria-checked", "true");

  await user.type(screen.getByPlaceholderText("e.g. vadim-vless#0001"), "hy2-self-signed");
  await user.click(screen.getByRole("button", { name: "Save" }));

  await waitFor(() => expect(api.createInbound).toHaveBeenCalledTimes(1));
  expect(api.createInbound).toHaveBeenCalledWith(
    expect.objectContaining({
      protocol: "naive",
      tls: "tls",
      allowInsecure: true,
    }),
  );
});
