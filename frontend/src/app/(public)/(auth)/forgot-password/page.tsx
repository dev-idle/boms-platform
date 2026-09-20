import { Button } from "@/components/ui/button";
import { BRAND } from "@/constants/brand";
import { ROUTE } from "@/constants/routes";
import { AuthFormShell } from "@/features/auth";
import { PAGE_TITLES, pageTitle } from "@/lib/metadata/page-title";

export const metadata = pageTitle(PAGE_TITLES.forgotPassword);

export default function ForgotPasswordPage() {
  return (
    <AuthFormShell
      description={`Email us at ${BRAND.contactEmail} and we will reset your password.`}
      footer={{
        href: ROUTE.login,
        linkLabel: "Sign in",
        prompt: "Remember your password?",
      }}
      title={PAGE_TITLES.forgotPassword}
    >
      <Button asChild className="w-full" variant="outline">
        <a
          href={`mailto:${BRAND.contactEmail}?subject=${encodeURIComponent("Password reset request")}`}
        >
          Email {BRAND.contactEmail}
        </a>
      </Button>
    </AuthFormShell>
  );
}
