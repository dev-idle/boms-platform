import {
  ChangePasswordForm,
  DashboardAccountProfileLayout,
  DashboardProfileSection,
} from "@/features/user";

export default function StaffAccountPasswordPage() {
  return (
    <DashboardAccountProfileLayout
      description="You will be signed out after updating your password."
      title="Change password"
    >
      <DashboardProfileSection id="staff-password" title="New password">
        <ChangePasswordForm />
      </DashboardProfileSection>
    </DashboardAccountProfileLayout>
  );
}
