import type { ReactNode } from "react";

import { ThemeScope } from "@/components/theme/theme-scope";
import { APP_THEME } from "@/constants/themes";
import { PublicSessionGate } from "@/features/auth";

export default function PublicLayout({ children }: { children: ReactNode }) {
  return (
    <PublicSessionGate>
      <ThemeScope theme={APP_THEME.storefront}>{children}</ThemeScope>
    </PublicSessionGate>
  );
}
