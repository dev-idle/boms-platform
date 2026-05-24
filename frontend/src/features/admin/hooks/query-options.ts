import type { ListFilterInput } from "../schemas";

export const adminQueryKeys = {
  users: (filter: ListFilterInput) =>
    ["admin", "users", filter.page, filter.page_size, filter.search] as const,
  usersRoot: ["admin", "users"] as const,
  user: (id: string) => ["admin", "user", id] as const,
};
