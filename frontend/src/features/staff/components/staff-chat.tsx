"use client";

import { useParams } from "next/navigation";
import type { ReactNode } from "react";

import { StaffChatList } from "./staff-chat-list";

type StaffChatProps = {
  /** The open conversation, or the prompt to choose one. */
  children: ReactNode;
};

/**
 * The counter's inbox beside the conversation it opens. It frames every chat
 * page, so the inbox keeps its filter, page and scroll from one conversation
 * to the next.
 */
export function StaffChat({ children }: StaffChatProps) {
  const { orderId } = useParams<{ orderId?: string }>();
  return (
    <div className={orderId ? "staff-chat staff-chat--open" : "staff-chat"}>
      <StaffChatList activeOrderId={orderId} />
      {children}
    </div>
  );
}
