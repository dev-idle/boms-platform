import Link from "next/link";
import type { ReactNode } from "react";

import { ROUTE } from "@/constants/routes";

export default function CustomerLayout({ children }: { children: ReactNode }) {
  return (
    <div className="min-h-full bg-zinc-50 dark:bg-black">
      <header className="border-b border-zinc-200 bg-white dark:border-zinc-800 dark:bg-zinc-950">
        <nav className="mx-auto flex max-w-5xl items-center gap-6 px-6 py-4 text-sm font-medium text-zinc-700 dark:text-zinc-200">
          <Link href={ROUTE.products}>Products</Link>
          <Link href={ROUTE.cart}>Cart</Link>
          <Link href={ROUTE.orders}>Orders</Link>
          <Link className="ml-auto text-zinc-500 dark:text-zinc-400" href={ROUTE.home}>
            Exit
          </Link>
        </nav>
      </header>
      <div className="mx-auto max-w-5xl px-6 py-10">{children}</div>
    </div>
  );
}
