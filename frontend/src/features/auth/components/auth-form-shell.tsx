import type { ReactNode, Ref } from "react";
import Link from "next/link";

import { StorefrontBrowseLink } from "@/components/layouts/storefront-browse-link";
import { STOREFRONT_NAV_COPY } from "@/constants/storefront-nav-copy";

type AuthFormShellFooter = {
  href: string;
  linkLabel: string;
  prompt: string;
};

type AuthFormShellProps = {
  children: ReactNode;
  description: string;
  eyebrow?: string;
  /** One secondary line: a prompt left, a single accent link right. */
  footer?: AuthFormShellFooter;
  /** Quiet line under the footer rule (e.g. the guest path). */
  footerNote?: ReactNode;
  title: string;
  /** Set on a view that replaces a form, to move focus to its heading. */
  titleRef?: Ref<HTMLHeadingElement>;
};

export function AuthFormShell({
  children,
  description,
  eyebrow = "Your account",
  footer,
  footerNote,
  title,
  titleRef,
}: AuthFormShellProps) {
  return (
    <div className="auth-page-form-stack">
      <StorefrontBrowseLink>{STOREFRONT_NAV_COPY.backToShop}</StorefrontBrowseLink>

      <header>
        <p className="auth-page-form-eyebrow">{eyebrow}</p>
        <h1 className="auth-page-form-title" ref={titleRef} tabIndex={titleRef ? -1 : undefined}>
          {title}
        </h1>
        <p className="auth-page-form-lead">{description}</p>
      </header>

      <div className="auth-page-form-body">{children}</div>

      {footer ? (
        <div className="auth-page-form-footer">
          <p className="auth-page-switch">{footer.prompt}</p>
          <Link className="auth-page-switch-link" href={footer.href}>
            {footer.linkLabel}
            <span aria-hidden="true">→</span>
          </Link>
        </div>
      ) : null}

      {footerNote ? <p className="auth-page-guest">{footerNote}</p> : null}
    </div>
  );
}
