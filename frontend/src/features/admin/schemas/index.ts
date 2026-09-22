import { z } from "zod";

import { USER_ROLE } from "@/constants/roles";
import { vietnamPhoneZodString } from "@/lib/validation/phone";
import type { PaginatedListResult } from "@/lib/pagination/parse-paginated-list";

const optionalNullableTrimmedString = (max: number, label: string) =>
  z
    .string()
    .trim()
    .max(max, `${label} must be at most ${max} characters`)
    .optional()
    .nullable();

export const adminUserSchema = z.object({
  id: z.string().uuid(),
  email: z.string().email(),
  role: z.enum([
    USER_ROLE.customer,
    USER_ROLE.staff,
    USER_ROLE.baker,
    USER_ROLE.manager,
    USER_ROLE.admin,
  ]),
  email_verified: z.boolean(),
  must_change_password: z.boolean(),
  disabled: z.boolean(),
  created_at: z.string(),
  updated_at: z.string(),
  display_name: optionalNullableTrimmedString(255, "Display name"),
  full_name: optionalNullableTrimmedString(255, "Full name"),
  /* Read side stays permissive: rows written before the phone rule existed still have to render. */
  phone: optionalNullableTrimmedString(50, "Phone"),
  employee_code: optionalNullableTrimmedString(64, "Employee code"),
});

const createOperationalBaseSchema = z.object({
  email: z.string().trim().email("Enter a valid email").max(255),
  full_name: z.string().trim().min(1, "Full name is required").max(255),
  phone: vietnamPhoneZodString(),
});

/** Admin API only accepts staff, baker, manager — never admin (use dev seed). */
export const createOperationalSchema = createOperationalBaseSchema.extend({
  role: z.enum([
    USER_ROLE.staff,
    USER_ROLE.baker,
    USER_ROLE.manager,
  ]),
});

export const nextEmployeeCodeSchema = z.object({
  employee_code: z.string().min(1),
});

export const createOperationalResponseSchema = z.object({
  user: adminUserSchema,
  temp_password: z.string().min(1),
});

export const adminResetPasswordResponseSchema = z.object({
  user: adminUserSchema,
  temp_password: z.string().min(1),
});

const updateRoleBaseSchema = z.object({
  role: z.enum([
    USER_ROLE.staff,
    USER_ROLE.baker,
    USER_ROLE.manager,
  ]),
  full_name: z.string().trim().max(255).optional(),
  phone: vietnamPhoneZodString(),
});

export const updateRoleSchema = updateRoleBaseSchema;

const adminUserRoleFilterSchema = z.enum([
  USER_ROLE.customer,
  USER_ROLE.staff,
  USER_ROLE.baker,
  USER_ROLE.manager,
  USER_ROLE.admin,
]);

export const listFilterSchema = z.object({
  page: z.coerce.number().int().min(1).default(1),
  page_size: z.coerce.number().int().min(1).max(100).default(20),
  search: z.string().trim().default(""),
  role: adminUserRoleFilterSchema.optional(),
});

export const adminUserActivityLogSchema = z.object({
  id: z.string().uuid(),
  action: z.string().min(1),
  summary: z.string().min(1),
  actor_id: z.string().uuid(),
  actor_email: z.string().email(),
  actor_role: z.enum([
    USER_ROLE.customer,
    USER_ROLE.staff,
    USER_ROLE.baker,
    USER_ROLE.manager,
    USER_ROLE.admin,
  ]),
  created_at: z.string(),
});

export const userActivityFilterSchema = z.object({
  page: z.coerce.number().int().min(1).default(1),
  page_size: z.coerce.number().int().min(1).max(100).default(20),
});

export type AdminUser = z.infer<typeof adminUserSchema>;
export type CreateOperationalInput = z.infer<typeof createOperationalSchema>;
export type NextEmployeeCode = z.infer<typeof nextEmployeeCodeSchema>;
export type CreateOperationalResponse = z.infer<
  typeof createOperationalResponseSchema
>;
export type AdminResetPasswordResponse = z.infer<
  typeof adminResetPasswordResponseSchema
>;
export type AdminTempPasswordPayload = CreateOperationalResponse | AdminResetPasswordResponse;
export type UpdateRoleInput = z.infer<typeof updateRoleSchema>;
export type AdminUserRoleFilter = z.infer<typeof adminUserRoleFilterSchema>;
export type ListFilterInput = z.infer<typeof listFilterSchema>;
export type UsersListResult = {
  users: AdminUser[];
  pagination: PaginatedListResult<AdminUser>["pagination"];
  request_id?: string;
};
export type AdminUserActivityLog = z.infer<typeof adminUserActivityLogSchema>;
export type UserActivityFilterInput = z.infer<typeof userActivityFilterSchema>;
export type UserActivityListResult = {
  entries: AdminUserActivityLog[];
  pagination: PaginatedListResult<AdminUserActivityLog>["pagination"];
  request_id?: string;
};
