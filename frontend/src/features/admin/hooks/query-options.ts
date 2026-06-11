import type { ListFilterInput, UserActivityFilterInput } from "../schemas";

export const adminQueryKeys = {
  users: (filter: ListFilterInput) =>
    ["admin", "users", filter.page, filter.page_size, filter.search, filter.role] as const,
  usersRoot: ["admin", "users"] as const,
  user: (id: string) => ["admin", "user", id] as const,
  userActivity: (id: string, filter: UserActivityFilterInput) =>
    ["admin", "user", id, "activity", filter.page, filter.page_size] as const,
  userActivityRoot: (id: string) => ["admin", "user", id, "activity"] as const,
};
