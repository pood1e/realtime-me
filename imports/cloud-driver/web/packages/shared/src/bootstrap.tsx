import { StrictMode, type ReactElement } from "react";
import { createRoot } from "react-dom/client";

import { AppProviders } from "./components/app-providers";

export function mountApp(app: ReactElement, rootId = "root"): void {
  const root = document.getElementById(rootId);
  if (!root) throw new Error(`Missing application root #${rootId}.`);
  createRoot(root).render(
    <StrictMode>
      <AppProviders>{app}</AppProviders>
    </StrictMode>,
  );
}
