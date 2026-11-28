import { cacheLife } from "next/cache";
import Link from "next/link";

import { BrandLogo } from "@/components/brand/brand-logo";
import { BRAND, BRAND_PHONE_TEL_HREF } from "@/constants/brand";
import { ROUTE } from "@/constants/routes";
import { StoreHours } from "@/features/customer";
import { PAGE_TITLES } from "@/lib/metadata/page-title";

import { StorefrontFooterShopLinks } from "./storefront-footer-shop-links";

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
          <div>
            <BrandLogo linked={false} size="md" />
            <p className="storefront-footer__tagline">{BRAND.tagline}</p>
          </div>

          <div>
            <p className="storefront-footer__label">Shop</p>
            <StorefrontFooterShopLinks />
          </div>

          <div>
            <p className="storefront-footer__label">Visit us</p>
            <address className="storefront-footer__address">
              <span>{BRAND.addressLine}</span>
              <a href={`mailto:${BRAND.contactEmail}`}>{BRAND.contactEmail}</a>
              <a href={BRAND_PHONE_TEL_HREF}>{BRAND.contactPhone}</a>
            </address>
          </div>

          <div>
            <p className="storefront-footer__label">Hours</p>
            <ul className="storefront-footer__hours">
              <li>Every day</li>
              <li className="storefront-footer__hours-time">
                <StoreHours />
              </li>
              <li className="storefront-footer__hours-note">Pickup only, no delivery</li>
            </ul>
          </div>
        </div>

        <div className="storefront-footer__bar">
          <p className="storefront-footer__copyright">
            © {await getCurrentYear()} {BRAND.name}. All rights reserved.
          </p>
          <nav aria-label="Policies" className="storefront-footer__policies">
            <Link href={ROUTE.privacy}>{PAGE_TITLES.privacy}</Link>
            <Link href={ROUTE.terms}>{PAGE_TITLES.terms}</Link>
            <Link href={ROUTE.refundPolicy}>{PAGE_TITLES.refundPolicy}</Link>
          </nav>
        </div>
      </div>
    </footer>
  );
}
