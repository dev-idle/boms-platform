import { DeleteAccountCard } from "@/features/user";

export default function CustomerAccountDeletePage() {
  return (
    <div className="max-w-2xl space-y-6">
      <div>
        <h1 className="text-2xl font-semibold text-zinc-900 dark:text-zinc-50">
          Delete account
        </h1>
        <p className="mt-2 text-sm text-zinc-600 dark:text-zinc-300">
          This is available for customers only.
        </p>
      </div>
      <DeleteAccountCard />
    </div>
  );
}
