import Link from "next/link";

import { ROUTE } from "@/constants/routes";

export default function RegisterPage() {
  return (
    <div className="mx-auto flex min-h-full max-w-md flex-col justify-center px-6 py-16">
      <h1 className="text-2xl font-semibold text-zinc-900 dark:text-zinc-50">Create account</h1>
      <p className="mt-2 text-sm text-zinc-600 dark:text-zinc-400">
        Registration flows should validate with Zod on the server and delegate persistence to Fiber.
      </p>
      <Link className="mt-8 text-sm font-medium text-zinc-900 underline dark:text-zinc-100" href={ROUTE.login}>
        Already registered?
      </Link>
    </div>
  );
}
