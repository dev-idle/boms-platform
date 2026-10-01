"use client";

import { useEffect } from "react";

import { Button } from "@/components/ui/button";
import { InlineLoadingState } from "@/components/ui/loading-state";
import { MessageComposer } from "@/components/ui/message-composer";
import { MessageThread } from "@/components/ui/message-thread";
import { BRAND } from "@/constants/brand";
import { usePageVisible } from "@/lib/hooks/use-page-visible";
import type { Message } from "@/lib/schemas/message";

import { useMarkOrderMessagesRead, useOrderMessages, usePostOrderMessage } from "../hooks";

type OrderMessagesProps = {
  orderId: string;
};

/** The counter's messages read as the bakery's: who at the counter wrote one is theirs to keep. */
function authorOf(message: Message): string | null {
  return message.from === "staff" ? BRAND.name : null;
}

/** The customer and the counter writing to each other about an order the bakery took. */
export function OrderMessages({ orderId }: OrderMessagesProps) {
  const thread = useOrderMessages(orderId);
  const post = usePostOrderMessage(orderId);
  const { mutate: markRead } = useMarkOrderMessagesRead(orderId);
  const pages = thread.data?.pages ?? [];
  const messages = [...pages].reverse().flatMap((page) => page.messages);
  const unread = pages[0]?.unread ?? 0;
  const latestId = messages.at(-1)?.id;
  const visible = usePageVisible();

  // What arrives while the thread is on screen is read.
  useEffect(() => {
    if (unread > 0 && visible) {
      markRead();
    }
  }, [unread, latestId, visible, markRead]);

  return (
    <section aria-labelledby="order-messages" className="storefront-panel storefront-order-messages">
      <div className="storefront-order-messages__header">
        <h2 className="text-section-heading" id="order-messages">
          Messages
        </h2>
        <p className="text-caption">Ask us anything about this order; our reply appears here.</p>
      </div>
      {thread.isPending ? (
        <InlineLoadingState />
      ) : thread.isError ? (
        <div className="storefront-order-messages__error">
          <p className="text-sm text-error">We could not load the messages.</p>
          <Button type="button" variant="outline" onClick={() => void thread.refetch()}>
            Try again
          </Button>
        </div>
      ) : (
        <MessageThread
          authorOf={authorOf}
          emptyText="No messages yet."
          hasMore={thread.hasNextPage}
          label="Messages about this order"
          loadingMore={thread.isFetchingNextPage}
          messages={messages}
          otherName={BRAND.name}
          ownSide="customer"
          onLoadMore={() => void thread.fetchNextPage()}
        />
      )}
      <MessageComposer
        label="Write a message to the bakery"
        pending={post.isPending}
        placeholder="Write a message…"
        onSend={(input, sent) => post.mutate(input, { onSuccess: sent })}
      />
    </section>
  );
}
