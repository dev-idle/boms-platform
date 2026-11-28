import Link from "next/link";

import { Button } from "@/components/ui/button";

import { CUSTOM_CAKE_BROWSE_HREF } from "../lib/storefront-links";

export function HomeCta() {
  return (
    <section aria-labelledby="home-cta-heading" className="storefront-cta">
      <div className="storefront-container">
        <div className="storefront-cta__inner">
          <div className="storefront-cta__copy">
            <h2 className="storefront-cta__heading" id="home-cta-heading">
              Planning something
              <br />
              worth celebrating?
            </h2>
            <p className="storefront-cta__lead">
              Custom cakes by advance order. Choose the size, flavor, filling
              and decoration, then pick a pickup time.
            </p>
          </div>
          <div className="storefront-cta__actions">
            <Button asChild showArrow size="lg">
              <Link href={CUSTOM_CAKE_BROWSE_HREF}>Browse custom cakes</Link>
            </Button>
          </div>
        </div>
      </div>
    </section>
  );
}
