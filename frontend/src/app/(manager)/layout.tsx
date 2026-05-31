import type { ReactNode } from "react";

import { OperationalRoleShell } from "@/components/layouts/operational-role-shell";
import { ROUTE } from "@/constants/routes";
import { ManagerGate } from "@/features/auth";

export default function ManagerLayout({ children }: { children: ReactNode }) {
  return (
    <ManagerGate>
      <OperationalRoleShell
        roleLabel="Manager"
        labelClassName="text-sky-700 dark:text-sky-400"
        profileHref={ROUTE.manager.account.profile}
        passwordHref={ROUTE.manager.account.password}
      >
        {children}
      </OperationalRoleShell>
    </ManagerGate>
  );
}
