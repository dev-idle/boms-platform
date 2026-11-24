"use client";

import { useEffect } from "react";

import { Button } from "@/components/ui/button";
import { InlineLoadingState } from "@/components/ui/loading-state";
import { MessageComposer } from "@/components/ui/message-composer";
import { MessageThread } from "@/components/ui/message-thread";
import { BRAND } from "@/constants/brand";
import { usePageVisible } from "@/lib/hooks/use-page-visible";

import { useGeneralMessages, useMarkGeneralMessagesRead, usePostGeneralMessage } from "../hooks";
import { bakeryAuthor } from "../lib/messages";

/** The customer and the counter writing to each other about anything but one of their orders. */
export function GeneralMessages() {
  const thread = useGeneralMessages();
  const post = usePostGeneralMessage();
  const { mutate: markRead } = useMarkGeneralMessagesRead();
  const pages = thread.data?.pages ?? [];
  const messages = [...pages].reverse().flatMap((page) => page.messages);
  const unread = pages[0]?.unread ?? 0;
  const latestId = messages.at(-1)?.id;
  const visible = usePageVisible();

  // What arrives while the conversation is on screen is read.
  useEffect(() => {
    if (unread > 0 && visible) {
      markRead();
    }
  }, [unread, latestId, visible, markRead]);

  return (
    <div className="general-messages">
      {thread.isPending ? (
        <div className="general-messages__state">
          <InlineLoadingState />
        </div>
      ) : thread.isError ? (
        <div className="general-messages__state">
          <p className="text-sm text-error">We could not load the messages.</p>
          <Button type="button" variant="outline" onClick={() => void thread.refetch()}>
            Try again
          </Button>
        </div>
      ) : (
        <MessageThread
          authorOf={bakeryAuthor}
          emptyText="Ask us anything; our staff reply here."
          hasMore={thread.hasNextPage}
          label="Messages with our staff"
          loadingMore={thread.isFetchingNextPage}
          messages={messages}
          otherName={BRAND.name}
          ownSide="customer"
          onLoadMore={() => void thread.fetchNextPage()}
        />
      )}
      <MessageComposer
        label="Write a message to our staff"
        pending={post.isPending}
        placeholder="Write a message…"
        onSend={(input, sent) => post.mutate(input, { onSuccess: sent })}
      />
    </div>
  );
}
