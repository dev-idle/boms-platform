import Link from "next/link";

import { BRAND } from "@/constants/brand";
import { Button } from "@/components/ui/button";
import { ROUTE } from "@/constants/routes";
import { STOREFRONT_HERO_IMAGE_URL } from "@/constants/storefront-imagery";

import { CUSTOM_CAKE_BROWSE_HREF } from "../lib/storefront-links";

export function HomeHero() {
  return (
    <section
      aria-labelledby="home-hero-heading"
      className="storefront-hero"
    >
      <div className="storefront-container storefront-hero__grid">
        <div className="storefront-hero__copy">
          <p className="text-overline">Artisan patisserie · Pickup only</p>
          <h1 className="storefront-hero__heading" id="home-hero-heading">
            Sweetly baked,
            <br />
            <span className="storefront-hero__accent">ready</span> when you are.
          </h1>
          <div aria-hidden="true" className="storefront-hero__rule" />
          <p className="storefront-hero__lead">
            Handcrafted pastries and celebration cakes, baked in small batches
            each morning. Order online and choose the pickup window that suits
            you.
          </p>
          <div className="storefront-hero__actions">
            <Button asChild showArrow size="lg">
              <Link href={ROUTE.products}>Shop the collection</Link>
            </Button>
            <Button asChild size="lg" variant="outline">
              <Link href={CUSTOM_CAKE_BROWSE_HREF}>Custom cake</Link>
            </Button>
          </div>
        </div>

        <div className="storefront-hero__visual">
          <div className="storefront-hero__media">
            {/* eslint-disable-next-line @next/next/no-img-element -- hero uses a curated self-hosted photo */}
            <img
              alt=""
              aria-hidden="true"
              className="storefront-hero__image"
              fetchPriority="high"
              src={STOREFRONT_HERO_IMAGE_URL}
            />
          </div>
          <p className="storefront-hero__stat">
            <span className="storefront-hero__stat-value">
              {BRAND.yearsInBusiness}
            </span>
            <span className="storefront-hero__stat-label">years of baking</span>
          </p>
        </div>
      </div>
    </section>
  );
}
