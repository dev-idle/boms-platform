"use client";

import { useEffect } from "react";

import { Button } from "@/components/ui/button";
import { DashboardTableActionLink } from "@/components/ui/dashboard-table-action-link";
import { InlineLoadingState } from "@/components/ui/loading-state";
import { MessageComposer } from "@/components/ui/message-composer";
import { MessageThread } from "@/components/ui/message-thread";
import { ROUTE } from "@/constants/routes";
import { usePageVisible } from "@/lib/hooks/use-page-visible";
import type { Message } from "@/lib/schemas/message";
import { initialsOf } from "@/lib/user-initials";

import {
  useMarkStaffMessagesRead,
  usePostStaffMessage,
  useSetConversationStatus,
  useStaffOrderMessages,
} from "../hooks";
import type { StaffOrder } from "../schemas";

type StaffChatThreadProps = {
  /** An order placed from an account: a guest has none to write to. */
  order: StaffOrder;
  customerEmail: string;
};

/** Who at the counter wrote each staff message; the customer's need no name. */
function authorOf(message: Message): string | null {
  return message.from === "staff" ? message.author_name : null;
}

/** One order's conversation at the counter: read on arrival, answered, resolved. */
export function StaffChatThread({ order, customerEmail }: StaffChatThreadProps) {
  const thread = useStaffOrderMessages(order.id);
  const post = usePostStaffMessage(order.id);
  const setStatus = useSetConversationStatus(order.id);
  const { mutate: markRead } = useMarkStaffMessagesRead(order.id);
  const pages = thread.data?.pages ?? [];
  const conversation = pages[0]?.conversation ?? null;
  const messages = [...pages].reverse().flatMap((page) => page.messages);
  const unread = conversation?.unread ?? 0;
  const latestId = messages.at(-1)?.id;
  const name = order.customer.display_name || customerEmail;
  const resolved = conversation?.status === "closed";
  const visible = usePageVisible();

  // What arrives while the conversation is on screen is read.
  useEffect(() => {
    if (unread > 0 && visible) {
      markRead();
    }
  }, [unread, latestId, visible, markRead]);

  return (
    <section aria-label={`Conversation with ${name}`} className="staff-chat__thread">
      <header className="staff-chat__thread-header">
        <DashboardTableActionLink
          className="staff-chat__back"
          href={ROUTE.staff.chat}
          label="Back to all messages"
          text="← All messages"
        />
        <div className="staff-chat__who">
          <span aria-hidden="true" className="staff-chat__avatar staff-chat__avatar--subject">
            {initialsOf(order.customer.display_name, customerEmail)}
          </span>
          <span className="staff-chat__who-text">
            <h2 className="staff-chat__thread-name">{name}</h2>
            <span className="staff-chat__handled">
              {conversation?.assigned_staff_name
                ? `Answered by ${conversation.assigned_staff_name}`
                : "Not answered yet"}
            </span>
          </span>
        </div>
        <div className="staff-chat__thread-actions">
          <DashboardTableActionLink
            className="staff-chat__order-link"
            href={ROUTE.staff.orderDetail(order.id)}
            label={`Open order ${order.code}`}
            showArrow
            text="Open order"
          />
          <Button
            aria-busy={setStatus.isPending || undefined}
            disabled={!conversation || setStatus.isPending}
            size="sm"
            title={conversation ? undefined : "Nothing to resolve before the first message"}
            type="button"
            variant="outline"
            onClick={() => setStatus.mutate(resolved ? "open" : "closed")}
          >
            {resolved ? "Open again" : "Mark resolved"}
          </Button>
        </div>
      </header>

      {thread.isPending ? (
        <div className="staff-chat__state">
          <InlineLoadingState />
        </div>
      ) : thread.isError ? (
        <div className="staff-chat__state">
          <p className="text-sm text-error">The messages could not be loaded.</p>
          <Button type="button" variant="outline" onClick={() => void thread.refetch()}>
            Try again
          </Button>
        </div>
      ) : (
        <MessageThread
          authorOf={authorOf}
          emptyText={`No messages yet. Write to ${name} about ${order.code}.`}
          hasMore={thread.hasNextPage}
          label={`Messages about ${order.code}`}
          loadingMore={thread.isFetchingNextPage}
          messages={messages}
          otherName={name}
          ownSide="staff"
          onLoadMore={() => void thread.fetchNextPage()}
        />
      )}

      <div className="staff-chat__composer">
        <MessageComposer
          label={`Write a reply to ${name}`}
          pending={post.isPending}
          placeholder="Write a reply…"
          onSend={(input, sent) => post.mutate(input, { onSuccess: sent })}
        />
      </div>
    </section>
  );
}
