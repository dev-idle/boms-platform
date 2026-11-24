import type { ReactNode } from "react";

import { DashboardTableActionLink } from "@/components/ui/dashboard-table-action-link";
import { ROUTE } from "@/constants/routes";

type StaffChatStateProps = {
  /** Offers the way back to the inbox, from a conversation the link named. */
  back: boolean;
  children: ReactNode;
};

/** Where a conversation would be: a prompt to choose one, or why it cannot be shown. */
export function StaffChatState({ back, children }: StaffChatStateProps) {
  return (
    <section aria-label="Conversation" className="staff-chat__thread">
      <div className="staff-chat__state">
        {back ? (
          <DashboardTableActionLink
            className="staff-chat__back"
            href={ROUTE.staff.chat}
            label="Back to all messages"
            text="← All messages"
          />
        ) : null}
        {children}
      </div>
    </section>
  );
}
