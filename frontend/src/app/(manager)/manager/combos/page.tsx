import { ManagerCombosTable } from "@/features/manager";
import { PAGE_TITLES, pageTitle } from "@/lib/metadata/page-title";

export const metadata = pageTitle(PAGE_TITLES.combos);

export default function ManagerCombosPage() {
  return <ManagerCombosTable />;
}
