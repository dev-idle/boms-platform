import { VerifyEmailView } from "@/features/auth";
import { PAGE_TITLES, pageTitle } from "@/lib/metadata/page-title";

export const metadata = pageTitle(PAGE_TITLES.verifyEmail);

export default function VerifyEmailPage() {
  return <VerifyEmailView />;
}
