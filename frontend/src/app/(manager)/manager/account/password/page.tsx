import { ChangePasswordForm } from "@/features/user";

export default function ManagerAccountPasswordPage() {
  return (
    <div className="max-w-2xl space-y-6">
      <div>
        <h1 className="text-page-title">
          Change password
        </h1>
        <p className="mt-2 text-sm text-ink-2">
          You will be signed out after updating your password.
        </p>
      </div>
      <ChangePasswordForm />
    </div>
  );
}
