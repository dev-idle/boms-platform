import Link from "next/link";

import { cn } from "@/lib/utils";

type DashboardTableActionLinkProps = {
  className?: string;
  href: string;
  /** Accessible name — carries the row identity (e.g. "View order CH-1042"). */
  label?: string;
  showArrow?: boolean;
  /** Visible word. Words in a row need no legend; icons do. */
  text?: string;
};

/** Standard row action for internal dashboard tables. */
export function DashboardTableActionLink({
  className,
  href,
  label = "View details",
  showArrow = false,
  text = "View",
}: DashboardTableActionLinkProps) {
  return (
    <Link
      aria-label={label}
      className={cn("db-table-action db-table-action--detail", className)}
      href={href}
    >
      {text}
      {showArrow ? <span aria-hidden> →</span> : null}
    </Link>
  );
}
