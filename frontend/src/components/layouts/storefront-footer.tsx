import Link from "next/link";

import { BrandLogo } from "@/components/brand/brand-logo";
import { BRAND } from "@/constants/brand";
import { ROUTE } from "@/constants/routes";

export function StorefrontFooter() {
  return (
    <footer className="relative overflow-hidden bg-mint">
      <div
        aria-hidden="true"
        className="pointer-events-none absolute -right-8 bottom-0 select-none opacity-[0.04]"
      >
        <BrandLogo className="scale-[2.5] origin-bottom-right" linked={false} size="lg" />
      </div>

      <div className="storefront-container relative py-16 sm:py-20">
        <div className="grid gap-10 sm:grid-cols-2 lg:grid-cols-3">
          <div>
            <BrandLogo linked={false} size="md" />
            <p className="text-body mt-3">
              {BRAND.tagline}
            </p>
          </div>

          <div>
            <p className="text-body font-medium text-ink">Visit us</p>
            <address className="text-body mt-2 space-y-1 not-italic">
              <p>{BRAND.addressLine}</p>
              <p>
                <a
                  className="transition-colors duration-standard ease-default hover:text-matcha-500"
                  href={`mailto:${BRAND.contactEmail}`}
                >
                  {BRAND.contactEmail}
                </a>
              </p>
              <p>
                <a
                  className="transition-colors duration-standard ease-default hover:text-matcha-500"
                  href={`tel:${BRAND.contactPhone.replace(/\D/g, "")}`}
                >
                  {BRAND.contactPhone}
                </a>
              </p>
            </address>
          </div>

          <div>
            <p className="text-body font-medium text-ink">Hours</p>
            <ul className="text-body mt-2 space-y-1">
              <li>Mon–Fri · 7:00 AM – 6:00 PM</li>
              <li>Sat · 8:00 AM – 5:00 PM</li>
              <li>Sun · 8:00 AM – 2:00 PM</li>
            </ul>
          </div>
        </div>

        <div className="mt-12 flex flex-col gap-4 pt-8 sm:flex-row sm:items-center sm:justify-between">
          <p className="text-caption">
            © {BRAND.name}. All rights reserved.
          </p>
          <nav
            aria-label="Policies"
            className="flex flex-wrap gap-x-6 gap-y-2 text-caption"
          >
            <Link
              className="transition-colors duration-standard ease-default hover:text-matcha-500"
              href={ROUTE.home}
            >
              Privacy
            </Link>
            <Link
              className="transition-colors duration-standard ease-default hover:text-matcha-500"
              href={ROUTE.home}
            >
              Terms
            </Link>
            <Link
              className="transition-colors duration-standard ease-default hover:text-matcha-500"
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
