"use client";

import { Button } from "@/components/ui/button";

import { useOperations } from "../hooks";
import type { Operations } from "../schemas";

const KPIS: ReadonlyArray<{ label: string; value: (operations: Operations) => string }> = [
  {
    label: "Active orders",
    value: ({ orders }) => String(orders.pending + orders.confirmed + orders.in_production + orders.ready),
  },
  { label: "To collect today", value: ({ pickups }) => String(pickups.to_collect) },
  { label: "Late", value: ({ late }) => String(late.total) },
  {
    label: "Production today",
    value: ({ production }) => (production.average_minutes === null ? "—" : `${production.average_minutes} min`),
  },
];

/** The work in four numbers: orders still open, today's pickups still to come, the late ones, and today's production time. */
export function ManagerOperationsKpis() {
  const operationsQuery = useOperations();
  const operations = operationsQuery.data;

  if (operationsQuery.isError) {
    return (
      <div className="db-kpi-error">
        <p className="db-kpi-error__title">We could not read where the work stands.</p>
        <Button type="button" variant="outline" onClick={() => void operationsQuery.refetch()}>
          Try again
        </Button>
      </div>
    );
  }

  return (
    <dl aria-busy={operations ? undefined : true} className="db-kpi-strip">
      {KPIS.map((kpi) => (
        <div className="db-kpi" key={kpi.label}>
          <dt className="db-kpi__label">{kpi.label}</dt>
          <dd className="db-kpi__value">
            {operations ? kpi.value(operations) : <span className="skeleton db-kpi__skeleton" />}
          </dd>
        </div>
      ))}
    </dl>
  );
}
