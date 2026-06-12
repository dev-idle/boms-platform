"use client";

import { useState } from "react";

import { DashboardTablePagination } from "@/components/ui/dashboard-table-pagination";
import { DashboardProfileSection } from "@/features/user";
import { formatDateTime } from "@/lib/validation/datetime";
import { cn } from "@/lib/utils";

import { useUserActivity } from "../hooks";

const PAGE_SIZE = 10;

type AdminUserActivityLogProps = {
  userId: string;
};

export function AdminUserActivityLog({ userId }: AdminUserActivityLogProps) {
  const [page, setPage] = useState(1);
  const activityQuery = useUserActivity(userId, {
    page,
    page_size: PAGE_SIZE,
  });

  const entries = activityQuery.data?.entries ?? [];
  const pagination = activityQuery.data?.pagination;

  return (
    <DashboardProfileSection
      description="Administrative and account events for this user."
      id="admin-user-activity"
      title="Activity log"
    >
      <div
        className={cn(
          "db-table-wrap",
          activityQuery.isFetching && "is-refetching",
        )}
      >
        <table className="db-table">
          <thead>
            <tr>
              <th>When</th>
              <th>Event</th>
              <th>Actor</th>
            </tr>
          </thead>
          <tbody>
            {activityQuery.isPending ? (
              <tr>
                <td className="text-muted" colSpan={3}>
                  Loading activity…
                </td>
              </tr>
            ) : activityQuery.isError ? (
              <tr>
                <td className="text-error" colSpan={3}>
                  Failed to load activity log.
                </td>
              </tr>
            ) : entries.length === 0 ? (
              <tr>
                <td className="text-muted" colSpan={3}>
                  No activity recorded yet.
                </td>
              </tr>
            ) : (
              entries.map((entry) => (
                <tr key={entry.id}>
                  <td className="text-sm text-muted whitespace-nowrap">
                    {formatDateTime(entry.created_at)}
                  </td>
                  <td className="text-sm text-ink">{entry.summary}</td>
                  <td className="text-sm">
                    <span className="text-ink">{entry.actor_email}</span>
                    <span className="mt-0.5 block text-xs uppercase text-muted">
                      {entry.actor_role}
                    </span>
                  </td>
                </tr>
              ))
            )}
          </tbody>
        </table>
        <DashboardTablePagination
          disabled={activityQuery.isFetching}
          hideWhenSinglePage
          itemLabel="events"
          onPageChange={setPage}
          page={pagination?.page ?? page}
          pageSize={pagination?.page_size ?? PAGE_SIZE}
          totalItems={pagination?.total ?? entries.length}
          totalPages={pagination?.total_pages ?? 1}
        />
      </div>
    </DashboardProfileSection>
  );
}
