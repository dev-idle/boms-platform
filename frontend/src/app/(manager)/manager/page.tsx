import { ManagerOperations } from "@/features/manager";
import { PAGE_TITLES, pageTitle } from "@/lib/metadata/page-title";

export const metadata = pageTitle(PAGE_TITLES.dashboard);

export default function ManagerDashboardPage() {
  return <ManagerOperations />;
}
