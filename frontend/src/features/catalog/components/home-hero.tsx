import Link from "next/link";

import { Button } from "@/components/ui/button";
import { ROUTE } from "@/constants/routes";

const HERO_IMAGE_URL =
  "https://images.unsplash.com/photo-1486427944299-d1955d23e34d?auto=format&fit=crop&w=1200&q=80";

export function HomeHero() {
  return (
    <section className="bg-bg py-24 sm:py-28 lg:py-32">
      <div className="mx-auto grid max-w-7xl items-center gap-12 px-4 sm:px-6 lg:grid-cols-2 lg:gap-16 lg:px-8">
        <div className="reveal max-w-xl">
          <div className="divider-gold" />
          <h1 className="text-display mt-8">
            <span className="font-heading italic text-rose-500">Sweetly</span>{" "}
            baked, ready when you are.
          </h1>
          <p className="mt-6 text-base leading-relaxed text-ink-2 sm:text-lg">
            Handcrafted pastries and celebration cakes — order online, pick up
            at your chosen time.
          </p>
          <div className="mt-10 flex flex-wrap gap-3">
            <Button asChild size="lg">
              <Link href={ROUTE.products}>Shop pastries</Link>
            </Button>
            <Button asChild size="lg" variant="outline">
              <Link href={ROUTE.register}>Create account</Link>
            </Button>
          </div>
        </div>

        <div className="reveal reveal-delay-1 relative lg:-mr-8 xl:-mr-16">
          <div className="overflow-hidden rounded-[24px] shadow-hover">
            {/* eslint-disable-next-line @next/next/no-img-element -- hero uses a curated external stock photo */}
            <img
              alt="Fresh pastries and cakes on a bakery counter"
              className="aspect-[4/5] w-full object-cover sm:aspect-[5/4] lg:aspect-square"
              src={HERO_IMAGE_URL}
            />
          </div>
        </div>
      </div>
    </section>
  );
}
