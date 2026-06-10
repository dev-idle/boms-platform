import { CustomerAccountProfileForm } from "@/features/user";

export default function CustomerAccountProfilePage() {
  return (
    <div className="max-w-2xl space-y-6">
      <div>
        <h1 className="text-2xl font-medium text-ink">
          My profile
        </h1>
        <p className="mt-2 text-sm text-ink-2">
          Update your customer profile details.
        </p>
      </div>
      <CustomerAccountProfileForm />
    </div>
  );
}
