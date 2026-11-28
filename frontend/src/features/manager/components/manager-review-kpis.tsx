"use client";

import { Button } from "@/components/ui/button";
import { formatAverageRating } from "@/lib/schemas/review";

import { useReviewSummary } from "../hooks";
import type { ReviewSummary } from "../schemas";

const KPIS: ReadonlyArray<{ label: string; value: (summary: ReviewSummary) => string }> = [
  {
    label: "Average rating",
    value: (summary) => (summary.average_rating === null ? "—" : formatAverageRating(summary.average_rating)),
  },
  { label: "Reviews", value: (summary) => String(summary.review_count) },
  { label: "4–5 stars", value: (summary) => String(summary.stars[3] + summary.stars[4]) },
  { label: "1–2 stars", value: (summary) => String(summary.stars[0] + summary.stars[1]) },
];

/** What customers think, in four numbers: hidden reviews do not count. */
export function ManagerReviewKpis() {
  const summaryQuery = useReviewSummary();
  const summary = summaryQuery.data;

  if (summaryQuery.isError) {
    return (
      <div className="db-kpi-error">
        <p className="db-kpi-error__title">We could not add up the reviews.</p>
        <Button type="button" variant="outline" onClick={() => void summaryQuery.refetch()}>
          Try again
        </Button>
      </div>
    );
  }

  return (
    <dl aria-busy={summary ? undefined : true} className="db-kpi-strip">
      {KPIS.map((kpi) => (
        <div className="db-kpi" key={kpi.label}>
          <dt className="db-kpi__label">{kpi.label}</dt>
          <dd className="db-kpi__value">{summary ? kpi.value(summary) : <span className="skeleton db-kpi__skeleton" />}</dd>
        </div>
      ))}
    </dl>
  );
}
