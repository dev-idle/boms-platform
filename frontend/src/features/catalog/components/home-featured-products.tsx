import Link from "next/link";
import type { ReactNode } from "react";

import { ROUTE } from "@/constants/routes";

type HomeFeaturedProductsProps = {
  /** The product grid — streamed, see `HomeFeaturedProductGrid`. */
  children: ReactNode;
};

export function HomeFeaturedProducts({ children }: HomeFeaturedProductsProps) {
  return (
    <section
      aria-labelledby="home-featured-heading"
      className="storefront-section border-t border-border"
    >
      <div className="storefront-container">
        <div className="storefront-section-header">
          <div className="storefront-section-header__copy">
            <p className="storefront-section-header__eyebrow">Selection</p>
            <h2
              className="storefront-section-header__heading"
              id="home-featured-heading"
            >
              Fresh{" "}
              <span className="storefront-section-header__accent">
                from the oven
              </span>
            </h2>
          </div>
          <Link className="storefront-section-header__link" href={ROUTE.products}>
            All products →
          </Link>
        </div>

        {children}
      </div>
    </section>
  );
}
