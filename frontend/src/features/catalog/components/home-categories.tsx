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
        <div className="storefront-section-header">
          <div className="storefront-section-header__copy">
            <h2
              className="storefront-section-header__heading"
              id="home-categories-heading"
            >
              Shop by <span className="storefront-section-header__accent">category</span>
            </h2>
          </div>
          <Link className="storefront-section-header__link" href={ROUTE.products}>
            All products
          </Link>
        </div>

        {children}
      </div>
    </section>
  );
}
