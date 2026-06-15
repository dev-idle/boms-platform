import type { ReactNode } from "react";

import { PAGE_TITLES, pageTitle } from "@/lib/metadata/page-title";

export const metadata = pageTitle(PAGE_TITLES.newUser);

export default function AdminUsersNewLayout({
  children,
}: {
  children: ReactNode;
}) {
  return children;
}
