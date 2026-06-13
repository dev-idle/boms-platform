import type { ReactNode } from "react";

import { PAGE_TITLES, pageTitle } from "@/lib/metadata/page-title";

export const metadata = pageTitle(PAGE_TITLES.editCategory);

export default function ManagerEditCategoryLayout({
  children,
}: {
  children: ReactNode;
}) {
  return children;
}
