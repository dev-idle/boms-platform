import Link from "next/link";

import { Button } from "@/components/ui/button";
import { ROUTE } from "@/constants/routes";
import { STOREFRONT_BRAND_STORY_IMAGE_URL } from "@/constants/storefront-imagery";

export function HomeBrandStory() {
  return (
    <section className="storefront-section bg-bg">
      <div className="storefront-container grid gap-12 lg:grid-cols-2 lg:items-center lg:gap-20">
        <div className="reveal relative aspect-[4/5] overflow-hidden rounded-card bg-mint shadow-rest sm:aspect-[5/4] lg:aspect-square">
          {/* eslint-disable-next-line @next/next/no-img-element -- brand story imagery */}
          <img
            alt="Vanilla eclair with white icing on a bright backdrop"
            className="size-full object-cover object-center"
            src={STOREFRONT_BRAND_STORY_IMAGE_URL}
          />
        </div>
        <div className="reveal reveal-delay-1">
          <div className="divider-gold" />
          <h2 className="text-h2 mt-6">
            Baked with care, picked up on your schedule.
          </h2>
          <p className="mt-5 text-base leading-relaxed text-muted">
            Every loaf, tart, and tiered cake starts in our kitchen at dawn. We
            bake in small batches so what you take home is still warm with
            intention — never rushed, never mass-produced.
          </p>
          <p className="mt-4 text-base leading-relaxed text-muted">
            Order ahead, choose your pickup window, and we will have everything
            ready at the counter.
          </p>
          <div className="mt-8">
            <Button asChild variant="outline">
              <Link href={ROUTE.products}>Explore the menu</Link>
            </Button>
          </div>
        </div>
      </div>
    </section>
  );
}
