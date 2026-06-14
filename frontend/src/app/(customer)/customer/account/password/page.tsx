import { ChangePasswordForm } from "@/features/user";
import { PAGE_TITLES, pageTitle } from "@/lib/metadata/page-title";

export const metadata = pageTitle(PAGE_TITLES.changePassword);

export default function CustomerAccountPasswordPage() {
  return (
    <div className="max-w-2xl space-y-6">
      <div>
        <h1 className="text-page-title">{PAGE_TITLES.changePassword}</h1>
        <p className="mt-2 text-sm text-ink-2">
          After changing password, you will be signed out from all sessions.
        </p>
      </div>
      <ChangePasswordForm />
    </div>
  );
}
