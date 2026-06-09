import type { ReactNode } from "react";

import { AdminShell } from "@/components/layouts/admin-shell";
import { ThemeScope } from "@/components/theme/theme-scope";
import { APP_THEME } from "@/constants/themes";
import { AdminGate } from "@/features/auth";

export default function AdminLayout({ children }: { children: ReactNode }) {
  return (
    <AdminGate>
      <ThemeScope theme={APP_THEME.dashboard}>
        <AdminShell>{children}</AdminShell>
      </ThemeScope>
    </AdminGate>
  );
}
