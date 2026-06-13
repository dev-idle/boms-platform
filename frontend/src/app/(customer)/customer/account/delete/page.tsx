import { DeleteAccountCard } from "@/features/user";
import { PAGE_TITLES, pageTitle } from "@/lib/metadata/page-title";

export const metadata = pageTitle(PAGE_TITLES.deleteAccount);

export default function CustomerAccountDeletePage() {
  return (
    <div className="max-w-2xl space-y-6">
      <div>
        <h1 className="text-page-title">{PAGE_TITLES.deleteAccount}</h1>
        <p className="mt-2 text-sm text-ink-2">
          This is available for customers only.
        </p>
      </div>
      <DeleteAccountCard />
    </div>
  );
}
