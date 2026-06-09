import type { ReactNode } from "react";

import { StorefrontShell } from "@/components/layouts/storefront-shell";
import { ThemeScope } from "@/components/theme/theme-scope";
import { APP_THEME } from "@/constants/themes";
import { CustomerGate } from "@/features/auth";

export default function CustomerLayout({ children }: { children: ReactNode }) {
  return (
    <CustomerGate>
      <ThemeScope theme={APP_THEME.storefront}>
        <StorefrontShell>
          <div className="mx-auto w-full max-w-7xl px-4 py-10 sm:px-6 lg:px-8">
            {children}
          </div>
        </StorefrontShell>
      </ThemeScope>
    </CustomerGate>
  );
}
