import Link from "next/link";

import { ThemeScope } from "@/components/theme/theme-scope";
import { APP_THEME } from "@/constants/themes";
import { ROUTE } from "@/constants/routes";

export default function NotFound() {
  return (
    <ThemeScope theme={APP_THEME.storefront}>
      <div className="mx-auto flex max-w-lg flex-col gap-4 px-6 py-24 text-center">
        <h1 className="font-heading text-xl font-medium text-foreground">
          Page not found
        </h1>
        <p className="text-sm text-muted">The requested resource does not exist.</p>
        <Link
          className="text-sm font-medium text-foreground underline underline-offset-4"
          href={ROUTE.home}
        >
          Return home
        </Link>
      </div>
    </ThemeScope>
  );
}
