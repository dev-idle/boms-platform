import type { StaffProfile } from "../types";

type ReadonlyStaffProfileFieldsProps = {
  profile: StaffProfile;
};

export function ReadonlyStaffProfileFields({
  profile,
}: ReadonlyStaffProfileFieldsProps) {
  return (
    <div className="dashboard-readonly-fields">
      <p className="dashboard-readonly-fields-label">Set by admin</p>
      <div>
        <p className="dashboard-readonly-fields-key">Employee code</p>
        <p className="dashboard-readonly-fields-value">{profile.employee_code}</p>
      </div>
    </div>
  );
}
