import Link from "next/link";

import { cn } from "@/lib/utils";

type DashboardTableActionLinkProps = {
  className?: string;
  href: string;
  label?: string;
};

function OpenDetailIcon({ className }: { className?: string }) {
  return (
    <svg
      aria-hidden
      className={className}
      fill="none"
      stroke="currentColor"
      strokeLinecap="round"
      strokeLinejoin="round"
      strokeWidth={2}
      viewBox="0 0 24 24"
    >
      <path d="M7 17 17 7" />
      <path d="M7 7h10v10" />
    </svg>
  );
}

/** Standard row action for internal dashboard tables. */
export function DashboardTableActionLink({
  className,
  href,
  label = "View details",
}: DashboardTableActionLinkProps) {
  return (
    <Link
      aria-label={label}
      className={cn("db-table-action db-table-action--detail", className)}
      href={href}
      title={label}
    >
      <OpenDetailIcon className="db-table-action-icon" />
    </Link>
  );
}
