import { ResetPasswordForm } from "@/features/auth";
import { PAGE_TITLES, pageTitle } from "@/lib/metadata/page-title";

export const metadata = pageTitle(PAGE_TITLES.resetPassword);

export default function ResetPasswordPage() {
  return <ResetPasswordForm />;
}
