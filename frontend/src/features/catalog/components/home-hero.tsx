import Link from "next/link";

import { Button } from "@/components/ui/button";
import { ROUTE } from "@/constants/routes";
import { STOREFRONT_HERO_IMAGE_URL } from "@/constants/storefront-imagery";
import { formatPriceCents } from "@/lib/validation/catalog";

import { CUSTOM_CAKE_BROWSE_HREF } from "../lib/storefront-links";

const HERO_STARTING_PRICE_CENTS = 450;

export function HomeHero() {
  return (
    <section className="storefront-section-hero bg-bg">
      <div className="storefront-container grid items-center gap-14 lg:grid-cols-2 lg:gap-20">
        <div className="reveal flex max-w-xl flex-col justify-center">
          <p className="text-overline">Artisan patisserie · Pickup only</p>
          <div className="divider-gold mt-6" />
          <h1 className="text-display mt-8 text-balance lg:mt-10">
            <span className="italic text-matcha-500">Sweetly</span> baked,
            <br />
            ready when you are.
          </h1>
          <p className="text-body-lg mt-7 max-w-md">
            Handcrafted pastries and celebration cakes — order online, pick up
            at your chosen time.
          </p>
          <div className="mt-10 flex flex-wrap items-center gap-x-6 gap-y-3">
            <div className="flex flex-wrap gap-2.5">
              <Button asChild showArrow size="lg">
                <Link href={ROUTE.products}>Shop pastries</Link>
              </Button>
              <Button asChild size="lg" variant="outline">
                <Link href={CUSTOM_CAKE_BROWSE_HREF}>Custom cake</Link>
              </Button>
            </div>
            <p className="font-display text-lg text-gold-700">
              From {formatPriceCents(HERO_STARTING_PRICE_CENTS)}
            </p>
          </div>
        </div>

        <div className="reveal reveal-delay-1 relative self-center lg:-mr-6 xl:-mr-12">
          <div className="rounded-[28px] bg-mint p-[10px] shadow-hover">
            <div className="overflow-hidden rounded-[22px]">
              {/* eslint-disable-next-line @next/next/no-img-element -- hero uses a curated external stock photo */}
              <img
                alt="Matcha cream puffs with whipped cream on a speckled plate"
                className="aspect-[4/5] max-h-[78vh] w-full object-cover object-[center_42%] sm:aspect-[5/4] lg:aspect-[4/5]"
                src={STOREFRONT_HERO_IMAGE_URL}
              />
            </div>
          </div>
        </div>
      </div>
    </section>
  );
}
