import { RefundPolicy } from "@/features/legal";
import { PAGE_TITLES, pageTitle } from "@/lib/metadata/page-title";

export const instant = true;

export const metadata = pageTitle(PAGE_TITLES.refundPolicy, "Cancellations, refunds and missed pickups.");

export default function RefundPolicyPage() {
  return <RefundPolicy />;
}
