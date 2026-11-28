import Link from "next/link";
import type { ReactNode } from "react";

import { StorefrontPageHeader } from "@/components/layouts/storefront-page-header";
import { POLICIES_EFFECTIVE, TERMS_VERSION } from "@/constants/policies";
import { ROUTE } from "@/constants/routes";

/** The menu's short names; each page carries its full title. */
const POLICY_LINKS = [
  { key: "terms", href: ROUTE.terms, label: "Terms" },
  { key: "privacy", href: ROUTE.privacy, label: "Privacy" },
  { key: "refunds", href: ROUTE.refundPolicy, label: "Refunds" },
] as const;

type PolicyDocumentProps = {
  children: ReactNode;
  current: (typeof POLICY_LINKS)[number]["key"];
  lead: string;
  title: string;
};

/** One customer policy: a reading column under its title, beside the menu of all three. */
export function PolicyDocument({ children, current, lead, title }: PolicyDocumentProps) {
  return (
    <div className="storefront-container storefront-page">
      <div className="storefront-subnav-shell">
        <StorefrontPageHeader lead={lead} title={title} />
        <nav aria-label="Customer policies" className="storefront-subnav">
          <ul className="storefront-subnav__list">
            {POLICY_LINKS.map((link) => (
              <li key={link.key}>
                <Link
                  aria-current={link.key === current ? "page" : undefined}
                  className="storefront-subnav__link"
                  href={link.href}
                >
                  {link.label}
                </Link>
              </li>
            ))}
          </ul>
        </nav>
        <div className="storefront-subnav-shell__main storefront-policy">
          <p className="storefront-policy__meta">
            Effective {POLICIES_EFFECTIVE} · Version{" "}
            <span className="storefront-policy__version">{TERMS_VERSION}</span>
          </p>
          {children}
        </div>
      </div>
    </div>
  );
}
