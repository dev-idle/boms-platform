import type { ReactNode } from "react";

import { StaffShell } from "@/components/layouts/staff-shell";
import { StaffGate } from "@/features/auth";

export default function StaffLayout({ children }: { children: ReactNode }) {
  return (
    <StaffGate>
      <StaffShell>{children}</StaffShell>
    </StaffGate>
  );
}
