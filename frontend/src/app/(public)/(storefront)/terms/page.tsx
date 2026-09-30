import { TermsOfSale } from "@/features/legal";
import { PAGE_TITLES, pageTitle } from "@/lib/metadata/page-title";

export const instant = true;

export const metadata = pageTitle(PAGE_TITLES.terms, "The terms you accept when you sign up and order for pickup.");

export default function TermsPage() {
  return <TermsOfSale />;
}
