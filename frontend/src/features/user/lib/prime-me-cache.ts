import type { QueryClient } from "@tanstack/react-query";

import type { Me } from "../types";

import { userQueryKeys } from "../hooks/query-options";

/** Seed TanStack Query after an explicit GET /me (bootstrap, login). Avoids a second fetch from useMe. */
export function primeMeQueryCache(queryClient: QueryClient, user: Me): void {
  queryClient.setQueryData(userQueryKeys.me, user);
}
