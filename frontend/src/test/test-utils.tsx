import type { ReactElement } from "react";
import { Suspense } from "react";
import { render, type RenderOptions } from "@testing-library/react";
import { LazyMotion, domMax } from "framer-motion";
import { MemoryRouter } from "react-router-dom";

import { Toaster } from "@/components/ui/toaster";
import { I18nProvider } from "@/lib/i18n";
import { MockStoreProvider } from "@/lib/mock/store";

type Options = RenderOptions & {
  route?: string;
  withStore?: boolean;
};

export function renderWithProviders(
  ui: ReactElement,
  { route = "/", withStore = true, ...options }: Options = {}
) {
  const body = withStore ? <MockStoreProvider>{ui}</MockStoreProvider> : ui;
  return render(
    <I18nProvider>
      <LazyMotion features={domMax} strict>
        <Toaster>
          <MemoryRouter initialEntries={[route]}>
            <Suspense fallback={null}>{body}</Suspense>
          </MemoryRouter>
        </Toaster>
      </LazyMotion>
    </I18nProvider>,
    options
  );
}
