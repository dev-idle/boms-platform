import Link from "next/link";

import { Button } from "@/components/ui/button";
import { ROUTE } from "@/constants/routes";

const HERO_IMAGE_URL =
  "https://images.unsplash.com/photo-1486427944299-d1955d23e34d?auto=format&fit=crop&w=1920&q=80";

export function HomeHero() {
  return (
    <section className="relative isolate min-h-[min(72vh,640px)] overflow-hidden">
      {/* eslint-disable-next-line @next/next/no-img-element -- hero uses a curated external stock photo */}
      <img
        alt=""
        className="absolute inset-0 size-full object-cover"
        src={HERO_IMAGE_URL}
      />
      <div
        aria-hidden="true"
        className="absolute inset-0 bg-linear-to-r from-greige-900/75 via-greige-900/45 to-greige-900/20"
      />
      <div className="relative mx-auto flex min-h-[min(72vh,640px)] max-w-7xl items-center px-4 py-20 sm:px-6 lg:px-8">
        <div className="max-w-xl">
          <div className="h-px w-12 border-t border-accent" />
          <h1 className="mt-8 font-heading text-4xl font-medium leading-tight tracking-tight text-white sm:text-5xl lg:text-6xl">
            Freshly baked, ready when you are.
          </h1>
          <p className="mt-5 text-base leading-relaxed text-greige-100 sm:text-lg">
            Handcrafted pastries and celebration cakes — order online, pick up
            at your chosen time.
          </p>
          <div className="mt-8 flex flex-wrap gap-3">
            <Button asChild size="lg">
              <Link href={ROUTE.products}>Shop pastries</Link>
            </Button>
            <Button
              asChild
              className="border-white/30 bg-white/10 text-white hover:bg-white/20"
              size="lg"
              variant="outline"
            >
              <Link href={ROUTE.register}>Create account</Link>
            </Button>
          </div>
        </div>
      </div>
    </section>
  );
}
