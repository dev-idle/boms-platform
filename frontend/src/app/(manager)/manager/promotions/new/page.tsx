import { ManagerNewPromotion } from "@/features/manager";
import { PAGE_TITLES, pageTitle } from "@/lib/metadata/page-title";

export const metadata = pageTitle(PAGE_TITLES.newPromotion);

export default function ManagerNewPromotionPage() {
  return <ManagerNewPromotion />;
}
