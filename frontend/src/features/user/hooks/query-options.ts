import { queryOptions } from "@tanstack/react-query";

import { getMe } from "../api";

export const userQueryKeys = {
  me: ["user", "me"] as const,
};

export function meQueryOptions() {
  return queryOptions({
    queryKey: userQueryKeys.me,
    queryFn: getMe,
    staleTime: 5 * 60_000,
  });
}
