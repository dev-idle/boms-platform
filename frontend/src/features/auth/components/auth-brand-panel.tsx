import { BrandLogo } from "@/components/brand/brand-logo";

import { AUTH_PATISSERIE_IMAGE_URL } from "@/constants/storefront-imagery";
import { AUTH_BRAND_DESCRIPTION, AuthBrandTitle } from "../lib/auth-brand-copy";

/** Persistent auth imagery — mounted in `(auth)/layout`, not per route. */
export function AuthBrandPanel() {
  return (
    <section className="auth-page-brand">
      <div className="auth-page-brand-stage">
        {/* eslint-disable-next-line @next/next/no-img-element -- curated stock photo for auth panel */}
        <img
          alt=""
          className="auth-page-brand-image"
          src={AUTH_PATISSERIE_IMAGE_URL}
        />
        <div className="auth-page-brand-scrim" />
        <div className="auth-page-brand-logo">
          <BrandLogo className="auth-brand-logo" size="header" />
        </div>
        <div className="auth-page-brand-copy">
          <div className="auth-page-brand-divider" />
          <h2 className="auth-page-brand-title">
            <AuthBrandTitle />
          </h2>
          <p className="auth-page-brand-description">{AUTH_BRAND_DESCRIPTION}</p>
        </div>
      </div>
    </section>
  );
}
