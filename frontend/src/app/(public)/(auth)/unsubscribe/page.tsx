import { UnsubscribeView } from "@/features/auth";
import { PAGE_TITLES, pageTitle } from "@/lib/metadata/page-title";

export const metadata = pageTitle(PAGE_TITLES.unsubscribe);

export default function UnsubscribePage() {
  return <UnsubscribeView />;
}
