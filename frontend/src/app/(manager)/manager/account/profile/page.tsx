import { USER_ROLE } from "@/constants/roles";
import {
  DashboardAccountProfileLayout,
  DashboardProfileSection,
  OperationalAccountProfileForm,
} from "@/features/user";

export default function ManagerAccountProfilePage() {
  return (
    <DashboardAccountProfileLayout
      description="Update your contact details. Team assignments are managed by an admin."
      title="Manager profile"
    >
      <DashboardProfileSection id="manager-profile-details" title="Account details">
        <OperationalAccountProfileForm
          expectedRole={USER_ROLE.manager}
          roleLabel="Manager"
        />
      </DashboardProfileSection>
    </DashboardAccountProfileLayout>
  );
}
