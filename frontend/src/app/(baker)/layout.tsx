import type { ReactNode } from "react";

import { OperationalRoleShell } from "@/components/layouts/operational-role-shell";
import { ROUTE } from "@/constants/routes";
import { BakerGate } from "@/features/auth";

export default function BakerLayout({ children }: { children: ReactNode }) {
  return (
    <BakerGate>
      <OperationalRoleShell
        roleLabel="Baker"
        labelClassName="text-amber-700 dark:text-amber-400"
        profileHref={ROUTE.baker.account.profile}
        passwordHref={ROUTE.baker.account.password}
      >
        {children}
      </OperationalRoleShell>
    </BakerGate>
  );
}
