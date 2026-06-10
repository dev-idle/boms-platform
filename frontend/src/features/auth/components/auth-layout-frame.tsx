import type { ReactNode } from "react";

import { AuthBrandPanel } from "./auth-brand-panel";

type AuthLayoutFrameProps = {
  children: ReactNode;
};

/** Two-column auth chrome — brand panel persists across child routes. */
export function AuthLayoutFrame({ children }: AuthLayoutFrameProps) {
  return (
    <div className="auth-page">
      <div className="auth-page-grid">
        <AuthBrandPanel />
        <section className="auth-page-form-panel">{children}</section>
      </div>
    </div>
  );
}
