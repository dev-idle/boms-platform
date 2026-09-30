import Link from "next/link";
import type { ReactNode } from "react";

import { StorefrontPageHeader } from "@/components/layouts/storefront-page-header";
import { POLICIES_EFFECTIVE, TERMS_VERSION } from "@/constants/policies";
import { ROUTE } from "@/constants/routes";
import { PAGE_TITLES } from "@/lib/metadata/page-title";

const POLICY_LINKS = [
  { key: "terms", href: ROUTE.terms, label: PAGE_TITLES.terms },
  { key: "privacy", href: ROUTE.privacy, label: PAGE_TITLES.privacy },
  { key: "refunds", href: ROUTE.refundPolicy, label: PAGE_TITLES.refundPolicy },
] as const;

type PolicyDocumentProps = {
  children: ReactNode;
  current: (typeof POLICY_LINKS)[number]["key"];
  lead: string;
  title: string;
};

/** One customer policy: a reading column under the page header, beside its sibling policies. */
export function PolicyDocument({ children, current, lead, title }: PolicyDocumentProps) {
  return (
    <section className="storefront-section">
      <div className="storefront-container storefront-policy">
        <StorefrontPageHeader eyebrow="Policies" lead={lead} title={title} />
        <div className="storefront-policy__meta">
          <p>
            Effective {POLICIES_EFFECTIVE} · Version{" "}
            <span className="storefront-policy__version">{TERMS_VERSION}</span>
          </p>
          <nav aria-label="Customer policies" className="storefront-policy__nav">
            {POLICY_LINKS.map((link) => (
              <Link
                key={link.key}
                aria-current={link.key === current ? "page" : undefined}
                href={link.href}
              >
                {link.label}
              </Link>
            ))}
          </nav>
        </div>
        <div className="storefront-policy__body">{children}</div>
      </div>
    </section>
  );
}
