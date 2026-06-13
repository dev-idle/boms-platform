import type { ReactNode } from "react";

import { PAGE_TITLES, pageTitle } from "@/lib/metadata/page-title";

export const metadata = pageTitle(PAGE_TITLES.editDiscountCode);

export default function ManagerEditDiscountCodeLayout({
  children,
}: {
  children: ReactNode;
}) {
  return children;
}
