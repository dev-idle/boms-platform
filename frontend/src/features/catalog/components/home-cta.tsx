import Link from "next/link";

import { Button } from "@/components/ui/button";
import { ROUTE } from "@/constants/routes";

import { CUSTOM_CAKE_BROWSE_HREF } from "../lib/storefront-links";

export function HomeCta() {
  return (
    <section className="storefront-section bg-mint">
      <div className="storefront-container">
        <div className="reveal rounded-card bg-surface px-8 py-14 text-center shadow-rest sm:px-12 sm:py-16">
          <h2 className="text-h2">
            Planning a celebration?
          </h2>
          <p className="mx-auto mt-4 max-w-lg text-sm leading-relaxed text-muted sm:text-base">
            Custom cakes and party trays available for advance order. Browse
            our menu and check out when you are ready for pickup.
          </p>
          <div className="mt-10 flex flex-wrap justify-center gap-2.5">
            <Button asChild showArrow size="lg">
              <Link href={ROUTE.products}>Order for pickup</Link>
            </Button>
            <Button asChild size="lg" variant="outline">
              <Link href={CUSTOM_CAKE_BROWSE_HREF}>Custom cakes</Link>
            </Button>
          </div>
        </div>
      </div>
    </section>
  );
}
