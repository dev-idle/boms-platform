import { DashboardPageHeader } from "@/components/ui/dashboard-page-header";
import { PAGE_TITLES, pageTitle } from "@/lib/metadata/page-title";

export const metadata = pageTitle(PAGE_TITLES.dashboard);

export default function AdminDashboardPage() {
  return (
    <DashboardPageHeader
      description="Welcome to the admin workspace. Use the sidebar to manage users or update your profile."
      title={PAGE_TITLES.dashboard}
    />
  );
}
