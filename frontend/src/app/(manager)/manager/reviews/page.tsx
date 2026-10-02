import { ManagerReviews } from "@/features/manager";
import { PAGE_TITLES, pageTitle } from "@/lib/metadata/page-title";

export const metadata = pageTitle(PAGE_TITLES.reviews);

export default function ManagerReviewsPage() {
  return <ManagerReviews />;
}
