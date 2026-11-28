"use client";

import { useEffect } from "react";

import { StorefrontErrorState } from "@/components/layouts/storefront-state";
import { MessageComposer } from "@/components/ui/message-composer";
import { MessageThread } from "@/components/ui/message-thread";
import { BRAND } from "@/constants/brand";
import { usePageVisible } from "@/lib/hooks/use-page-visible";

import { useMarkOrderMessagesRead, useOrderMessages, usePostOrderMessage } from "../hooks";
import { bakeryAuthor } from "../lib/messages";

type OrderMessagesProps = {
  orderId: string;
};

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
    <section aria-labelledby="order-messages" className="storefront-order-section storefront-order-messages">
      <div className="storefront-order-section__header">
        <h2 className="storefront-section-heading" id="order-messages">
          Messages
        </h2>
        <p className="text-caption">Ask our staff anything about this order; their reply appears here.</p>
      </div>
      {thread.isPending ? (
        <div aria-busy="true" className="grid gap-3">
          <span className="skeleton block h-12 w-2/3" />
          <span className="skeleton block h-12 w-1/2 justify-self-end" />
        </div>
      ) : thread.isError ? (
        <StorefrontErrorState onRetry={() => void thread.refetch()} title="We could not load the messages" />
      ) : (
        <MessageThread
          authorOf={bakeryAuthor}
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
        label="Write a message to our staff"
        pending={post.isPending}
        placeholder="Write a message…"
        onSend={(input, sent) => post.mutate(input, { onSuccess: sent })}
      />
    </section>
  );
}
