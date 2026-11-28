import type { ReactNode } from "react";

import { StorefrontShell } from "@/components/layouts/storefront-shell";
import { ThemeScope } from "@/components/theme/theme-scope";
import { APP_THEME } from "@/constants/themes";
import { CustomerGate } from "@/features/auth";

// The gate waits inside the storefront chrome, so restoring a session keeps the
// header and footer on screen instead of a blank page.
export default function CustomerLayout({ children }: { children: ReactNode }) {
  return (
    <ThemeScope theme={APP_THEME.storefront}>
      <StorefrontShell>
        <div className="storefront-container storefront-page">
          <CustomerGate>{children}</CustomerGate>
        </div>
      </StorefrontShell>
    </ThemeScope>
  );
}
