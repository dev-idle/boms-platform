import { USER_ROLE } from "@/constants/roles";
import { OperationalAccountProfileForm } from "@/features/user";

export default function StaffAccountProfilePage() {
  return (
    <div className="max-w-2xl space-y-6">
      <div>
        <h1 className="text-2xl font-semibold text-zinc-900 dark:text-zinc-50">
          Staff profile
        </h1>
        <p className="mt-2 text-sm text-zinc-600 dark:text-zinc-300">
          Update your contact details. Operational fields are managed by an admin.
        </p>
      </div>
      <OperationalAccountProfileForm
        expectedRole={USER_ROLE.staff}
        roleLabel="Staff"
      />
    </div>
  );
}
