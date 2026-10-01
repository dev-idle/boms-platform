"use client";

import { useEffect, useRef } from "react";

import { Button } from "@/components/ui/button";
import type { Message } from "@/lib/schemas/message";
import { splitDateTime } from "@/lib/validation/datetime";

type MessageThreadProps = {
  /** Oldest first. */
  messages: readonly Message[];
  /** The side the reader writes for: their messages sit on the right. */
  ownSide: Message["from"];
  /** Who wrote a message, shown after its time; null for none. */
  authorOf: (message: Message) => string | null;
  /** The other side, as assistive tech names their messages that carry no author. */
  otherName: string;
  label: string;
  emptyText: string;
  hasMore: boolean;
  loadingMore: boolean;
  onLoadMore: () => void;
};

/** "Today", "Yesterday" or the date: what the day rule between messages reads. */
function dayLabel(iso: string, now: Date): string {
  const day = new Date(iso);
  const yesterday = new Date(now);
  yesterday.setDate(now.getDate() - 1);
  if (day.toDateString() === now.toDateString()) {
    return "Today";
  }
  if (day.toDateString() === yesterday.toDateString()) {
    return "Yesterday";
  }
  return splitDateTime(iso).date;
}

/**
 * A conversation read top to bottom, a rule between days. It keeps the latest
 * message in view as messages arrive; loading older ones leaves the reader
 * where they were.
 */
export function MessageThread({
  messages,
  ownSide,
  authorOf,
  otherName,
  label,
  emptyText,
  hasMore,
  loadingMore,
  onLoadMore,
}: MessageThreadProps) {
  const logRef = useRef<HTMLDivElement>(null);
  const latestId = messages.at(-1)?.id;

  useEffect(() => {
    const log = logRef.current;
    if (log && latestId) {
      log.scrollTop = log.scrollHeight;
    }
  }, [latestId]);

  const now = new Date();
  const days = messages.map((message) => dayLabel(message.created_at, now));
  return (
    <div aria-label={label} className="message-thread" ref={logRef} role="log" tabIndex={0}>
      {hasMore ? (
        <Button
          aria-busy={loadingMore || undefined}
          className="message-thread__more"
          disabled={loadingMore}
          type="button"
          variant="outline"
          onClick={onLoadMore}
        >
          {loadingMore ? "Loading…" : "Earlier messages"}
        </Button>
      ) : null}
      {messages.length === 0 ? <p className="message-thread__empty">{emptyText}</p> : null}
      {messages.map((message, index) => {
        const author = authorOf(message);
        const sender = author ?? (message.from === ownSide ? "You" : otherName);
        return (
          <div className="message-thread__entry" key={message.id}>
            {days[index] !== days[index - 1] ? <p className="message-thread__day">{days[index]}</p> : null}
            <div className={message.from === ownSide ? "message message--own" : "message"}>
              <p className="message__bubble">
                <span className="sr-only">{sender}: </span>
                {message.body}
              </p>
              <p className="message__meta">
                <time dateTime={message.created_at}>{splitDateTime(message.created_at).time}</time>
                {author ? ` · ${author}` : null}
              </p>
            </div>
          </div>
        );
      })}
    </div>
  );
}
