import type { ReactNode } from "react";

import { BrandLogo } from "@/components/brand/brand-logo";
import { BRAND } from "@/constants/brand";

type AuthLayoutFrameProps = {
  children: ReactNode;
};

/** Split auth chrome — paper-3 statement panel left, centred 396px form block right. */
export function AuthLayoutFrame({ children }: AuthLayoutFrameProps) {
  const { accent, lead } = BRAND.authStatement;

  return (
    <div className="auth-page">
      <aside className="auth-page__aside">
        <div className="auth-page__aside-logo">
          <BrandLogo />
        </div>
        <div>
          <p className="auth-page__aside-eyebrow">{BRAND.establishedLabel}</p>
          <p className="auth-page__aside-statement">
            {lead.map((line) => (
              <span key={line}>
                {line}
                <br />
              </span>
            ))}
            <em>{accent}</em>
          </p>
          <dl className="auth-page__aside-meta">
            <div>
              <dt>Open daily</dt>
              <dd>{BRAND.pickupHours}</dd>
            </div>
            <div>
              <dt>Find us</dt>
              <dd>{BRAND.addressLine}</dd>
            </div>
          </dl>
        </div>
      </aside>
      <div className="auth-page-center">
        <section className="auth-page-form-panel">{children}</section>
      </div>
    </div>
  );
}
