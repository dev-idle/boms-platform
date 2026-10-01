import { StaffChatConversation } from "@/features/staff";
import { PAGE_TITLES, pageTitle } from "@/lib/metadata/page-title";

export const metadata = pageTitle(PAGE_TITLES.messages);

export default function StaffChatPage() {
  return <StaffChatConversation />;
}
