import { USER_ROLE } from "@/constants/roles";
import { OperationalAccountProfileView } from "@/features/user";
import { PAGE_TITLES, pageTitle } from "@/lib/metadata/page-title";

export const metadata = pageTitle(PAGE_TITLES.profile);

export default function StaffAccountProfilePage() {
  return <OperationalAccountProfileView role={USER_ROLE.staff} />;
}
