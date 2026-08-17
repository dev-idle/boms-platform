import Link from "next/link";

import { Button } from "@/components/ui/button";
import { ROUTE } from "@/constants/routes";

import { CUSTOM_CAKE_BROWSE_HREF } from "../lib/storefront-links";

export function HomeCta() {
  return (
    <section className="storefront-section border-t border-border bg-bg">
      <div className="storefront-container text-center">
        <h2 className="text-h2">Planning a celebration?</h2>
        <p className="mx-auto mt-3 max-w-lg text-sm leading-relaxed text-muted sm:text-base">
          Custom cakes and party trays available for advance order. Browse our
          menu and check out when you are ready for pickup.
        </p>
        <div className="mt-7 flex flex-wrap justify-center gap-2.5">
          <Button asChild showArrow size="lg">
            <Link href={ROUTE.products}>Order for pickup</Link>
          </Button>
          <Button asChild size="lg" variant="outline">
            <Link href={CUSTOM_CAKE_BROWSE_HREF}>Custom cakes</Link>
          </Button>
        </div>
      </div>
    </section>
  );
}
