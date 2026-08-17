import Link from "next/link";

import { Button } from "@/components/ui/button";
import { ROUTE } from "@/constants/routes";
import { STOREFRONT_HERO_IMAGE_URL } from "@/constants/storefront-imagery";

import { CUSTOM_CAKE_BROWSE_HREF } from "../lib/storefront-links";

export function HomeHero() {
  return (
    <section aria-labelledby="home-hero-heading" className="storefront-hero-banner">
      {/* eslint-disable-next-line @next/next/no-img-element -- hero uses a curated self-hosted photo */}
      <img
        alt=""
        aria-hidden="true"
        className="storefront-hero-banner__media"
        fetchPriority="high"
        src={STOREFRONT_HERO_IMAGE_URL}
      />
      <div className="storefront-container storefront-hero-banner__content">
        <div className="storefront-hero-banner__panel">
          <p className="text-overline">Artisan patisserie · Pickup only</p>
          <h1 className="text-display mt-4 text-balance" id="home-hero-heading">
            <span className="italic text-matcha-500">Sweetly</span> baked,
            ready when you are.
          </h1>
          <p className="text-body mt-4">
            Handcrafted pastries and celebration cakes — order online, pick up
            at your chosen time.
          </p>
          <div className="mt-6 flex flex-wrap gap-2.5">
            <Button asChild showArrow size="lg">
              <Link href={ROUTE.products}>Shop pastries</Link>
            </Button>
            <Button asChild size="lg" variant="outline">
              <Link href={CUSTOM_CAKE_BROWSE_HREF}>Custom cake</Link>
            </Button>
          </div>
        </div>
      </div>
    </section>
  );
}
