import type { ReactNode } from "react";

import { PAGE_TITLES, pageTitle } from "@/lib/metadata/page-title";

export const metadata = pageTitle(PAGE_TITLES.newCombo);

export default function ManagerNewComboLayout({
  children,
}: {
  children: ReactNode;
}) {
  return children;
}
