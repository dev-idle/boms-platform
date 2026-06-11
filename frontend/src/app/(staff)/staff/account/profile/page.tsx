import { USER_ROLE } from "@/constants/roles";
import {
  DashboardAccountProfileLayout,
  DashboardProfileSection,
  OperationalAccountProfileForm,
} from "@/features/user";

export default function StaffAccountProfilePage() {
  return (
    <DashboardAccountProfileLayout
      description="Update your contact details. Operational fields are managed by an admin."
      title="Staff profile"
    >
      <DashboardProfileSection id="staff-profile-details" title="Account details">
        <OperationalAccountProfileForm
          expectedRole={USER_ROLE.staff}
          roleLabel="Staff"
        />
      </DashboardProfileSection>
    </DashboardAccountProfileLayout>
  );
}
