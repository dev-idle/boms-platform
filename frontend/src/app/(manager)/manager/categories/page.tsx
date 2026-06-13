import { ManagerCategoriesTable } from "@/features/manager";
import { PAGE_TITLES, pageTitle } from "@/lib/metadata/page-title";

export const metadata = pageTitle(PAGE_TITLES.categories);

export default function ManagerCategoriesPage() {
  return <ManagerCategoriesTable />;
}
