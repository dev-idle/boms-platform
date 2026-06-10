import Link from "next/link";

import { Button } from "@/components/ui/button";
import { ROUTE } from "@/constants/routes";

import { CUSTOM_CAKE_BROWSE_HREF } from "../lib/storefront-links";

export function HomeCta() {
  return (
    <section className="storefront-section bg-blush">
      <div className="storefront-container">
        <div className="reveal rounded-card bg-surface px-8 py-14 text-center shadow-rest sm:px-12 sm:py-16">
          <h2 className="font-heading text-3xl font-medium tracking-tight text-ink sm:text-4xl">
            Planning a celebration?
          </h2>
          <p className="mx-auto mt-4 max-w-lg text-sm leading-relaxed text-ink-2 sm:text-base">
            Custom cakes and party trays available for advance order. Start with
            our menu, then pick your pickup date and time.
          </p>
          <div className="mt-10 flex flex-wrap justify-center gap-3">
            <Button asChild size="lg">
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
