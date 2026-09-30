import { ForgotPasswordForm } from "@/features/auth";
import { PAGE_TITLES, pageTitle } from "@/lib/metadata/page-title";

export const metadata = pageTitle(PAGE_TITLES.forgotPassword);

export default function ForgotPasswordPage() {
  return <ForgotPasswordForm />;
}
