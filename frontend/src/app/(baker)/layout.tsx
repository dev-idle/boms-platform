import type { ReactNode } from "react";

import { OperationalRoleShell } from "@/components/layouts/operational-role-shell";
import { ThemeScope } from "@/components/theme/theme-scope";
import { APP_THEME } from "@/constants/themes";
import { ROUTE } from "@/constants/routes";
import { BakerGate } from "@/features/auth";

export default function BakerLayout({ children }: { children: ReactNode }) {
  return (
    <BakerGate>
      <ThemeScope theme={APP_THEME.dashboard}>
        <OperationalRoleShell
          roleLabel="Baker"
          profileHref={ROUTE.baker.account.profile}
          passwordHref={ROUTE.baker.account.password}
        >
          {children}
        </OperationalRoleShell>
      </ThemeScope>
    </BakerGate>
  );
}
