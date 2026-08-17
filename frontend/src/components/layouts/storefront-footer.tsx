import Link from "next/link";

import { BrandLogo } from "@/components/brand/brand-logo";
import { BRAND, BRAND_PHONE_TEL_HREF } from "@/constants/brand";
import { ROUTE } from "@/constants/routes";
import { CATALOG_COMBOS_HEADING_ID } from "@/features/catalog/lib/scroll-to-storefront-anchor";
import { CUSTOM_CAKE_BROWSE_HREF } from "@/features/catalog/lib/storefront-links";

const FOOTER_SHOP_LINKS = [
  { label: "All products", href: ROUTE.products },
  { label: "Combos", href: `${ROUTE.products}#${CATALOG_COMBOS_HEADING_ID}` },
  { label: "Custom cakes", href: CUSTOM_CAKE_BROWSE_HREF },
] as const;

export function StorefrontFooter() {
  return (
    <footer className="border-t border-border bg-bg">
      <div className="storefront-container py-10 sm:py-12">
        <div className="grid gap-10 sm:grid-cols-2 lg:grid-cols-4">
          <div>
            <BrandLogo linked={false} size="md" />
            <p className="text-body mt-3">
              {BRAND.tagline}
            </p>
          </div>

          <div>
            <p className="text-body font-medium text-ink">Shop</p>
            <ul className="text-body mt-2 space-y-1">
              {FOOTER_SHOP_LINKS.map((link) => (
                <li key={link.label}>
                  <Link
                    className="transition-colors duration-standard ease-default hover:text-matcha-500"
                    href={link.href}
                  >
                    {link.label}
                  </Link>
                </li>
              ))}
            </ul>
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
                  href={BRAND_PHONE_TEL_HREF}
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

        <div className="mt-10 flex flex-col gap-4 border-t border-border pt-6 sm:flex-row sm:items-center sm:justify-between">
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
