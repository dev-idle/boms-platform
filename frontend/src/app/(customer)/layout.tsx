import type { ReactNode } from "react";

import { StorefrontShell } from "@/components/layouts/storefront-shell";
import { StorefrontCustomerLayout } from "@/components/layouts/storefront-customer-layout";
import { ThemeScope } from "@/components/theme/theme-scope";
import { APP_THEME } from "@/constants/themes";
import { CustomerGate } from "@/features/auth";

export default function CustomerLayout({ children }: { children: ReactNode }) {
  return (
    <CustomerGate>
      <ThemeScope theme={APP_THEME.storefront}>
        <StorefrontShell>
          <div className="storefront-container storefront-customer-page">
            <StorefrontCustomerLayout>{children}</StorefrontCustomerLayout>
          </div>
        </StorefrontShell>
      </ThemeScope>
    </CustomerGate>
  );
}
