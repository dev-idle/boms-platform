import type { ReactNode } from "react";

import { BrandLogo } from "@/components/brand/brand-logo";

type AuthLayoutFrameProps = {
  children: ReactNode;
};

/** Centered single-column auth chrome — logo persists across child routes. */
export function AuthLayoutFrame({ children }: AuthLayoutFrameProps) {
  return (
    <div className="auth-page">
      <div className="auth-page-center">
        <div className="auth-page-logo">
          <BrandLogo size="header" />
        </div>
        <section className="auth-page-form-panel">{children}</section>
      </div>
    </div>
  );
}
