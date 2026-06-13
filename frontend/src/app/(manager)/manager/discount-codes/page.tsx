import { ManagerDiscountCodesTable } from "@/features/manager";
import { PAGE_TITLES, pageTitle } from "@/lib/metadata/page-title";

export const metadata = pageTitle(PAGE_TITLES.discountCodes);

export default function ManagerDiscountCodesPage() {
  return <ManagerDiscountCodesTable />;
}
