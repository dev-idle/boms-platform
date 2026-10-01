import { StaffChatConversation } from "@/features/staff";
import { PAGE_TITLES, pageTitle } from "@/lib/metadata/page-title";

export const metadata = pageTitle(PAGE_TITLES.messages);

type StaffChatThreadPageProps = {
  params: Promise<{ orderId: string }>;
};

export default async function StaffChatThreadPage({ params }: StaffChatThreadPageProps) {
  const { orderId } = await params;
  return <StaffChatConversation orderId={orderId} />;
}
