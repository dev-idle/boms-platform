import Link from "next/link";

import { BrandLogo } from "@/components/brand/brand-logo";
import { ThemeScope } from "@/components/theme/theme-scope";
import { Button } from "@/components/ui/button";
import { APP_THEME } from "@/constants/themes";
import { ROUTE } from "@/constants/routes";

export default function NotFound() {
  return (
    <ThemeScope theme={APP_THEME.storefront}>
      {/* Outside the storefront layout: the logo is the way back the header would give. */}
      <div className="storefront-container storefront-page grid justify-items-center gap-10">
        <BrandLogo size="md" />
        <div className="storefront-state w-full">
          <h1 className="storefront-state__title">Page not found</h1>
          <p className="storefront-state__message">The link may be mistyped, or the page has moved.</p>
          <div className="storefront-state__action">
            <Button asChild>
              <Link href={ROUTE.home}>Back to home</Link>
            </Button>
          </div>
        </div>
      </div>
    </ThemeScope>
  );
}
