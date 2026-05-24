import { z } from "zod";

import {
  browserRequest,
  browserRequestVoid,
  browserRequestWithMeta,
} from "@/lib/browser-api-client";

import {
  adminUserSchema,
  createOperationalResponseSchema,
  createOperationalSchema,
  listFilterSchema,
  updateOperationalProfileSchema,
  updateRoleSchema,
  type AdminUser,
  type CreateOperationalInput,
  type CreateOperationalResponse,
  type ListFilterInput,
  type UpdateOperationalProfileInput,
  type UpdateRoleInput,
  type UsersListResult,
} from "../schemas";

const usersMetaSchema = z.object({
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

function usersPath(filter: ListFilterInput): string {
  const params = new URLSearchParams();
  params.set("page", String(filter.page));
  params.set("page_size", String(filter.page_size));
  if (filter.search) {
    params.set("search", filter.search);
  }
  return `/api/v1/admin/users?${params.toString()}`;
}

export async function listUsers(input: ListFilterInput): Promise<UsersListResult> {
  const filter = listFilterSchema.parse(input);

  const result = await browserRequestWithMeta<AdminUser[]>(usersPath(filter), {
    method: "GET",
    schema: z.array(adminUserSchema),
  });

  const parsedMeta = usersMetaSchema.safeParse(result.meta);
  const rawPagination = parsedMeta.success ? parsedMeta.data.pagination : undefined;

  const page = rawPagination?.page ?? filter.page;
  const pageSize = rawPagination?.page_size ?? filter.page_size;
  const total = rawPagination?.total ?? result.data.length;
  const totalPages =
    rawPagination?.total_pages ?? Math.max(1, Math.ceil(total / pageSize));

  return {
    users: result.data,
    pagination: {
      page,
      page_size: pageSize,
      total,
      total_pages: totalPages,
    },
    request_id: parsedMeta.success ? parsedMeta.data.request_id : undefined,
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

export async function updateUserProfile(
  id: string,
  input: UpdateOperationalProfileInput,
): Promise<AdminUser> {
  const parsedId = z.string().uuid().parse(id);
  const body = updateOperationalProfileSchema.parse(input);
  return browserRequest<AdminUser>(`/api/v1/admin/users/${parsedId}`, {
    method: "PATCH",
    json: body,
    schema: adminUserSchema,
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

export async function revokeUserSessions(id: string): Promise<void> {
  const parsedId = z.string().uuid().parse(id);
  await browserRequestVoid(`/api/v1/admin/users/${parsedId}/revoke-sessions`, {
    method: "POST",
  });
}
