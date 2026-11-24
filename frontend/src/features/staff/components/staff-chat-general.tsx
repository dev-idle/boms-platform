"use client";

import { z } from "zod";

import { Button } from "@/components/ui/button";
import { InlineLoadingState } from "@/components/ui/loading-state";
import { isApiError } from "@/lib/errors";

import { useStaffMessages } from "../hooks";
import { StaffChatState } from "./staff-chat-state";
import { StaffChatThread } from "./staff-chat-thread";

type StaffChatGeneralProps = {
  customerId: string;
};

/** A customer's general conversation, about none of their orders, or why there is none to show. */
export function StaffChatGeneral({ customerId }: StaffChatGeneralProps) {
  if (!z.uuid().safeParse(customerId).success) {
    return (
      <StaffChatState back>
        <p className="text-sm text-muted">Invalid conversation link.</p>
      </StaffChatState>
    );
  }
  return <StaffChatGeneralThread customerId={customerId} />;
}

function StaffChatGeneralThread({ customerId }: StaffChatGeneralProps) {
  const thread = { kind: "customer", id: customerId } as const;
  const messages = useStaffMessages(thread);
  const customer = messages.data?.pages[0]?.customer;

  if (messages.isPending) {
    return (
      <StaffChatState back>
        <InlineLoadingState />
      </StaffChatState>
    );
  }
  if (messages.isError || !customer) {
    return (
      <StaffChatState back>
        {isApiError(messages.error) && messages.error.status === 404 ? (
          <p className="text-sm text-error">Conversation not found.</p>
        ) : (
          <>
            <p className="text-sm text-error">The conversation could not be loaded.</p>
            <Button size="sm" type="button" variant="outline" onClick={() => void messages.refetch()}>
              Try again
            </Button>
          </>
        )}
      </StaffChatState>
    );
  }
  return (
    <StaffChatThread customerEmail={customer.email} customerName={customer.display_name} order={null} thread={thread} />
  );
}
