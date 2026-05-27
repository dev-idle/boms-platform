import Link from "next/link";
import type { ReactNode } from "react";

import { ROUTE } from "@/constants/routes";
import { BakerGate, LogoutButton } from "@/features/auth";

export default function BakerLayout({ children }: { children: ReactNode }) {
  return (
    <BakerGate>
      <div className="min-h-full bg-zinc-50 dark:bg-black">
        <header className="border-b border-zinc-200 bg-white dark:border-zinc-800 dark:bg-zinc-950">
          <nav className="mx-auto flex max-w-5xl items-center gap-6 px-6 py-4 text-sm font-medium text-zinc-700 dark:text-zinc-200">
            <span className="text-xs font-semibold uppercase tracking-wide text-amber-700 dark:text-amber-400">
              Baker
            </span>
            <Link href={ROUTE.baker.account.profile}>Profile</Link>
            <Link href={ROUTE.baker.account.password}>Password</Link>
            <div className="ml-auto flex items-center gap-3">
              <LogoutButton />
            </div>
          </nav>
        </header>
        <div className="mx-auto max-w-5xl px-6 py-10">{children}</div>
      </div>
    </BakerGate>
  );
}
