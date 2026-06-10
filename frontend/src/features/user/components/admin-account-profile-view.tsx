import { AdminAccountProfileForm } from "./admin-account-profile-form";
import { ChangePasswordForm } from "./change-password-form";

/** Admin self-service profile + password (used by `/admin/account/profile`). */
export function AdminAccountProfileView() {
  return (
    <div className="max-w-2xl space-y-10">
      <header>
        <h1 className="text-2xl font-medium text-ink">
          Profile
        </h1>
        <p className="mt-2 text-sm text-ink-2">
          Your admin account details and password.
        </p>
      </header>

      <section className="space-y-4" aria-labelledby="admin-profile-details">
        <h2
          id="admin-profile-details"
          className="text-lg font-medium text-ink"
        >
          Account details
        </h2>
        <AdminAccountProfileForm />
      </section>

      <section
        className="space-y-4 border-t border-border pt-10"
        aria-labelledby="admin-profile-password"
      >
        <div>
          <h2
            id="admin-profile-password"
            className="text-lg font-medium text-ink"
          >
            Password
          </h2>
          <p className="mt-1 text-sm text-ink-2">
            Changing your password signs out all active sessions.
          </p>
        </div>
        <ChangePasswordForm />
      </section>
    </div>
  );
}
