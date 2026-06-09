import Link from "next/link";

import { Button } from "@/components/ui/button";
import { ROUTE } from "@/constants/routes";

export function HomeCta() {
  return (
    <section className="border-t border-border py-16 sm:py-20">
      <div className="mx-auto max-w-7xl px-4 sm:px-6 lg:px-8">
        <div className="rounded-lg border border-border bg-primary-subtle px-8 py-12 text-center sm:px-12 sm:py-16">
          <h2 className="font-heading text-2xl font-medium tracking-tight text-foreground sm:text-3xl">
            Planning a celebration?
          </h2>
          <p className="mx-auto mt-3 max-w-lg text-sm leading-relaxed text-muted sm:text-base">
            Custom cakes and party trays available for advance order. Start with
            our menu, then pick your pickup date and time.
          </p>
          <div className="mt-8 flex flex-wrap justify-center gap-3">
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
