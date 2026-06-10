import type { StaffProfile } from "../types";

type ReadonlyStaffProfileFieldsProps = {
  profile: StaffProfile;
};

export function ReadonlyStaffProfileFields({
  profile,
}: ReadonlyStaffProfileFieldsProps) {
  return (
    <div className="space-y-3 rounded-md border border-border bg-surface-alt p-4 rounded-card">
      <p className="text-xs font-medium uppercase tracking-wide text-muted">
        Set by admin
      </p>
      <div>
        <p className="text-xs text-muted">Employee code</p>
        <p className="text-sm font-medium text-ink">
          {profile.employee_code}
        </p>
      </div>
    </div>
  );
}
