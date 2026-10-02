"use client";

import { useState } from "react";

import { DashboardFilterGroup } from "@/components/ui/dashboard-filter-group";
import { DashboardTableActionButton } from "@/components/ui/dashboard-table-action-button";
import { DashboardTableRowActions } from "@/components/ui/dashboard-table-actions";
import { DashboardTableDateTimeCell } from "@/components/ui/dashboard-table-datetime-cell";
import { DashboardTablePagination } from "@/components/ui/dashboard-table-pagination";
import { DashboardTableStateRows } from "@/components/ui/dashboard-table-state-rows";
import { DashboardTableWrap } from "@/components/ui/dashboard-table-wrap";
import { RatingStars } from "@/components/ui/rating-stars";
import { reviewStatusToPillVariant, StatusPill } from "@/components/ui/status-pill";
import { DASHBOARD_TABLE_PAGE_SIZE } from "@/constants/dashboard-table";
import { getQuerySurface } from "@/lib/react-query/query-surface";
import type { ReviewStatus } from "@/lib/schemas/review";

import { useModerateReview, useReviews } from "../hooks";

const PAGE_SIZE = DASHBOARD_TABLE_PAGE_SIZE;
const COLUMN_COUNT = 6;

const STATUS_LABEL: Record<ReviewStatus, string> = {
  pending: "Pending",
  published: "Published",
  hidden: "Hidden",
};

const FILTERS: Array<{ value: ReviewStatus | undefined; label: string }> = [
  { value: "pending", label: "Pending" },
  { value: "published", label: "Published" },
  { value: "hidden", label: "Hidden" },
  { value: undefined, label: "All" },
];

/** Reviews to publish or hide, the ones waiting first. */
export function ManagerReviewsTable() {
  const [page, setPage] = useState(1);
  const [status, setStatus] = useState<ReviewStatus | undefined>("pending");
  const reviewsQuery = useReviews({ page, page_size: PAGE_SIZE, status });
  const moderate = useModerateReview();
  const { initialLoading, refetching } = getQuerySurface(reviewsQuery);
  const reviews = reviewsQuery.data?.reviews ?? [];
  const pagination = reviewsQuery.data?.pagination;

  return (
    <>
      <div className="db-table-filters db-table-filters--end">
        <DashboardFilterGroup
          aria-label="Filter by status"
          onChange={(next) => {
            setStatus(next);
            setPage(1);
          }}
          options={FILTERS}
          value={status}
        />
      </div>

      <DashboardTableWrap refetching={refetching}>
        <table className="db-table db-table--reviews">
          <colgroup>
            <col className="db-table-col-product" />
            <col className="db-table-col-review" />
            <col className="db-table-col-code" />
            <col className="db-table-col-datetime" />
            <col className="db-table-col-status" />
            <col className="db-table-col-actions" />
          </colgroup>
          <thead>
            <tr>
              <th>Product</th>
              <th>Review</th>
              <th>Order</th>
              <th>Written</th>
              <th className="db-table-status">Status</th>
              <th className="db-table-detail">Actions</th>
            </tr>
          </thead>
          <tbody>
            <DashboardTableStateRows
              columnCount={COLUMN_COUNT}
              emptyFilteredMessage={status === "pending" ? "No review is waiting for you." : undefined}
              entityLabel="reviews"
              hasActiveFilter={status !== undefined}
              isEmpty={reviews.length === 0}
              isError={reviewsQuery.isError}
              initialLoading={initialLoading}
            />
            {!initialLoading && !reviewsQuery.isError
              ? reviews.map((review) => {
                  const updating = moderate.isPending && moderate.variables?.id === review.id;
                  return (
                    <tr key={review.id}>
                      <td className="db-table-cell-primary">{review.product_name}</td>
                      <td>
                        <div className="db-table-review">
                          <RatingStars rating={review.rating} />
                          {review.comment ? (
                            <p className="review-comment">{review.comment}</p>
                          ) : (
                            <p className="text-caption-dashboard text-muted">No comment</p>
                          )}
                        </div>
                      </td>
                      <td>
                        <div className="db-table-stacked-cell min-w-0">
                          <span className="text-order-code">{review.order_code}</span>
                          <span className="truncate text-caption-dashboard text-muted">
                            {review.customer_name ?? "No name given"}
                          </span>
                        </div>
                      </td>
                      <DashboardTableDateTimeCell iso={review.created_at} />
                      <td className="db-table-status">
                        <div className="db-table-stacked-cell min-w-0">
                          <StatusPill label={STATUS_LABEL[review.status]} variant={reviewStatusToPillVariant(review.status)} />
                          {review.moderator_name ? (
                            <span className="truncate text-caption-dashboard text-muted">by {review.moderator_name}</span>
                          ) : null}
                        </div>
                      </td>
                      <td className="db-table-detail">
                        <DashboardTableRowActions>
                          <DashboardTableActionButton
                            blockedReason={
                              updating ? "Updating the review" : review.status === "hidden" ? "Already hidden" : undefined
                            }
                            label={`Hide the review of ${review.product_name} on ${review.order_code}`}
                            onClick={() => moderate.mutate({ id: review.id, status: "hidden" })}
                            text="Hide"
                            tone="warning"
                          />
                          <DashboardTableActionButton
                            blockedReason={
                              updating
                                ? "Updating the review"
                                : review.status === "published"
                                  ? "Already published"
                                  : undefined
                            }
                            label={`Publish the review of ${review.product_name} on ${review.order_code}`}
                            onClick={() => moderate.mutate({ id: review.id, status: "published" })}
                            text="Publish"
                            tone="accent"
                          />
                        </DashboardTableRowActions>
                      </td>
                    </tr>
                  );
                })
              : null}
          </tbody>
        </table>
        <DashboardTablePagination
          disabled={refetching}
          itemLabel="reviews"
          onPageChange={setPage}
          page={pagination?.page ?? page}
          pageSize={pagination?.page_size ?? PAGE_SIZE}
          totalItems={pagination?.total ?? reviews.length}
          totalPages={pagination?.total_pages ?? 1}
        />
      </DashboardTableWrap>
    </>
  );
}
