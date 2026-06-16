import type { ReactNode } from "react";

import { LOADING_MESSAGE } from "@/constants/loading-copy";
import { cn } from "@/lib/utils";

type DashboardTableWrapProps = {
  children: ReactNode;
  className?: string;
  refetching: boolean;
};

/** Dashboard data table shell — refetch affordance + a11y during background fetch. */
export function DashboardTableWrap({
  children,
  className,
  refetching,
}: DashboardTableWrapProps) {
  return (
    <div
      aria-busy={refetching || undefined}
      aria-label={refetching ? LOADING_MESSAGE : undefined}
      className={cn("db-table-wrap", refetching && "is-refetching", className)}
    >
      {children}
    </div>
  );
}
