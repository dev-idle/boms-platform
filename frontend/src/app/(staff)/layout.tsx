import type { ReactNode } from "react";

import { StaffShell } from "@/components/layouts/staff-shell";
import { ThemeScope } from "@/components/theme/theme-scope";
import { APP_THEME } from "@/constants/themes";
import { StaffGate } from "@/features/auth";
import { StaffLiveUpdates } from "@/features/staff";

export default function StaffLayout({ children }: { children: ReactNode }) {
  return (
    <StaffGate>
      <StaffLiveUpdates />
      <ThemeScope theme={APP_THEME.dashboard}>
        <StaffShell>{children}</StaffShell>
      </ThemeScope>
    </StaffGate>
  );
}
