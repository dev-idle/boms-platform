"use client";

import Link from "next/link";
import { useState } from "react";

import { AsyncPanel } from "@/components/ui/async-panel";
import { Button } from "@/components/ui/button";
import { DashboardTablePagination } from "@/components/ui/dashboard-table-pagination";
import { ROUTE } from "@/constants/routes";
import { getQuerySurface } from "@/lib/react-query/query-surface";
import { initialsOf } from "@/lib/user-initials";
import { splitDateTime } from "@/lib/validation/datetime";

import { useStaffConversationCounts, useStaffConversations } from "../hooks";
import type { ConversationStatus } from "../schemas";

const PAGE_SIZE = 20;
const PLACEHOLDER_ROWS = 6;

type StaffChatListProps = {
  activeOrderId: string | undefined;
};

/** Today's messages show their time; older ones their date. */
function whenLabel(iso: string): string {
  const { date, time } = splitDateTime(iso);
  return new Date(iso).toDateString() === new Date().toDateString() ? time : date;
}

/** The counter's inbox: what is unresolved first, or every conversation, latest message first. */
export function StaffChatList({ activeOrderId }: StaffChatListProps) {
  const [status, setStatus] = useState<ConversationStatus | undefined>("open");
  const [page, setPage] = useState(1);
  const listQuery = useStaffConversations({ page, page_size: PAGE_SIZE, status });
  const counts = useStaffConversationCounts();
  const { initialLoading, refetching } = getQuerySurface(listQuery);
  const conversations = listQuery.data?.conversations ?? [];
  const pagination = listQuery.data?.pagination;

  function show(next: ConversationStatus | undefined): void {
    setStatus(next);
    setPage(1);
  }

  return (
    <section aria-labelledby="staff-chat-title" className="staff-chat__list">
      <div className="staff-chat__list-header">
        <h1 className="text-page-title" id="staff-chat-title">
          Messages
        </h1>
        <div aria-label="Show conversations" className="staff-chat__tabs" role="group">
          <button
            aria-pressed={status === "open"}
            className="staff-chat__tab"
            type="button"
            onClick={() => show("open")}
          >
            Unresolved
            {counts.data ? <span className="staff-chat__tab-count">{counts.data.open}</span> : null}
          </button>
          <button
            aria-pressed={status === undefined}
            className="staff-chat__tab"
            type="button"
            onClick={() => show(undefined)}
          >
            All
          </button>
        </div>
      </div>

      <AsyncPanel className="staff-chat__conversations" overlayOnInitialLoad={false} refetching={refetching}>
        {initialLoading ? (
          <ul aria-label="Loading conversations" className="staff-chat__rows">
            {Array.from({ length: PLACEHOLDER_ROWS }, (_, index) => (
              <li className="staff-chat__row" key={index}>
                <span className="staff-chat__avatar skeleton" />
                <span className="staff-chat__row-body">
                  <span className="staff-chat__placeholder-line skeleton" />
                  <span className="staff-chat__placeholder-line staff-chat__placeholder-line--short skeleton" />
                </span>
              </li>
            ))}
          </ul>
        ) : listQuery.isError ? (
          <div className="staff-chat__state">
            <p className="text-sm text-error">The conversations could not be loaded.</p>
            <Button type="button" variant="outline" onClick={() => void listQuery.refetch()}>
              Try again
            </Button>
          </div>
        ) : conversations.length === 0 ? (
          <p className="staff-chat__state text-sm text-muted">
            {status === "open" ? "Nothing waits on the counter." : "No customer has written yet."}
          </p>
        ) : (
          <ul className="staff-chat__rows">
            {conversations.map((conversation) => {
              const name = conversation.customer_name ?? conversation.customer_email;
              const unread = conversation.unread > 0;
              return (
                <li key={conversation.order_id}>
                  <Link
                    aria-current={conversation.order_id === activeOrderId ? "true" : undefined}
                    className={unread ? "staff-chat__row staff-chat__row--unread" : "staff-chat__row"}
                    href={ROUTE.staff.chatThread(conversation.order_id)}
                  >
                    <span aria-hidden="true" className="staff-chat__avatar">
                      {initialsOf(conversation.customer_name, conversation.customer_email)}
                    </span>
                    <span className="staff-chat__row-body">
                      <span className="staff-chat__row-top">
                        <span className="staff-chat__name">{name}</span>
                        <time className="staff-chat__time" dateTime={conversation.last_message_at}>
                          {whenLabel(conversation.last_message_at)}
                        </time>
                      </span>
                      <span className="staff-chat__preview">{conversation.preview}</span>
                      <span className="staff-chat__row-meta">
                        <span className="text-order-code">{conversation.order_code}</span>
                        {unread ? (
                          <span className="staff-chat__unread">
                            {conversation.unread}
                            <span className="sr-only"> unread</span>
                          </span>
                        ) : null}
                      </span>
                    </span>
                  </Link>
                </li>
              );
            })}
          </ul>
        )}
      </AsyncPanel>

      <DashboardTablePagination
        disabled={refetching}
        itemLabel="conversations"
        onPageChange={setPage}
        page={pagination?.page ?? page}
        pageSize={pagination?.page_size ?? PAGE_SIZE}
        showPageNumbers={false}
        showRange={false}
        totalItems={pagination?.total ?? conversations.length}
        totalPages={pagination?.total_pages ?? 1}
      />
    </section>
  );
}
