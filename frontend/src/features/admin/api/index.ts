import { z } from "zod";

import {
  browserRequest,
  browserRequestVoid,
  browserRequestWithMeta,
} from "@/lib/browser-api-client";

import {
  adminResetPasswordResponseSchema,
  adminUserSchema,
  createOperationalResponseSchema,
  createOperationalSchema,
  listFilterSchema,
  updateRoleSchema,
  adminUserActivityLogSchema,
  type AdminUser,
  type AdminUserActivityLog,
  type CreateOperationalInput,
  type AdminResetPasswordResponse,
  type CreateOperationalResponse,
  type ListFilterInput,
  type UpdateRoleInput,
  type UserActivityFilterInput,
  type UserActivityListResult,
  type UsersListResult,
  userActivityFilterSchema,
} from "../schemas";

const paginatedMetaSchema = z.object({
  request_id: z.string().optional(),
  pagination: z
    .object({
      page: z.number().int().min(1),
      page_size: z.number().int().min(1),
      total: z.number().int().min(0),
      total_pages: z.number().int().min(1).optional(),
    })
    .optional(),
});

type PaginatedMeta = z.infer<typeof paginatedMetaSchema>;

function resolvePaginatedResult<TEntry>(
  result: { data: TEntry[]; meta?: unknown },
  fallback: { page: number; page_size: number },
): {
  entries: TEntry[];
  pagination: {
    page: number;
    page_size: number;
    total: number;
    total_pages: number;
  };
  request_id?: string;
} {
  const parsedMetaResult = paginatedMetaSchema.safeParse(result.meta);
  const parsedMeta: PaginatedMeta | undefined = parsedMetaResult.success
    ? parsedMetaResult.data
    : undefined;
  const rawPagination = parsedMeta?.pagination;

  const page = rawPagination?.page ?? fallback.page;
  const pageSize = rawPagination?.page_size ?? fallback.page_size;
  const total = rawPagination?.total ?? result.data.length;
  const totalPages =
    rawPagination?.total_pages ?? Math.max(1, Math.ceil(total / pageSize));

  return {
    entries: result.data,
    pagination: {
      page,
      page_size: pageSize,
      total,
      total_pages: totalPages,
    },
    request_id: parsedMeta?.request_id,
  };
}

function usersPath(filter: ListFilterInput): string {
  const params = new URLSearchParams();
  params.set("page", String(filter.page));
  params.set("page_size", String(filter.page_size));
  if (filter.search) {
    params.set("search", filter.search);
  }
  if (filter.role) {
    params.set("role", filter.role);
  }
  return `/api/v1/admin/users?${params.toString()}`;
}

export async function listUsers(input: ListFilterInput): Promise<UsersListResult> {
  const filter = listFilterSchema.parse(input);

  const result = await browserRequestWithMeta<AdminUser[]>(usersPath(filter), {
    method: "GET",
    schema: z.array(adminUserSchema),
  });

  const resolved = resolvePaginatedResult(result, filter);

  return {
    users: resolved.entries,
    pagination: resolved.pagination,
    request_id: resolved.request_id,
  };
}

export async function getUserById(id: string): Promise<AdminUser> {
  const parsedId = z.string().uuid().parse(id);

  return browserRequest<AdminUser>(`/api/v1/admin/users/${parsedId}`, {
    method: "GET",
    schema: adminUserSchema,
  });
}

export async function createOperational(
  input: CreateOperationalInput,
): Promise<CreateOperationalResponse> {
  const body = createOperationalSchema.parse(input);
  return browserRequest<CreateOperationalResponse>("/api/v1/admin/users", {
    method: "POST",
    json: body,
    schema: createOperationalResponseSchema,
  });
}

export async function updateRole(
  id: string,
  input: UpdateRoleInput,
): Promise<AdminUser> {
  const parsedId = z.string().uuid().parse(id);
  const body = updateRoleSchema.parse(input);
  return browserRequest<AdminUser>(`/api/v1/admin/users/${parsedId}/role`, {
    method: "PATCH",
    json: body,
    schema: adminUserSchema,
  });
}

export async function disableUser(id: string): Promise<void> {
  const parsedId = z.string().uuid().parse(id);
  await browserRequestVoid(`/api/v1/admin/users/${parsedId}/disable`, {
    method: "PATCH",
  });
}

export async function enableUser(id: string): Promise<void> {
  const parsedId = z.string().uuid().parse(id);
  await browserRequestVoid(`/api/v1/admin/users/${parsedId}/enable`, {
    method: "PATCH",
  });
}

export async function resetUserPassword(id: string): Promise<AdminResetPasswordResponse> {
  const parsedId = z.string().uuid().parse(id);
  return browserRequest<AdminResetPasswordResponse>(
    `/api/v1/admin/users/${parsedId}/reset-password`,
    {
      method: "POST",
      schema: adminResetPasswordResponseSchema,
    },
  );
}

export async function revokeUserSessions(id: string): Promise<void> {
  const parsedId = z.string().uuid().parse(id);
  await browserRequestVoid(`/api/v1/admin/users/${parsedId}/revoke-sessions`, {
    method: "POST",
  });
}

function userActivityPath(id: string, filter: UserActivityFilterInput): string {
  const params = new URLSearchParams();
  params.set("page", String(filter.page));
  params.set("page_size", String(filter.page_size));
  return `/api/v1/admin/users/${id}/activity?${params.toString()}`;
}

export async function listUserActivity(
  id: string,
  input: UserActivityFilterInput,
): Promise<UserActivityListResult> {
  const parsedId = z.string().uuid().parse(id);
  const filter = userActivityFilterSchema.parse(input);

  const result = await browserRequestWithMeta<AdminUserActivityLog[]>(
    userActivityPath(parsedId, filter),
    {
      method: "GET",
      schema: z.array(adminUserActivityLogSchema),
    },
  );

  const resolved = resolvePaginatedResult(result, filter);

  return {
    entries: resolved.entries,
    pagination: resolved.pagination,
    request_id: resolved.request_id,
  };
}
