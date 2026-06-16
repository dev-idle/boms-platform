/** Shared stroke icons for internal dashboard UI (nav uses `dashboard-nav-icons`). */

const DASHBOARD_ICON_STROKE = 1.75;

type DashboardUiIconProps = {
  className?: string;
};

export function DashboardCloseIcon({ className }: DashboardUiIconProps) {
  return (
    <svg
      aria-hidden
      className={className}
      fill="none"
      stroke="currentColor"
      strokeLinecap="round"
      strokeLinejoin="round"
      strokeWidth={DASHBOARD_ICON_STROKE}
      viewBox="0 0 24 24"
    >
      <path d="M18 6 6 18" />
      <path d="m6 6 12 12" />
    </svg>
  );
}

/** Lucide-style pencil-on-line — canonical edit affordance at dashboard action size. */
export function DashboardEditIcon({ className }: DashboardUiIconProps) {
  return (
    <svg
      aria-hidden
      className={className}
      fill="none"
      stroke="currentColor"
      strokeLinecap="round"
      strokeLinejoin="round"
      strokeWidth={DASHBOARD_ICON_STROKE}
      viewBox="0 0 24 24"
    >
      <path d="M12 20h9" />
      <path d="M16.5 3.5a2.12 2.12 0 0 1 3 3L7 19l-4 1 1-4Z" />
    </svg>
  );
}

export function DashboardDeleteIcon({ className }: DashboardUiIconProps) {
  return (
    <svg
      aria-hidden
      className={className}
      fill="none"
      stroke="currentColor"
      strokeLinecap="round"
      strokeLinejoin="round"
      strokeWidth={DASHBOARD_ICON_STROKE}
      viewBox="0 0 24 24"
    >
      <path d="M3 6h18" />
      <path d="M8 6V4h8v2" />
      <path d="M19 6v14a2 2 0 0 1-2 2H7a2 2 0 0 1-2-2V6" />
      <path d="M10 11v6" />
      <path d="M14 11v6" />
    </svg>
  );
}

export function DashboardOpenDetailIcon({ className }: DashboardUiIconProps) {
  return (
    <svg
      aria-hidden
      className={className}
      fill="none"
      stroke="currentColor"
      strokeLinecap="round"
      strokeLinejoin="round"
      strokeWidth={DASHBOARD_ICON_STROKE}
      viewBox="0 0 24 24"
    >
      <path d="M7 17 17 7" />
      <path d="M7 7h10v10" />
    </svg>
  );
}

/** Star outline — set gallery image as primary (first in list). */
export function DashboardPrimaryIcon({ className }: DashboardUiIconProps) {
  return (
    <svg
      aria-hidden
      className={className}
      fill="none"
      stroke="currentColor"
      strokeLinecap="round"
      strokeLinejoin="round"
      strokeWidth={DASHBOARD_ICON_STROKE}
      viewBox="0 0 24 24"
    >
      <path d="m12 3 2.35 4.76 5.25.77-3.8 3.7.9 5.23L12 15.77l-4.7 2.47.9-5.23-3.8-3.7 5.25-.77L12 3Z" />
    </svg>
  );
}

/** Eye icon — preview / view affordance in dashboard forms and tables. */
export function DashboardViewIcon({ className }: DashboardUiIconProps) {
  return (
    <svg
      aria-hidden
      className={className}
      fill="none"
      stroke="currentColor"
      strokeLinecap="round"
      strokeLinejoin="round"
      strokeWidth={DASHBOARD_ICON_STROKE}
      viewBox="0 0 24 24"
    >
      <path d="M2 12s3.5-7 10-7 10 7 10 7-3.5 7-10 7-10-7-10-7Z" />
      <circle cx="12" cy="12" r="3" />
    </svg>
  );
}
