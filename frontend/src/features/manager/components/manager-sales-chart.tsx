import { formatPriceCents } from "@/lib/validation/catalog";

import { formatPeriod } from "../lib/sales-report";
import type { SalesReport } from "../schemas";

/** The share of each period's width a bar leaves empty, half on each side. */
const BAR_GAP = 0.2;
/** The plot is drawn in units: one per period across, 100 down. */
const PLOT_HEIGHT = 100;

type ManagerSalesChartProps = {
  report: SalesReport;
};

/**
 * Net revenue by period as bars over a baseline: a period that gave back more
 * than it took dips below it. Drawn as plain SVG and stretched to the width;
 * the scale and the first and last periods stay HTML so they keep their size,
 * hidden from screen readers, which hear the image's label and read the table.
 */
export function ManagerSalesChart({ report }: ManagerSalesChartProps) {
  const { periods, group } = report;
  const first = periods[0];
  const last = periods[periods.length - 1];
  if (!first || !last) {
    return null;
  }
  const nets = periods.map((period) => period.net_cents);
  const top = Math.max(0, ...nets);
  const bottom = Math.min(0, ...nets);
  const span = top - bottom;
  // Nothing moved: the baseline sits at the foot of the plot.
  const baseline = span === 0 ? PLOT_HEIGHT : (top / span) * PLOT_HEIGHT;

  return (
    <figure className="db-chart">
      <p aria-hidden className="db-chart__scale">
        {formatPriceCents(top)}
      </p>
      <svg
        aria-label={`Net revenue by ${group}, ${formatPeriod(first.start, group)} to ${formatPeriod(last.start, group)}; the table below lists each.`}
        className="db-chart__plot"
        preserveAspectRatio="none"
        role="img"
        viewBox={`0 0 ${periods.length} ${PLOT_HEIGHT}`}
      >
        {periods.map((period, index) => {
          const height = span === 0 ? 0 : (Math.abs(period.net_cents) / span) * PLOT_HEIGHT;
          return (
            <rect
              className={period.net_cents < 0 ? "db-chart__bar db-chart__bar--negative" : "db-chart__bar"}
              height={height}
              key={period.start}
              width={1 - BAR_GAP}
              x={index + BAR_GAP / 2}
              y={period.net_cents < 0 ? baseline : baseline - height}
            >
              <title>{`${formatPeriod(period.start, group)}: ${formatPriceCents(period.net_cents)}`}</title>
            </rect>
          );
        })}
        <line
          className="db-chart__baseline"
          shapeRendering="crispEdges"
          vectorEffect="non-scaling-stroke"
          x1={0}
          x2={periods.length}
          y1={baseline}
          y2={baseline}
        />
      </svg>
      <div aria-hidden className="db-chart__labels">
        <span>{formatPeriod(first.start, group)}</span>
        {last === first ? null : <span>{formatPeriod(last.start, group)}</span>}
      </div>
    </figure>
  );
}
