"use client";

import type { OrderStatus, OrderTimelineEntry } from "@/lib/schemas/order";
import { cn } from "@/lib/utils";
import { formatDateTime, splitDateTime } from "@/lib/validation/datetime";

import {
  activeOrderProgressIndex,
  isUncollected,
  ORDER_PROGRESS_STEPS,
  statusReachedAt,
} from "../lib/order-progress";

type OrderProgressStepperProps = {
  status: OrderStatus;
  timeline: ReadonlyArray<OrderTimelineEntry>;
  /** When the order was placed: an order staff took starts confirmed, so its timeline has no "Placed" entry. */
  placedAt: string;
};

/** How an order that was not picked up ended. */
function endedNote(status: OrderStatus): string {
  switch (status) {
    case "expired":
      return "This order was not paid in time and expired";
    case "no_show":
      return "This order was not picked up by closing time";
    default:
      return "This order was cancelled";
  }
}

/** What the customer can still do about an order that ended that way; nothing about one cancelled. */
function endedNextStep(status: OrderStatus): string | null {
  switch (status) {
    case "expired":
      return " Nothing was charged; you can order again.";
    case "no_show":
      return " Message our staff if you still want to pick it up.";
    default:
      return null;
  }
}

export function OrderProgressStepper({ status, timeline, placedAt }: OrderProgressStepperProps) {
  const activeIndex = activeOrderProgressIndex(status);

  if (isUncollected(status)) {
    const endedAt = statusReachedAt(timeline, status);
    const reason = timeline.find((entry) => entry.status === "cancelled")?.reason;
    return (
      <p
        className={cn(
          "storefront-order-progress storefront-order-progress--ended text-caption",
          status === "cancelled" && "storefront-order-progress--cancelled",
        )}
      >
        {endedNote(status)}
        {endedAt ? (
          <>
            {" on "}
            <time dateTime={endedAt}>{formatDateTime(endedAt)}</time>
          </>
        ) : null}
        .{reason ? ` Our reason: ${reason}` : null}
        {endedNextStep(status)}
      </p>
    );
  }

  return (
    <ol aria-label="Order progress" className="storefront-order-progress">
      {ORDER_PROGRESS_STEPS.map((step, index) => {
        const isComplete = index < activeIndex;
        const isCurrent = index === activeIndex;
        const reachedAt = index === 0 ? placedAt : statusReachedAt(timeline, step.key);
        const reached = reachedAt ? splitDateTime(reachedAt) : null;
        return (
          <li
            key={step.key}
            className={cn(
              "storefront-order-progress__step",
              isComplete && "storefront-order-progress__step--complete",
              isCurrent && "storefront-order-progress__step--current",
            )}
          >
            <span aria-hidden="true" className="storefront-order-progress__marker" />
            <span className="storefront-order-progress__label">{step.label}</span>
            {reachedAt && reached ? (
              <time className="storefront-order-progress__time" dateTime={reachedAt}>
                <span>{reached.date}</span> <span>{reached.time}</span>
              </time>
            ) : null}
          </li>
        );
      })}
    </ol>
  );
}
