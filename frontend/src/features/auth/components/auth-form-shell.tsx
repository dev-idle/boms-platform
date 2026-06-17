import type { ReactNode } from "react";

import { StorefrontBrowseLink } from "@/components/layouts/storefront-browse-link";

type AuthFormShellProps = {
  title: string;
  description: string;
  children: ReactNode;
  footer?: ReactNode;
};

export function AuthFormShell({
  title,
  description,
  children,
  footer,
}: AuthFormShellProps) {
  return (
    <div className="auth-page-form-stack">
      <div className="auth-page-form-exit">
        <StorefrontBrowseLink />
      </div>

      <header className="auth-page-form-intro">
        <h1 className="auth-page-form-title">{title}</h1>
        <p className="auth-page-form-lead">{description}</p>
      </header>

      <div className="auth-page-form-body">{children}</div>

      {footer ? <div className="auth-page-form-footer">{footer}</div> : null}
    </div>
  );
}
