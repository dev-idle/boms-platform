import type { ReactNode } from "react";

import { cn } from "@/lib/utils";

export type ReadonlyFieldRow = {
  label: string;
  value: string;
  /** Codes and IDs render in the mono face. */
  mono?: boolean;
};

type ReadonlyFieldGroupProps = {
  heading: string;
  /** Why these values are locked and where they change — stated once. */
  reason: ReactNode;
  rows: readonly ReadonlyFieldRow[];
};

/** Values the user cannot edit: one grouped `<dl>` with a single reason. */
export function ReadonlyFieldGroup({ heading, reason, rows }: ReadonlyFieldGroupProps) {
  return (
    <div className="field-readonly-group">
      <p className="field-readonly-group__label">{heading}</p>
      <dl className="field-readonly-group__list">
        {rows.map((row) => (
          <div className="field-readonly-group__row" key={row.label}>
            <dt>{row.label}</dt>
            <dd className={cn(row.mono && "field-readonly-group__value--code")}>
              {row.value}
            </dd>
          </div>
        ))}
      </dl>
      <p className="field-readonly-group__reason">{reason}</p>
    </div>
  );
}
