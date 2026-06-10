import type { ReactNode } from "react";

import { StorefrontAuthShell } from "@/components/layouts/storefront-auth-shell";
import { AuthLayoutFrame } from "@/features/auth/components/auth-layout-frame";

export default function AuthLayout({ children }: { children: ReactNode }) {
  return (
    <StorefrontAuthShell>
      <AuthLayoutFrame>{children}</AuthLayoutFrame>
    </StorefrontAuthShell>
  );
}
