import { PrivacyPolicy } from "@/features/legal";
import { PAGE_TITLES, pageTitle } from "@/lib/metadata/page-title";

export const instant = true;

export const metadata = pageTitle(PAGE_TITLES.privacy, "What personal data we keep, why, and your rights over it.");

export default function PrivacyPage() {
  return <PrivacyPolicy />;
}
