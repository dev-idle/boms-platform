import Link from "next/link";

import { Button } from "@/components/ui/button";
import { ROUTE } from "@/constants/routes";

export function HomeCta() {
  return (
    <section className="bg-blush py-24 sm:py-28">
      <div className="mx-auto max-w-7xl px-4 sm:px-6 lg:px-8">
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
              <Link href={ROUTE.register}>Sign up free</Link>
            </Button>
          </div>
        </div>
      </div>
    </section>
  );
}
