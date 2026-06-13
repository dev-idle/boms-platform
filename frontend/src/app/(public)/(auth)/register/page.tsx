import { RegisterForm } from "@/features/auth";
import { PAGE_TITLES, pageTitle } from "@/lib/metadata/page-title";

export const metadata = pageTitle(PAGE_TITLES.createAccount);

export default function RegisterPage() {
  return <RegisterForm />;
}
