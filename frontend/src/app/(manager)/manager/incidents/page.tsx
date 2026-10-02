import { ManagerIncidents } from "@/features/manager";
import { PAGE_TITLES, pageTitle } from "@/lib/metadata/page-title";

export const metadata = pageTitle(PAGE_TITLES.incidents);

export default function ManagerIncidentsPage() {
  return <ManagerIncidents />;
}
