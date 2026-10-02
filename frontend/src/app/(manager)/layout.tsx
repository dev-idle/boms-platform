import type { ReactNode } from "react";

import { ManagerShell } from "@/components/layouts/manager-shell";
import { ThemeScope } from "@/components/theme/theme-scope";
import { APP_THEME } from "@/constants/themes";
import { ManagerGate } from "@/features/auth";
import { ManagerLiveUpdates } from "@/features/manager";

export default function ManagerLayout({ children }: { children: ReactNode }) {
  return (
    <ManagerGate>
      <ManagerLiveUpdates />
      <ThemeScope theme={APP_THEME.dashboard}>
        <ManagerShell>{children}</ManagerShell>
      </ThemeScope>
    </ManagerGate>
  );
}
