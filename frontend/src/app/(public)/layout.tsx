import type { ReactNode } from "react";

import { PublicSessionGate } from "@/features/auth";

export default function PublicLayout({ children }: { children: ReactNode }) {
  return <PublicSessionGate>{children}</PublicSessionGate>;
}
