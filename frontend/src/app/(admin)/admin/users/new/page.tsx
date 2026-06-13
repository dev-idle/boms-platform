import { CreateOperationalUserForm } from "@/features/admin";
import { PAGE_TITLES, pageTitle } from "@/lib/metadata/page-title";

export const metadata = pageTitle(PAGE_TITLES.newUser);

export default function AdminUsersNewPage() {
  return <CreateOperationalUserForm />;
}
