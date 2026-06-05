import type { ReactElement, ReactNode } from "react";
import { beforeEach, expect, test, vi } from "vitest";
import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { LazyMotion, domMax } from "framer-motion";

import { AddClientModal } from "@/components/clients/add-client-modal";
import type { InboundDTO } from "@/api/inbounds";

// Every context hook the modal consumes is mocked so no real providers are
// needed (mocking "@/lib/store" alone splits the shared context module graph
// under Vitest). `mockInbounds` is swapped + the modal re-rendered to simulate
// StoreProvider's 3s polling delivering a fresh inbounds array (SIN-35).
let mockInbounds: InboundDTO[] = [];
const addClient = vi.fn().mockResolvedValue(undefined);

vi.mock("@/lib/store", () => ({
  useInbounds: () => mockInbounds,
  useStoreActions: () => ({ addClient }),
}));

vi.mock("@/api", () => ({
  listNodes: vi.fn().mockResolvedValue([]),
}));

vi.mock("@/components/ui/toast", () => ({
  useToast: () => ({ push: vi.fn(), dismiss: vi.fn(), toasts: [] }),
}));

vi.mock("@/lib/i18n", () => ({
  useI18n: () => ({ t: (key: string) => key, lang: "en", setLang: vi.fn() }),
}));

// Use RTL's `wrapper` so `rerender` re-applies the same wrapper instead of
// replacing the whole tree (which would remount the modal and lose its state).
function Wrapper({ children }: { children: ReactNode }) {
  return <LazyMotion features={domMax} strict>{children}</LazyMotion>;
}

function renderModal(ui: ReactElement) {
  return render(ui, { wrapper: Wrapper });
}

const IB_A: InboundDTO = {
  id: "ib-1", remark: "frankfurt", protocol: "vless", port: 27440,
  transmission: "ws", tls: "tls", enabled: true, clientCount: 0,
  createdAt: "2026-04-18T12:24:00Z",
};
const IB_B: InboundDTO = { ...IB_A, id: "ib-2", remark: "amsterdam" };

beforeEach(() => {
  mockInbounds = [IB_A];
  addClient.mockClear();
});

test("keeps typed fields when a polling refresh delivers a new inbounds array", async () => {
  const user = userEvent.setup();
  const { rerender } = renderModal(<AddClientModal open onClose={vi.fn()} />);

  const nameInput = screen.getByPlaceholderText(/vadim_denisych/i);
  await user.type(nameInput, "alice");

  const quotaInput = screen.getByPlaceholderText("0");
  await user.clear(quotaInput);
  await user.type(quotaInput, "250");

  const toggle = screen.getByRole("switch");
  await user.click(toggle);

  expect(nameInput).toHaveValue("alice");
  expect(quotaInput).toHaveValue("250");
  expect(toggle).toHaveAttribute("aria-checked", "true");

  // Simulate the next poll: a brand-new array with new object references.
  mockInbounds = [{ ...IB_A }];
  rerender(<AddClientModal open onClose={vi.fn()} />);

  expect(screen.getByPlaceholderText(/vadim_denisych/i)).toHaveValue("alice");
  expect(screen.getByPlaceholderText("0")).toHaveValue("250");
  expect(screen.getByRole("switch")).toHaveAttribute("aria-checked", "true");
});

test("when the selected inbound disappears, only the inbound selection changes", async () => {
  mockInbounds = [IB_A, IB_B];
  const user = userEvent.setup();
  const { rerender } = renderModal(<AddClientModal open onClose={vi.fn()} />);

  // Default selection is the first inbound; switch to the second, then edit name.
  await user.click(await screen.findByRole("button", { name: "frankfurt" }));
  await user.click(await screen.findByRole("option", { name: "amsterdam" }));
  await user.type(screen.getByPlaceholderText(/vadim_denisych/i), "bob");

  expect(screen.getByRole("button", { name: "amsterdam" })).toBeInTheDocument();

  // Poll removes the currently selected inbound (amsterdam).
  mockInbounds = [{ ...IB_A }];
  rerender(<AddClientModal open onClose={vi.fn()} />);

  // Name is preserved; the inbound select falls back to the remaining option.
  expect(screen.getByPlaceholderText(/vadim_denisych/i)).toHaveValue("bob");
  expect(await screen.findByRole("button", { name: "frankfurt" })).toBeInTheDocument();
});
