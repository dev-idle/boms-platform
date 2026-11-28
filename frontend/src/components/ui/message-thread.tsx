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
  /** True while the last message is still being written: it is read out once whole. */
  busy?: boolean;
  hasMore: boolean;
  loadingMore: boolean;
  onLoadMore: () => void;
};

/** Messages one side sends this close together read as one block. */
const RUN_GAP_MS = 5 * 60 * 1000;

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
 * A conversation read top to bottom, a rule between days. Messages one side
 * sends within a few minutes stand close together, the time under the last
 * of them only. It keeps the latest
 * message in view as messages arrive, or as the last one is written; loading
 * older ones leaves the reader where they were.
 */
export function MessageThread({
  messages,
  ownSide,
  authorOf,
  otherName,
  label,
  emptyText,
  busy = false,
  hasMore,
  loadingMore,
  onLoadMore,
}: MessageThreadProps) {
  const logRef = useRef<HTMLDivElement>(null);
  const latestId = messages.at(-1)?.id;
  const latestBody = messages.at(-1)?.body;

  useEffect(() => {
    const log = logRef.current;
    if (log && latestId) {
      log.scrollTop = log.scrollHeight;
    }
  }, [latestId, latestBody]);

  const now = new Date();
  const days = messages.map((message) => dayLabel(message.created_at, now));
  // Whether a message continues the one before it: same side, same author,
  // same day, sent within the gap.
  const continues = messages.map((message, index) => {
    const previous = messages[index - 1];
    return (
      previous !== undefined &&
      previous.from === message.from &&
      authorOf(previous) === authorOf(message) &&
      days[index - 1] === days[index] &&
      Date.parse(message.created_at) - Date.parse(previous.created_at) < RUN_GAP_MS
    );
  });
  return (
    <div aria-busy={busy || undefined} aria-label={label} className="message-thread" ref={logRef} role="log" tabIndex={0}>
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
          <div className={continues[index] ? "message-thread__entry message-thread__entry--continued" : "message-thread__entry"} key={message.id}>
            {days[index] !== days[index - 1] ? <p className="message-thread__day">{days[index]}</p> : null}
            <div className={message.from === ownSide ? "message message--own" : "message"}>
              <p className="message__bubble">
                <span className="sr-only">{sender}: </span>
                {message.body}
              </p>
              {continues[index + 1] ? null : (
                <p className="message__meta">
                  <time dateTime={message.created_at}>{splitDateTime(message.created_at).time}</time>
                  {author ? ` · ${author}` : null}
                </p>
              )}
            </div>
          </div>
        );
      })}
    </div>
  );
}
