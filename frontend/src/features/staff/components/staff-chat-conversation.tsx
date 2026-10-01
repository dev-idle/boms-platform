"use client";

import type { ReactNode } from "react";
import { z } from "zod";

import { DashboardTableActionLink } from "@/components/ui/dashboard-table-action-link";
import { InlineLoadingState } from "@/components/ui/loading-state";
import { ROUTE } from "@/constants/routes";
import { isApiError } from "@/lib/errors";
import { formatVietnamPhone } from "@/lib/validation/phone";

import { useStaffOrder } from "../hooks";
import { StaffChatOrder } from "./staff-chat-order";
import { StaffChatThread } from "./staff-chat-thread";

type StaffChatConversationProps = {
  /** The order whose conversation is open; none asks to choose one. */
  orderId?: string;
};

/** One order's conversation and the order beside it, or why there is none to show. */
export function StaffChatConversation({ orderId = "" }: StaffChatConversationProps) {
  const orderQuery = useStaffOrder(orderId);
  const order = orderQuery.data;

  let state: ReactNode;
  if (orderId === "") {
    state = <p className="text-sm text-muted">Choose a conversation to read it.</p>;
  } else if (!z.uuid().safeParse(orderId).success) {
    state = <p className="text-sm text-muted">Invalid conversation link.</p>;
  } else if (orderQuery.isPending) {
    state = <InlineLoadingState />;
  } else if (orderQuery.isError || !order) {
    state = (
      <p className="text-sm text-error">
        {isApiError(orderQuery.error) && orderQuery.error.status === 404
          ? "Order not found."
          : "The order could not be loaded."}
      </p>
    );
  } else if (order.customer.email === null) {
    const phone = formatVietnamPhone(order.customer.phone);
    state = (
      <p className="text-sm text-muted">
        {order.code} was taken for a guest, who has no account to write to
        {phone ? `. Call them on ${phone}.` : "."}
      </p>
    );
  } else {
    return (
      <>
        <StaffChatThread customerEmail={order.customer.email} order={order} />
        <StaffChatOrder customerEmail={order.customer.email} order={order} />
      </>
    );
  }

  return (
    <section aria-label="Conversation" className="staff-chat__thread">
      <div className="staff-chat__state">
        {orderId === "" ? null : (
          <DashboardTableActionLink
            className="staff-chat__back"
            href={ROUTE.staff.chat}
            label="Back to all messages"
            text="← All messages"
          />
        )}
        {state}
      </div>
    </section>
  );
}
