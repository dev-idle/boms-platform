import Link from "next/link";

import { DashboardOpenDetailIcon } from "@/components/icons/dashboard-ui-icons";
import { cn } from "@/lib/utils";

type DashboardTableActionLinkProps = {
  className?: string;
  href: string;
  label?: string;
};

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
      <DashboardOpenDetailIcon className="db-table-action-icon" />
    </Link>
  );
}
