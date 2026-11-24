import { StaffChatGeneral } from "@/features/staff";
import { PAGE_TITLES, pageTitle } from "@/lib/metadata/page-title";

export const metadata = pageTitle(PAGE_TITLES.messages);

type StaffChatGeneralPageProps = {
  params: Promise<{ customerId: string }>;
};

export default async function StaffChatGeneralPage({ params }: StaffChatGeneralPageProps) {
  const { customerId } = await params;
  return <StaffChatGeneral customerId={customerId} />;
}
