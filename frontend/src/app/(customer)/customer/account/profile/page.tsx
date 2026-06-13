import { CustomerAccountProfileForm } from "@/features/user";
import { PAGE_TITLES, pageTitle } from "@/lib/metadata/page-title";

export const metadata = pageTitle(PAGE_TITLES.profile);

export default function CustomerAccountProfilePage() {
  return (
    <div className="max-w-2xl space-y-6">
      <div>
        <h1 className="text-page-title">{PAGE_TITLES.myProfile}</h1>
        <p className="mt-2 text-sm text-ink-2">
          Update your customer profile details.
        </p>
      </div>
      <CustomerAccountProfileForm />
    </div>
  );
}
