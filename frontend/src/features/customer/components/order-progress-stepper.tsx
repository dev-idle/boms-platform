"use client";

import type { OrderStatus, OrderTimelineEntry } from "@/lib/schemas/order";
import { cn } from "@/lib/utils";
import { formatDateTime } from "@/lib/validation/datetime";

import {
  activeOrderProgressIndex,
  isOrderDropped,
  ORDER_PROGRESS_STEPS,
  statusReachedAt,
} from "../lib/order-progress";

type OrderProgressStepperProps = {
  status: OrderStatus;
  timeline: ReadonlyArray<OrderTimelineEntry>;
};

export function OrderProgressStepper({ status, timeline }: OrderProgressStepperProps) {
  const activeIndex = activeOrderProgressIndex(status);

  if (isOrderDropped(status)) {
    const droppedAt = statusReachedAt(timeline, status);
    return (
      <p className="storefront-order-progress storefront-order-progress--cancelled text-caption">
        {status === "expired" ? "This order was not paid in time and expired" : "This order was cancelled"}
        {droppedAt ? (
          <>
            {" on "}
            <time dateTime={droppedAt}>{formatDateTime(droppedAt)}</time>
          </>
        ) : null}
        .
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
