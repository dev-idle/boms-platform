import type { ReactNode } from "react";

import { AdminShell } from "@/components/layouts/admin-shell";
import { AdminGate } from "@/features/auth";

export default function AdminLayout({ children }: { children: ReactNode }) {
  return (
    <AdminGate>
      <AdminShell>{children}</AdminShell>
    </AdminGate>
  );
}
