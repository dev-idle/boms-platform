import { ManagerPromotionsTable } from "@/features/manager";
import { PAGE_TITLES, pageTitle } from "@/lib/metadata/page-title";

export const metadata = pageTitle(PAGE_TITLES.promotions);

export default function ManagerPromotionsPage() {
  return <ManagerPromotionsTable />;
}
