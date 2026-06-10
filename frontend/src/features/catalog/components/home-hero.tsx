import Link from "next/link";

import { Button } from "@/components/ui/button";
import { ROUTE } from "@/constants/routes";

import { CUSTOM_CAKE_BROWSE_HREF } from "../lib/storefront-links";

/** Warm cream-toned patisserie — aligned with Quiet Luxury palette */
const HERO_IMAGE_URL =
  "https://images.unsplash.com/photo-1578985545062-69928b1d9587?auto=format&fit=crop&w=1200&q=80";

export function HomeHero() {
  return (
    <section className="storefront-section-hero bg-bg">
      <div className="storefront-container grid items-center gap-14 lg:grid-cols-2 lg:gap-20">
        <div className="reveal max-w-xl lg:py-4">
          <p className="storefront-eyebrow">Artisan patisserie · Pickup only</p>
          <div className="divider-gold mt-6" />
          <h1 className="text-display mt-8 lg:mt-10">
            <span className="font-heading italic text-rose-500">Sweetly</span>{" "}
            baked, ready when you are.
          </h1>
          <p className="mt-7 max-w-md text-base leading-[1.75] text-ink-2 sm:text-[1.0625rem]">
            Handcrafted pastries and celebration cakes — order online, pick up
            at your chosen time.
          </p>
          <div className="mt-12 flex flex-wrap gap-4">
            <Button asChild size="lg">
              <Link href={ROUTE.products}>Shop pastries</Link>
            </Button>
            <Button asChild size="lg" variant="outline">
              <Link href={CUSTOM_CAKE_BROWSE_HREF}>Custom cake</Link>
            </Button>
          </div>
        </div>

        <div className="reveal reveal-delay-1 relative lg:-mr-6 xl:-mr-12">
          <div className="rounded-[28px] bg-blush p-3 shadow-rest sm:p-4">
            <div className="overflow-hidden rounded-[22px]">
              {/* eslint-disable-next-line @next/next/no-img-element -- hero uses a curated external stock photo */}
              <img
                alt="An elegant tiered celebration cake in a patisserie display"
                className="aspect-[4/5] w-full object-cover sm:aspect-[5/4] lg:aspect-[4/5]"
                src={HERO_IMAGE_URL}
              />
            </div>
          </div>
        </div>
      </div>
    </section>
  );
}
