import { AdminAccountProfileForm } from "./admin-account-profile-form";
import { ChangePasswordForm } from "./change-password-form";

/** Admin self-service profile + password (used by `/admin/account/profile`). */
export function AdminAccountProfileView() {
  return (
    <div className="max-w-2xl space-y-10">
      <header>
        <h1 className="text-2xl font-semibold text-zinc-900 dark:text-zinc-50">
          Profile
        </h1>
        <p className="mt-2 text-sm text-zinc-600 dark:text-zinc-300">
          Your admin account details and password.
        </p>
      </header>

      <section className="space-y-4" aria-labelledby="admin-profile-details">
        <h2
          id="admin-profile-details"
          className="text-lg font-medium text-zinc-900 dark:text-zinc-50"
        >
          Account details
        </h2>
        <AdminAccountProfileForm />
      </section>

      <section
        className="space-y-4 border-t border-zinc-200 pt-10 dark:border-zinc-800"
        aria-labelledby="admin-profile-password"
      >
        <div>
          <h2
            id="admin-profile-password"
            className="text-lg font-medium text-zinc-900 dark:text-zinc-50"
          >
            Password
          </h2>
          <p className="mt-1 text-sm text-zinc-600 dark:text-zinc-300">
            Changing your password signs out all active sessions.
          </p>
        </div>
        <ChangePasswordForm />
      </section>
    </div>
  );
}
