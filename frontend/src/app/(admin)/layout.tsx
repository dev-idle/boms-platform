import Link from "next/link";
import type { ReactNode } from "react";

import { ROUTE } from "@/constants/routes";

export default function AdminLayout({ children }: { children: ReactNode }) {
  return (
    <div className="min-h-full bg-zinc-100 dark:bg-zinc-950">
      <aside className="fixed inset-y-0 left-0 w-56 border-r border-zinc-200 bg-white px-4 py-6 dark:border-zinc-800 dark:bg-zinc-900">
        <p className="text-xs font-semibold uppercase tracking-wide text-zinc-500">Admin</p>
        <nav className="mt-6 flex flex-col gap-3 text-sm font-medium text-zinc-800 dark:text-zinc-100">
          <Link href={ROUTE.admin.dashboard}>Dashboard</Link>
          <Link href={ROUTE.admin.products}>Products</Link>
          <Link href={ROUTE.admin.orders}>Orders</Link>
          <Link href={ROUTE.admin.users}>Users</Link>
        </nav>
      </aside>
      <div className="pl-56">
        <div className="mx-auto max-w-6xl px-8 py-10">{children}</div>
      </div>
    </div>
  );
}
