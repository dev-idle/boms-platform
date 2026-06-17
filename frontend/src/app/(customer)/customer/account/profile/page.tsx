import { CustomerAccountView } from "@/features/user";
import { pageTitle, PAGE_TITLES } from "@/lib/metadata/page-title";

export const metadata = pageTitle(PAGE_TITLES.account);

export default function CustomerAccountProfilePage() {
  return <CustomerAccountView />;
}
