import { queryOptions } from "@tanstack/react-query";

import { getMe } from "../api";

export const authQueryKeys = {
  me: ["auth", "me"] as const,
};

export function meQueryOptions() {
  return queryOptions({
    queryKey: authQueryKeys.me,
    queryFn: getMe,
    staleTime: 5 * 60_000,
  });
}
