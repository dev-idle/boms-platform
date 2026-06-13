import { ManagerProductsTable } from "@/features/manager";
import { PAGE_TITLES, pageTitle } from "@/lib/metadata/page-title";

export const metadata = pageTitle(PAGE_TITLES.products);

export default function ManagerProductsPage() {
  return <ManagerProductsTable />;
}
