import { AdminAccountProfileForm } from "@/features/user";

export default function AdminAccountProfilePage() {
  return (
    <div className="max-w-2xl space-y-6">
      <div>
        <h1 className="text-2xl font-semibold text-zinc-900 dark:text-zinc-50">
          Admin profile
        </h1>
        <p className="mt-2 text-sm text-zinc-600 dark:text-zinc-300">
          Manage your administrative profile details.
        </p>
      </div>
      <AdminAccountProfileForm />
    </div>
  );
}
