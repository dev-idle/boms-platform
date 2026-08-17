import { BRAND, BRAND_PHONE_TEL_HREF } from "@/constants/brand";

/** Utility strip above the sticky header — scrolls away with the page. */
export function StorefrontTopbar() {
  return (
    <div className="storefront-topbar">
      <div className="storefront-container storefront-topbar__inner">
        <p className="storefront-topbar__note">
          Freshly baked daily — order online, pick up in store
        </p>
        <a className="storefront-topbar__phone" href={BRAND_PHONE_TEL_HREF}>
          {BRAND.contactPhone}
        </a>
      </div>
    </div>
  );
}
