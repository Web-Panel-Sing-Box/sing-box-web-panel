import type { ReactElement } from "react";
import { Suspense } from "react";
import { render, type RenderOptions } from "@testing-library/react";
import { LazyMotion, domMax } from "framer-motion";
import { MemoryRouter } from "react-router-dom";

import { Toaster } from "@/components/ui/toaster";
import { AuthProvider } from "@/lib/auth";
import { I18nProvider } from "@/lib/i18n";
import { StoreProvider, type InboundDTO, type ClientDTO } from "@/lib/store";

type Options = RenderOptions & {
  route?: string;
  seed?: { inbounds?: InboundDTO[]; clients?: ClientDTO[]; metrics?: { coreRunning?: boolean } };
};

export function renderWithProviders(
  ui: ReactElement,
  { route = "/", seed, ...options }: Options = {}
) {
  const body = seed ? <StoreProvider seed={seed}>{ui}</StoreProvider> : ui;
  return render(
    <I18nProvider>
      <LazyMotion features={domMax} strict>
        <Toaster>
          <AuthProvider>
            <MemoryRouter initialEntries={[route]}>
              <Suspense fallback={null}>{body}</Suspense>
            </MemoryRouter>
          </AuthProvider>
        </Toaster>
      </LazyMotion>
    </I18nProvider>,
    options
  );
}
