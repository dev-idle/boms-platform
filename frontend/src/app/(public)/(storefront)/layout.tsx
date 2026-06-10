import type { ReactNode } from "react";

import { StorefrontShell } from "@/components/layouts/storefront-shell";

export default function StorefrontLayout({ children }: { children: ReactNode }) {
  return <StorefrontShell>{children}</StorefrontShell>;
}
