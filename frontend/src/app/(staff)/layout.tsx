import type { ReactNode } from "react";

import { OperationalRoleShell } from "@/components/layouts/operational-role-shell";
import { ROUTE } from "@/constants/routes";
import { StaffGate } from "@/features/auth";

export default function StaffLayout({ children }: { children: ReactNode }) {
  return (
    <StaffGate>
      <OperationalRoleShell
        roleLabel="Staff"
        labelClassName="text-zinc-500"
        profileHref={ROUTE.staff.account.profile}
        passwordHref={ROUTE.staff.account.password}
      >
        {children}
      </OperationalRoleShell>
    </StaffGate>
  );
}
