import Link from "next/link";

import { Button } from "@/components/ui/button";
import { ROUTE } from "@/constants/routes";

export function HomeBrandStory() {
  return (
    <section className="storefront-section bg-bg">
      <div className="storefront-container grid gap-12 lg:grid-cols-2 lg:items-center lg:gap-20">
        <div className="reveal relative aspect-[4/5] overflow-hidden rounded-card bg-blush sm:aspect-[5/4] lg:aspect-square">
          {/* eslint-disable-next-line @next/next/no-img-element -- brand story imagery */}
          <img
            alt="Pastry chef decorating a cake"
            className="size-full object-cover"
            src="https://images.unsplash.com/photo-1555507036-ab1f4038808a?auto=format&fit=crop&w=900&q=80"
          />
        </div>
        <div className="reveal reveal-delay-1">
          <div className="divider-gold" />
          <h2 className="mt-6 font-heading text-3xl font-medium tracking-tight text-ink sm:text-4xl">
            Baked with care, picked up on your schedule.
          </h2>
          <p className="mt-5 text-base leading-relaxed text-ink-2">
            Every loaf, tart, and tiered cake starts in our kitchen at dawn. We
            bake in small batches so what you take home is still warm with
            intention — never rushed, never mass-produced.
          </p>
          <p className="mt-4 text-base leading-relaxed text-ink-2">
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
