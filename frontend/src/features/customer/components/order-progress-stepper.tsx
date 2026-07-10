"use client";

import { cn } from "@/lib/utils";

import {
  activeOrderProgressIndex,
  isOrderCancelled,
  ORDER_PROGRESS_STEPS,
} from "../lib/order-progress";
import type { OrderStatus } from "../schemas";

type OrderProgressStepperProps = {
  status: OrderStatus;
};

export function OrderProgressStepper({ status }: OrderProgressStepperProps) {
  const activeIndex = activeOrderProgressIndex(status);
  const cancelled = isOrderCancelled(status);

  if (cancelled) {
    return (
      <p className="storefront-order-progress storefront-order-progress--cancelled text-caption">
        This order was cancelled.
      </p>
    );
  }

  return (
    <ol aria-label="Order progress" className="storefront-order-progress">
      {ORDER_PROGRESS_STEPS.map((step, index) => {
        const isComplete = index < activeIndex;
        const isCurrent = index === activeIndex;
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
          </li>
        );
      })}
    </ol>
  );
}
