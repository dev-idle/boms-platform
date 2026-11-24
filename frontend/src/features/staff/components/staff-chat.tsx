"use client";

import { useParams } from "next/navigation";
import type { ReactNode } from "react";

import { cn } from "@/lib/utils";

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
  const { orderId, customerId } = useParams<{ orderId?: string; customerId?: string }>();
  const activeId = orderId ?? customerId;
  return (
    <div className={cn("staff-chat", activeId && "staff-chat--open", customerId && "staff-chat--general")}>
      <StaffChatList activeId={activeId} />
      {children}
    </div>
  );
}
