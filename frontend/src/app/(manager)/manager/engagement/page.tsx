import { ManagerEngagement } from "@/features/manager";
import { PAGE_TITLES, pageTitle } from "@/lib/metadata/page-title";

export const metadata = pageTitle(PAGE_TITLES.engagement);

export default function ManagerEngagementPage() {
  return <ManagerEngagement />;
}
