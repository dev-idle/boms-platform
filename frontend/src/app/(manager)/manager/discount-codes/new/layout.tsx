import type { ReactNode } from "react";

import { PAGE_TITLES, pageTitle } from "@/lib/metadata/page-title";

export const metadata = pageTitle(PAGE_TITLES.newDiscountCode);

export default function ManagerNewDiscountCodeLayout({
  children,
}: {
  children: ReactNode;
}) {
  return children;
}
