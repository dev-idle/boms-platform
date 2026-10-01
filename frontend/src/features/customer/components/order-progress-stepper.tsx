"use client";

import type { OrderStatus, OrderTimelineEntry } from "@/lib/schemas/order";
import { cn } from "@/lib/utils";
import { formatDateTime } from "@/lib/validation/datetime";

import {
  activeOrderProgressIndex,
  isUncollected,
  ORDER_PROGRESS_STEPS,
  statusReachedAt,
} from "../lib/order-progress";

type OrderProgressStepperProps = {
  status: OrderStatus;
  timeline: ReadonlyArray<OrderTimelineEntry>;
};

/** How an order that was not collected ended. */
function endedNote(status: OrderStatus): string {
  switch (status) {
    case "expired":
      return "This order was not paid in time and expired";
    case "no_show":
      return "This order was not collected by closing time";
    default:
      return "This order was cancelled";
  }
}

export function OrderProgressStepper({ status, timeline }: OrderProgressStepperProps) {
  const activeIndex = activeOrderProgressIndex(status);

  if (isUncollected(status)) {
    const endedAt = statusReachedAt(timeline, status);
    const reason = timeline.find((entry) => entry.status === "cancelled")?.reason;
    return (
      <p className="storefront-order-progress storefront-order-progress--cancelled text-caption">
        {endedNote(status)}
        {endedAt ? (
          <>
            {" on "}
            <time dateTime={endedAt}>{formatDateTime(endedAt)}</time>
          </>
        ) : null}
        .{reason ? ` Our reason: ${reason}` : null}
      </p>
    );
  }

  return (
    <ol aria-label="Order progress" className="storefront-order-progress">
      {ORDER_PROGRESS_STEPS.map((step, index) => {
        const isComplete = index < activeIndex;
        const isCurrent = index === activeIndex;
        const reachedAt = statusReachedAt(timeline, step.key);
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
            {reachedAt ? (
              <time className="storefront-order-progress__time" dateTime={reachedAt}>
                {formatDateTime(reachedAt)}
              </time>
            ) : null}
          </li>
        );
      })}
    </ol>
  );
}
