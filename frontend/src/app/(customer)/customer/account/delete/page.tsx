import { DeleteAccountCard } from "@/features/user";

export default function CustomerAccountDeletePage() {
  return (
    <div className="max-w-2xl space-y-6">
      <div>
        <h1 className="text-2xl font-medium text-ink">
          Delete account
        </h1>
        <p className="mt-2 text-sm text-ink-2">
          This is available for customers only.
        </p>
      </div>
      <DeleteAccountCard />
    </div>
  );
}
