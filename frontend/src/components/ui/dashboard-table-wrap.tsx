import type { ReactNode } from "react";

import { cn } from "@/lib/utils";

import { AsyncPanel } from "./async-panel";

type DashboardTableWrapProps = {
  children: ReactNode;
  className?: string;
  refetching: boolean;
};

/** Table card — refetch overlay only; initial load spinner lives in tbody row 2. */
export function DashboardTableWrap({
  children,
  className,
  refetching,
}: DashboardTableWrapProps) {
  return (
    <AsyncPanel
      className={cn("db-table-wrap", className)}
      overlayOnInitialLoad={false}
      refetching={refetching}
    >
      {children}
    </AsyncPanel>
  );
}
