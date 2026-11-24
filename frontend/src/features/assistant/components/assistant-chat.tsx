"use client";

import Link from "next/link";

import { Button } from "@/components/ui/button";
import { MessageComposer } from "@/components/ui/message-composer";
import { MessageThread } from "@/components/ui/message-thread";
import { ROUTE } from "@/constants/routes";
import type { Message } from "@/lib/schemas/message";

import type { AssistantChatState } from "../hooks";
import { ASSISTANT_QUESTION_MAX_LENGTH } from "../schemas";

const ASSISTANT_NAME = "Assistant";

type AssistantChatProps = {
  chat: AssistantChatState;
  /** Opens the conversation with staff. */
  onTalkToStaff: () => void;
};

function authorOf(message: Message): string | null {
  return message.from === "staff" ? ASSISTANT_NAME : null;
}

/** Questions to the assistant and its answers. */
export function AssistantChat({ chat, onTalkToStaff }: AssistantChatProps) {
  const handOff = chat.lastAnswer?.hand_off ?? null;

  return (
    <>
      <p className="assistant-panel__notice">Answers are written automatically and can be wrong.</p>
      <MessageThread
        authorOf={authorOf}
        busy={chat.pending}
        emptyText="Ask about our products, prices, opening hours or one of your orders."
        hasMore={false}
        label="Conversation with the assistant"
        loadingMore={false}
        messages={chat.turns}
        otherName={ASSISTANT_NAME}
        ownSide="customer"
        onLoadMore={() => undefined}
      />
      <div aria-live="polite" className="assistant-panel__status">
        {chat.failure ? <p className="text-sm text-error">{chat.failure}</p> : null}
        {handOff?.order_id && handOff.order_code ? (
          <Link className="storefront-inline-link" href={ROUTE.orderDetail(handOff.order_id)}>
            Open order {handOff.order_code}
          </Link>
        ) : null}
      </div>
      <div className="assistant-panel__staff">
        <Button
          size="sm"
          type="button"
          variant={chat.lastAnswer?.offer_staff || chat.failure ? "default" : "ghost"}
          onClick={onTalkToStaff}
        >
          Talk to staff
        </Button>
      </div>
      <MessageComposer
        label="Ask the assistant a question"
        maxLength={ASSISTANT_QUESTION_MAX_LENGTH}
        pending={chat.pending}
        placeholder="Ask a question…"
        onSend={(input, sent) => void chat.ask(input, sent)}
      />
    </>
  );
}
