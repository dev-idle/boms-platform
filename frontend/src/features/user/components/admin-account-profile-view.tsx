import { AdminAccountProfileForm } from "./admin-account-profile-form";
import {
  DashboardAccountProfileLayout,
  DashboardProfileSection,
} from "./dashboard-account-profile-layout";
import { ChangePasswordForm } from "./change-password-form";
import { PAGE_TITLES } from "@/lib/metadata/page-title";

/** Admin self-service profile + password (used by `/admin/account/profile`). */
export function AdminAccountProfileView() {
  return (
    <DashboardAccountProfileLayout
      description="Your admin account details and password."
      title={PAGE_TITLES.profile}
    >
      <DashboardProfileSection id="admin-profile-details" title="Account details">
        <AdminAccountProfileForm />
      </DashboardProfileSection>

      <DashboardProfileSection
        description="Changing your password signs out all active sessions."
        id="admin-profile-password"
        title="Password"
      >
        <ChangePasswordForm />
      </DashboardProfileSection>
    </DashboardAccountProfileLayout>
  );
}
