import Link from "next/link";
import type { ReactNode } from "react";

import { ROUTE } from "@/constants/routes";

type HomeCategoriesProps = {
  /** The category grid — streamed, see `HomeCategoryGrid`. */
  children: ReactNode;
};

export function HomeCategories({ children }: HomeCategoriesProps) {
  return (
    <section
      aria-labelledby="home-categories-heading"
      className="storefront-section storefront-section--tinted border-t border-border"
    >
      <div className="storefront-container">
        <div className="storefront-section-header storefront-section-header--ruled">
          <div className="storefront-section-header__copy">
            <p className="storefront-section-header__eyebrow">Categories</p>
            <h2
              className="storefront-section-header__heading"
              id="home-categories-heading"
            >
              From morning viennoiserie
              <br />
              to celebration cakes
            </h2>
          </div>
          <Link className="storefront-section-header__link" href={ROUTE.products}>
            View all →
          </Link>
        </div>

        {children}
      </div>
    </section>
  );
}
