type DashboardStep = {
  label: string;
  onClick: () => void;
  /** Which way the step goes, drawn as an arrow beside the label. */
  direction?: "previous" | "next";
  /** When set, the step cannot apply: greyed in place, with the reason as its title. */
  blockedReason?: string;
};

type DashboardStepGroupProps = {
  "aria-label": string;
  steps: readonly DashboardStep[];
};

/**
 * Moves a day or week view — previous, back to now, next — as one segmented
 * group beside its date field, drawn as the filter groups are.
 */
export function DashboardStepGroup({ "aria-label": ariaLabel, steps }: DashboardStepGroupProps) {
  return (
    <div aria-label={ariaLabel} className="db-filter-group" role="group">
      {steps.map((step) => (
        <button
          className="db-filter-chip"
          disabled={Boolean(step.blockedReason)}
          key={step.label}
          title={step.blockedReason}
          type="button"
          onClick={step.onClick}
        >
          {step.direction === "previous" ? <span aria-hidden="true">‹</span> : null}
          {step.label}
          {step.direction === "next" ? <span aria-hidden="true">›</span> : null}
        </button>
      ))}
    </div>
  );
}
