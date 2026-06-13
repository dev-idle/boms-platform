import { USER_ROLE } from "@/constants/roles";
import { PAGE_TITLES } from "@/lib/metadata/page-title";

import { ChangePasswordForm } from "./change-password-form";
import {
  DashboardAccountProfileLayout,
  DashboardProfileSection,
} from "./dashboard-account-profile-layout";
import { OperationalAccountProfileForm } from "./operational-account-profile-form";

type OperationalProfileRole =
  | typeof USER_ROLE.staff
  | typeof USER_ROLE.baker
  | typeof USER_ROLE.manager;

const OPERATIONAL_PROFILE_COPY: Record<
  OperationalProfileRole,
  {
    description: string;
    passwordSectionId: string;
    profileSectionId: string;
    roleLabel: string;
  }
> = {
  [USER_ROLE.staff]: {
    description:
      "Update your contact details and password. Operational fields are managed by an admin.",
    passwordSectionId: "staff-password",
    profileSectionId: "staff-profile-details",
    roleLabel: "Staff",
  },
  [USER_ROLE.baker]: {
    description:
      "Update your contact details and password. Production assignments are managed by an admin.",
    passwordSectionId: "baker-password",
    profileSectionId: "baker-profile-details",
    roleLabel: "Baker",
  },
  [USER_ROLE.manager]: {
    description:
      "Update your contact details and password. Team assignments are managed by an admin.",
    passwordSectionId: "manager-password",
    profileSectionId: "manager-profile-details",
    roleLabel: "Manager",
  },
};

type OperationalAccountProfileViewProps = {
  role: OperationalProfileRole;
};

/** Staff / baker / manager self-service profile + password on one page. */
export function OperationalAccountProfileView({
  role,
}: OperationalAccountProfileViewProps) {
  const copy = OPERATIONAL_PROFILE_COPY[role];

  return (
    <DashboardAccountProfileLayout
      description={copy.description}
      title={PAGE_TITLES.profile}
    >
      <DashboardProfileSection id={copy.profileSectionId} title="Account details">
        <OperationalAccountProfileForm
          expectedRole={role}
          roleLabel={copy.roleLabel}
        />
      </DashboardProfileSection>

      <DashboardProfileSection
        description="Changing your password signs out all active sessions."
        id={copy.passwordSectionId}
        title="Password"
      >
        <ChangePasswordForm />
      </DashboardProfileSection>
    </DashboardAccountProfileLayout>
  );
}
