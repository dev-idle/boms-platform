import { z } from "zod";

import {
  browserRequest,
  browserRequestVoid,
  browserRequestWithMeta,
} from "@/lib/browser-api-client";
import { parsePaginatedList } from "@/lib/pagination/parse-paginated-list";

import {
  adminResetPasswordResponseSchema,
  adminUserSchema,
  createOperationalResponseSchema,
  createOperationalSchema,
  listFilterSchema,
  nextEmployeeCodeSchema,
  updateRoleSchema,
  adminUserActivityLogSchema,
  type AdminUser,
  type AdminUserActivityLog,
  type CreateOperationalInput,
  type AdminResetPasswordResponse,
  type CreateOperationalResponse,
  type NextEmployeeCode,
  type ListFilterInput,
  type UpdateRoleInput,
  type UserActivityFilterInput,
  type UserActivityListResult,
  type UsersListResult,
  userActivityFilterSchema,
} from "../schemas";

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

  const parsed = parsePaginatedList(result.data, result.meta, filter);

  return {
    users: parsed.items,
    pagination: parsed.pagination,
    request_id: parsed.request_id,
  };
}

export async function getUserById(id: string): Promise<AdminUser> {
  const parsedId = z.string().uuid().parse(id);

  return browserRequest<AdminUser>(`/api/v1/admin/users/${parsedId}`, {
    method: "GET",
    schema: adminUserSchema,
  });
}

export async function getNextEmployeeCode(): Promise<NextEmployeeCode> {
  return browserRequest<NextEmployeeCode>("/api/v1/admin/employee-codes/next", {
    method: "GET",
    schema: nextEmployeeCodeSchema,
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

  const parsed = parsePaginatedList(result.data, result.meta, filter);

  return {
    entries: parsed.items,
    pagination: parsed.pagination,
    request_id: parsed.request_id,
  };
}
