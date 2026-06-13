import { AdminUsersTable } from "@/features/admin";
import { PAGE_TITLES, pageTitle } from "@/lib/metadata/page-title";

export const metadata = pageTitle(PAGE_TITLES.users);

export default function AdminUsersPage() {
  return <AdminUsersTable />;
}
