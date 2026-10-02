import { ManagerSalesReport } from "@/features/manager";
import { PAGE_TITLES, pageTitle } from "@/lib/metadata/page-title";

export const metadata = pageTitle(PAGE_TITLES.sales);

export default function ManagerReportsPage() {
  return <ManagerSalesReport />;
}
