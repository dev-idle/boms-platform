import { AdminAccountProfileView } from "@/features/user";
import { PAGE_TITLES, pageTitle } from "@/lib/metadata/page-title";

export const metadata = pageTitle(PAGE_TITLES.profile);

export default function AdminAccountProfilePage() {
  return <AdminAccountProfileView />;
}
