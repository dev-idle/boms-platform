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

export function DashboardCalendarIcon({ className }: DashboardUiIconProps) {
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
      <path d="M8 2v4" />
      <path d="M16 2v4" />
      <rect height="18" rx="2" width="18" x="3" y="4" />
      <path d="M3 10h18" />
    </svg>
  );
}

export function DashboardChevronLeftIcon({ className }: DashboardUiIconProps) {
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
      <path d="m15 18-6-6 6-6" />
    </svg>
  );
}

export function DashboardChevronRightIcon({ className }: DashboardUiIconProps) {
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
      <path d="m9 18 6-6-6-6" />
    </svg>
  );
}

export function DashboardChevronUpIcon({ className }: DashboardUiIconProps) {
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
      <path d="m6 15 6-6 6 6" />
    </svg>
  );
}
