import type { Metadata } from "next";

import { Button } from "@/components/ui/button";
import { BRAND } from "@/constants/brand";
import { ROUTE } from "@/constants/routes";
import { AuthFormShell } from "@/features/auth/components/auth-form-shell";
import { AuthInlineLink } from "@/features/auth/components/auth-inline-link";

export const metadata: Metadata = {
  title: "Forgot password",
};

export default function ForgotPasswordPage() {
  return (
    <AuthFormShell
      description={`Email us at ${BRAND.contactEmail} and we will reset your password.`}
      footer={
        <p className="auth-page-switch">
          Remember your password?{" "}
          <AuthInlineLink href={ROUTE.login}>Sign in</AuthInlineLink>
        </p>
      }
      title="Forgot password"
    >
      <div className="auth-forgot-actions">
        <Button asChild className="w-full" variant="outline">
          <a
            href={`mailto:${BRAND.contactEmail}?subject=${encodeURIComponent("Password reset request")}`}
          >
            Email {BRAND.contactEmail}
          </a>
        </Button>
        <AuthInlineLink className="text-caption text-muted" href={ROUTE.login}>
          Back to sign in
        </AuthInlineLink>
      </div>
    </AuthFormShell>
  );
}
