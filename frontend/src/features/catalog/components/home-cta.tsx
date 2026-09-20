import Link from "next/link";

import { Button } from "@/components/ui/button";
import { ROUTE } from "@/constants/routes";

import { CUSTOM_CAKE_BROWSE_HREF } from "../lib/storefront-links";

export function HomeCta() {
  return (
    <section aria-labelledby="home-cta-heading" className="storefront-cta">
      <div className="storefront-container">
        <div className="storefront-cta__inner">
          <div className="storefront-cta__copy">
            <p className="storefront-cta__eyebrow">Bespoke</p>
            <h2 className="storefront-cta__heading" id="home-cta-heading">
              Planning something
              <br />
              worth celebrating?
            </h2>
            <p className="storefront-cta__lead">
              Bespoke cakes and party trays by advance order. Pick the design,
              pick the time — we handle the rest.
            </p>
          </div>
          <div className="storefront-cta__actions">
            <Button asChild showArrow size="lg">
              <Link href={CUSTOM_CAKE_BROWSE_HREF}>Start a custom order</Link>
            </Button>
            <Button asChild size="lg" variant="outline">
              <Link href={ROUTE.products}>Speak to a baker</Link>
            </Button>
          </div>
        </div>
      </div>
    </section>
  );
}
