import { cacheLife } from "next/cache";
import Link from "next/link";

import { BrandLogo } from "@/components/brand/brand-logo";
import { BRAND, BRAND_PHONE_TEL_HREF } from "@/constants/brand";
import { ROUTE } from "@/constants/routes";
import { CATALOG_COMBOS_HEADING_ID, CUSTOM_CAKE_BROWSE_HREF } from "@/features/catalog";

const FOOTER_SHOP_LINKS = [
  { label: "All products", href: ROUTE.products },
  { label: "Combos", href: `${ROUTE.products}#${CATALOG_COMBOS_HEADING_ID}` },
  { label: "Custom cakes", href: CUSTOM_CAKE_BROWSE_HREF },
] as const;

async function getCurrentYear(): Promise<number> {
  "use cache";
  cacheLife("max");
  return new Date().getFullYear();
}

export async function StorefrontFooter() {
  return (
    <footer className="storefront-footer">
      <div className="storefront-container storefront-footer__inner">
        <div className="storefront-footer__grid">
          <div className="storefront-footer__brand">
            <BrandLogo linked={false} size="md" />
            <p className="storefront-footer__tagline">{BRAND.tagline}</p>
          </div>

          <div className="storefront-footer__column">
            <p className="storefront-footer__label">Shop</p>
            <ul className="storefront-footer__links">
              {FOOTER_SHOP_LINKS.map((link) => (
                <li key={link.label}>
                  <Link href={link.href}>{link.label}</Link>
                </li>
              ))}
            </ul>
          </div>

          <div className="storefront-footer__column">
            <p className="storefront-footer__label">Visit us</p>
            <address className="storefront-footer__address">
              <span>{BRAND.addressLine}</span>
              <a href={`mailto:${BRAND.contactEmail}`}>{BRAND.contactEmail}</a>
              <a href={BRAND_PHONE_TEL_HREF}>{BRAND.contactPhone}</a>
            </address>
          </div>

          <div className="storefront-footer__column">
            <p className="storefront-footer__label">Hours</p>
            <ul className="storefront-footer__hours">
              <li>{BRAND.pickupHoursFooter}</li>
              <li className="storefront-footer__hours-note">
                {BRAND.kitchenCloseNote}
              </li>
            </ul>
          </div>
        </div>

        <div className="storefront-footer__bar">
          <p className="storefront-footer__copyright">
            © {await getCurrentYear()} {BRAND.name}. All rights reserved.
          </p>
          <nav aria-label="Policies" className="storefront-footer__policies">
            <Link href={ROUTE.home}>Privacy</Link>
            <Link href={ROUTE.home}>Terms</Link>
            <Link href={ROUTE.home}>Pickup policy</Link>
          </nav>
        </div>
      </div>
    </footer>
  );
}
