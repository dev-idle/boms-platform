"use client";

import { useCallback, useLayoutEffect, useRef, useState } from "react";

import { DashboardTablePagination } from "@/components/ui/dashboard-table-pagination";
import { DashboardActivityFeedPagePlaceholders } from "@/components/ui/dashboard-activity-feed-page-placeholders";
import { roleDisplayLabel } from "@/constants/roles";
import { DashboardProfileSection } from "@/components/layouts/dashboard-profile-layout";
import {
  isInitialQueryLoad,
  isQueryRefetching,
} from "@/lib/react-query/query-surface";
import { paginatedPlaceholderCountFromMeta } from "@/lib/pagination/dashboard-pagination";
import { formatDateTime } from "@/lib/validation/datetime";
import { cn } from "@/lib/utils";

import { useUserActivity } from "../hooks";

const PAGE_SIZE = 10;

type AdminUserActivityLogProps = {
  userId: string;
};

export function AdminUserActivityLog({ userId }: AdminUserActivityLogProps) {
  const [page, setPage] = useState(1);
  const scrollAnchorRef = useRef<number | null>(null);
  const activityQuery = useUserActivity(userId, {
    page,
    page_size: PAGE_SIZE,
  });

  const entries = activityQuery.data?.entries ?? [];
  const pagination = activityQuery.data?.pagination;
  const totalItems = pagination?.total ?? entries.length;
  const totalPages = pagination?.total_pages ?? 1;
  const initialLoad = isInitialQueryLoad(
    activityQuery.isPending,
    activityQuery.data,
  );
  const refetching = isQueryRefetching(
    activityQuery.isFetching,
    activityQuery.isPending,
    activityQuery.data,
  );
  const showPagination =
    !initialLoad && !activityQuery.isError && totalItems > 0;
  const pagePlaceholderCount = paginatedPlaceholderCountFromMeta(
    entries.length,
    pagination,
    PAGE_SIZE,
  );

  const handlePageChange = useCallback((nextPage: number) => {
    scrollAnchorRef.current = window.scrollY;
    setPage(nextPage);
  }, []);

  useLayoutEffect(() => {
    if (scrollAnchorRef.current === null) {
      return;
    }

    window.scrollTo(0, scrollAnchorRef.current);

    if (!activityQuery.isFetching) {
      scrollAnchorRef.current = null;
    }
  }, [activityQuery.data, activityQuery.isFetching, page]);

  return (
    <DashboardProfileSection
      description="Administrative and account events for this user."
      id="admin-user-activity"
      title="Activity log"
    >
      <div
        className={cn(
          "dashboard-activity-feed-panel",
          refetching && "is-refetching",
        )}
      >
        {initialLoad ? (
          <p className="dashboard-activity-feed-status">Loading activity…</p>
        ) : activityQuery.isError ? (
          <p className="dashboard-activity-feed-status text-error">
            Failed to load activity log.
          </p>
        ) : entries.length === 0 ? (
          <p className="dashboard-activity-feed-status">
            No activity recorded yet.
          </p>
        ) : (
          <ul className="dashboard-activity-feed">
            {entries.map((entry) => (
              <li key={entry.id} className="dashboard-activity-feed-item">
                <div className="dashboard-activity-feed-head">
                  <p className="dashboard-activity-feed-summary">
                    {entry.summary}
                  </p>
                  <time
                    className="dashboard-activity-feed-when"
                    dateTime={entry.created_at}
                  >
                    {formatDateTime(entry.created_at)}
                  </time>
                </div>
                <p className="dashboard-activity-feed-meta">
                  <span>{entry.actor_email}</span>
                  <span aria-hidden className="dashboard-meta-sep">
                    |
                  </span>
                  <span>{roleDisplayLabel(entry.actor_role)}</span>
                </p>
              </li>
            ))}
            <DashboardActivityFeedPagePlaceholders count={pagePlaceholderCount} />
          </ul>
        )}
        {showPagination ? (
          <DashboardTablePagination
            className="dashboard-activity-feed-pagination"
            disabled={refetching}
            hideWhenSinglePage
            itemLabel="events"
            onPageChange={handlePageChange}
            page={pagination?.page ?? page}
            pageSize={pagination?.page_size ?? PAGE_SIZE}
            totalItems={totalItems}
            totalPages={totalPages}
          />
        ) : null}
      </div>
    </DashboardProfileSection>
  );
}
