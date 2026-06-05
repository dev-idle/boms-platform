import type { StaffProfile } from "../types";

type ReadonlyStaffProfileFieldsProps = {
  profile: StaffProfile;
};

export function ReadonlyStaffProfileFields({
  profile,
}: ReadonlyStaffProfileFieldsProps) {
  return (
    <div className="space-y-3 rounded-md border border-zinc-200 bg-zinc-50 p-4 dark:border-zinc-800 dark:bg-zinc-950">
      <p className="text-xs font-medium uppercase tracking-wide text-zinc-500">
        Set by admin
      </p>
      <div>
        <p className="text-xs text-zinc-500">Employee code</p>
        <p className="text-sm font-medium text-zinc-900 dark:text-zinc-100">
          {profile.employee_code}
        </p>
      </div>
    </div>
  );
}
