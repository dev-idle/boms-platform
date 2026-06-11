import { USER_ROLE } from "@/constants/roles";
import {
  DashboardAccountProfileLayout,
  DashboardProfileSection,
  OperationalAccountProfileForm,
} from "@/features/user";

export default function BakerAccountProfilePage() {
  return (
    <DashboardAccountProfileLayout
      description="Update your contact details. Production assignments are managed by an admin."
      title="Baker profile"
    >
      <DashboardProfileSection id="baker-profile-details" title="Account details">
        <OperationalAccountProfileForm
          expectedRole={USER_ROLE.baker}
          roleLabel="Baker"
        />
      </DashboardProfileSection>
    </DashboardAccountProfileLayout>
  );
}
