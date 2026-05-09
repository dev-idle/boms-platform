import Image from "next/image";
import Link from "next/link";

import {
  AUTH_ACCESS_COOKIE,
  AUTH_REFRESH_COOKIE,
} from "@/constants/cookies";
import { ROUTE } from "@/constants/routes";

export default function HomePage() {
  return (
    <div className="flex min-h-full flex-col items-center justify-center bg-zinc-50 px-6 py-24 font-sans dark:bg-black">
      <main className="flex w-full max-w-2xl flex-col gap-10 rounded-2xl border border-zinc-200 bg-white p-12 shadow-sm dark:border-zinc-800 dark:bg-zinc-950">
        <div className="flex flex-col gap-4">
          <p className="text-sm font-medium uppercase tracking-wide text-zinc-500 dark:text-zinc-400">
            Bakery Ordering and Management System
          </p>
          <h1 className="text-3xl font-semibold tracking-tight text-zinc-900 dark:text-zinc-50">
            BOMS storefront
          </h1>
          <p className="text-lg leading-relaxed text-zinc-600 dark:text-zinc-300">
            RSC-first ordering surface. Authenticated areas require HttpOnly{" "}
            <code className="rounded bg-zinc-100 px-1 py-0.5 text-sm dark:bg-zinc-900">
              {AUTH_ACCESS_COOKIE}
            </code>{" "}
            and{" "}
            <code className="rounded bg-zinc-100 px-1 py-0.5 text-sm dark:bg-zinc-900">
              {AUTH_REFRESH_COOKIE}
            </code>{" "}
            cookies issued by Fiber.
          </p>
        </div>
        <div className="flex flex-wrap gap-3 text-sm font-medium">
          <Link
            className="rounded-full bg-zinc-900 px-5 py-2.5 text-white transition hover:bg-zinc-800 dark:bg-zinc-100 dark:text-zinc-900 dark:hover:bg-white"
            href={ROUTE.products}
          >
            Browse products
          </Link>
          <Link
            className="rounded-full border border-zinc-300 px-5 py-2.5 text-zinc-900 transition hover:border-zinc-400 dark:border-zinc-700 dark:text-zinc-100 dark:hover:border-zinc-500"
            href={ROUTE.login}
          >
            Sign in
          </Link>
        </div>
        <div className="flex items-center gap-3 border-t border-zinc-100 pt-8 dark:border-zinc-800">
          <Image
            className="dark:invert"
            src="/next.svg"
            alt="Next.js"
            width={90}
            height={18}
            priority
          />
          <span className="text-xs text-zinc-500 dark:text-zinc-400">
            Next.js 16 · React 19 · Tailwind v4
          </span>
        </div>
      </main>
    </div>
  );
}
