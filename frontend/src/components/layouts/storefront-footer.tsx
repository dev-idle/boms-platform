import Link from "next/link";

import { ROUTE } from "@/constants/routes";

export function StorefrontFooter() {
  return (
    <footer className="border-t border-border bg-surface">
      <div className="mx-auto max-w-7xl px-4 py-12 sm:px-6 lg:px-8">
        <div className="grid gap-10 sm:grid-cols-2 lg:grid-cols-3">
          <div>
            <p className="font-heading text-lg font-medium text-foreground">
              BOMS Bakery
            </p>
            <p className="mt-2 text-sm leading-relaxed text-muted">
              Handcrafted pastries and celebration cakes, made fresh daily for
              pickup.
            </p>
          </div>

          <div>
            <p className="text-sm font-medium text-foreground">Visit us</p>
            <address className="mt-2 space-y-1 text-sm not-italic text-muted">
              <p>123 Greige Lane</p>
              <p>
                <a
                  className="transition-colors duration-default ease-default hover:text-foreground"
                  href="mailto:hello@boms.example"
                >
                  hello@boms.example
                </a>
              </p>
              <p>
                <a
                  className="transition-colors duration-default ease-default hover:text-foreground"
                  href="tel:+15551234567"
                >
                  (555) 123-4567
                </a>
              </p>
            </address>
          </div>

          <div>
            <p className="text-sm font-medium text-foreground">Hours</p>
            <ul className="mt-2 space-y-1 text-sm text-muted">
              <li>Mon–Fri · 7:00 AM – 6:00 PM</li>
              <li>Sat · 8:00 AM – 5:00 PM</li>
              <li>Sun · 8:00 AM – 2:00 PM</li>
            </ul>
          </div>
        </div>

        <div className="mt-10 flex flex-col gap-4 border-t border-border pt-8 sm:flex-row sm:items-center sm:justify-between">
          <p className="text-xs text-subtle">
            © BOMS Bakery. All rights reserved.
          </p>
          <nav
            aria-label="Policies"
            className="flex flex-wrap gap-x-6 gap-y-2 text-xs text-muted"
          >
            <Link
              className="transition-colors duration-default ease-default hover:text-foreground"
              href={ROUTE.home}
            >
              Privacy
            </Link>
            <Link
              className="transition-colors duration-default ease-default hover:text-foreground"
              href={ROUTE.home}
            >
              Terms
            </Link>
            <Link
              className="transition-colors duration-default ease-default hover:text-foreground"
              href={ROUTE.home}
            >
              Pickup policy
            </Link>
          </nav>
        </div>
      </div>
    </footer>
  );
}
