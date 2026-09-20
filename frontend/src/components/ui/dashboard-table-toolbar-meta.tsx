import type { ReactNode } from "react";

/** Result count opposite the search field — matches Premium dashboard tables. */
export function DashboardTableToolbarMeta({ children }: { children: ReactNode }) {
  return <p className="db-table-toolbar-meta">{children}</p>;
}
