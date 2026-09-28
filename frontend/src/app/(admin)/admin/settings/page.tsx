import { AdminStoreSettings } from "@/features/admin";
import { PAGE_TITLES, pageTitle } from "@/lib/metadata/page-title";

export const metadata = pageTitle(PAGE_TITLES.settings);

export default function AdminSettingsPage() {
  return <AdminStoreSettings />;
}
