import { ChangePasswordForm } from "@/features/user";

export default function CustomerAccountPasswordPage() {
  return (
    <div className="max-w-2xl space-y-6">
      <div>
        <h1 className="text-2xl font-semibold text-zinc-900 dark:text-zinc-50">
          Change password
        </h1>
        <p className="mt-2 text-sm text-zinc-600 dark:text-zinc-300">
          After changing password, you will be signed out from all sessions.
        </p>
      </div>
      <ChangePasswordForm />
    </div>
  );
}
