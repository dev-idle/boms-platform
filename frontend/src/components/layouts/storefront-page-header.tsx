import type { ReactNode } from "react";

import { StorefrontBackLink, type StorefrontBackTarget } from "./storefront-back-link";

type StorefrontPageHeaderProps = {
  /** The page one step up; a top-level page has none. */
  back?: StorefrontBackTarget;
  lead?: ReactNode;
  title: string;
};

/** The top of every storefront page: the way one step up when there is one, then the title. */
export function StorefrontPageHeader({ back, lead, title }: StorefrontPageHeaderProps) {
  return (
    <header className="storefront-page-head">
      {back ? <StorefrontBackLink {...back} /> : null}
      <h1 className="storefront-page-title">{title}</h1>
      {lead ? <p className="storefront-page-lead">{lead}</p> : null}
    </header>
  );
}
