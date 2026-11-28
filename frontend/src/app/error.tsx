"use client";

import Link from "next/link";
import { useEffect } from "react";

import { BrandLogo } from "@/components/brand/brand-logo";
import { ThemeScope } from "@/components/theme/theme-scope";
import { Button } from "@/components/ui/button";
import { APP_THEME } from "@/constants/themes";
import { ROUTE } from "@/constants/routes";

export default function Error({
  error,
  reset,
}: {
  error: Error & { digest?: string };
  reset: () => void;
}) {
  useEffect(() => {
    console.error(error);
  }, [error]);

  return (
    <ThemeScope theme={APP_THEME.storefront}>
      {/* As the not-found page: the logo is the way back the header would give. */}
      <div className="storefront-container storefront-page grid justify-items-center gap-10">
        <BrandLogo size="md" />
        <div className="storefront-state w-full" role="alert">
          <h1 className="storefront-state__title">Something went wrong</h1>
          <p className="storefront-state__message">Try again. If it keeps happening, come back in a few minutes.</p>
          <div className="storefront-state__action flex flex-wrap justify-center gap-3">
            <Button onClick={() => reset()} type="button">
              Try again
            </Button>
            <Button asChild variant="outline">
              {/* Every role reaches this page; the home page sends staff on to their own. */}
              <Link href={ROUTE.home}>Back to home</Link>
            </Button>
          </div>
        </div>
      </div>
    </ThemeScope>
  );
}
