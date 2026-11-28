import { ManagerStoreSettings } from "@/features/manager";
import { PAGE_TITLES, pageTitle } from "@/lib/metadata/page-title";

export const metadata = pageTitle(PAGE_TITLES.settings);

export default function ManagerSettingsPage() {
  return <ManagerStoreSettings />;
}
