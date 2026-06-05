import type { ReactNode } from "react";

import { ManagerShell } from "@/components/layouts/manager-shell";
import { ManagerGate } from "@/features/auth";

export default function ManagerLayout({ children }: { children: ReactNode }) {
  return (
    <ManagerGate>
      <ManagerShell>{children}</ManagerShell>
    </ManagerGate>
  );
}
