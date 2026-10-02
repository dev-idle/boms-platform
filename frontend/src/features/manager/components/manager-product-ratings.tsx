"use client";

import { DashboardProfileSection } from "@/components/layouts/dashboard-profile-layout";
import { DashboardTableStateRows } from "@/components/ui/dashboard-table-state-rows";
import { DashboardTableWrap } from "@/components/ui/dashboard-table-wrap";
import { RatingStars } from "@/components/ui/rating-stars";
import { getQuerySurface } from "@/lib/react-query/query-surface";
import { formatAverageRating } from "@/lib/schemas/review";

import { useReviewSummary } from "../hooks";

const COLUMN_COUNT = 3;

/** Each reviewed product's rating, the lowest first: what needs attention leads. */
export function ManagerProductRatings() {
  const summaryQuery = useReviewSummary();
  const { initialLoading, refetching } = getQuerySurface(summaryQuery);
  const products = summaryQuery.data?.products ?? [];

  return (
    <DashboardProfileSection
      description="The lowest average first. A hidden review does not count; a pending one does."
      id="manager-product-ratings"
      title="Ratings by product"
      variant="plain"
    >
      <DashboardTableWrap refetching={refetching}>
        <table className="db-table db-table--ratings">
          <colgroup>
            <col />
            <col className="db-table-col-number" />
            <col className="db-table-col-rating" />
          </colgroup>
          <thead>
            <tr>
              <th>Product</th>
              <th className="db-table-num">Reviews</th>
              <th className="db-table-num">Average</th>
            </tr>
          </thead>
          <tbody>
            <DashboardTableStateRows
              columnCount={COLUMN_COUNT}
              emptyMessage="No product has a review yet."
              entityLabel="ratings"
              isEmpty={products.length === 0}
              isError={summaryQuery.isError}
              initialLoading={initialLoading}
            />
            {!initialLoading && !summaryQuery.isError
              ? products.map((product) => (
                  <tr key={product.product_id}>
                    <td className="db-table-cell-primary">{product.product_name}</td>
                    <td className="db-table-num">{product.review_count}</td>
                    <td className="db-table-num">
                      <span className="db-table-rating">
                        <RatingStars rating={product.average_rating} />
                        {formatAverageRating(product.average_rating)}
                      </span>
                    </td>
                  </tr>
                ))
              : null}
          </tbody>
        </table>
      </DashboardTableWrap>
    </DashboardProfileSection>
  );
}
