import { z } from "zod";

import { USER_ROLE } from "@/constants/roles";

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
  phone: optionalNullableTrimmedString(50, "Phone"),
  employee_code: optionalNullableTrimmedString(64, "Employee code"),
});

const createOperationalBaseSchema = z.object({
  email: z.string().trim().email("Enter a valid email").max(255),
  full_name: z.string().trim().min(1, "Full name is required").max(255),
  phone: optionalNullableTrimmedString(50, "Phone"),
});

/** Admin API only accepts staff, baker, manager — never admin (use dev seed). */
export const createOperationalSchema = createOperationalBaseSchema.extend({
  role: z.enum([
    USER_ROLE.staff,
    USER_ROLE.baker,
    USER_ROLE.manager,
  ]),
  employee_code: z
    .string()
    .trim()
    .min(1, "Employee code is required")
    .max(64),
});

export const createOperationalResponseSchema = z.object({
  user: adminUserSchema,
  temp_password: z.string().min(1),
});

export const updateOperationalProfileSchema = z.object({
  full_name: z.string().trim().min(1, "Full name is required").max(255),
  phone: optionalNullableTrimmedString(50, "Phone"),
  employee_code: z.string().trim().max(64).optional().nullable(),
});

/** Admin user detail profile tab — string fields for controlled inputs. */
export const adminUserProfileFormSchema = z.object({
  full_name: z.string().trim().min(1, "Full name is required").max(255),
  phone: z.string().trim().max(50).optional(),
  employee_code: z.string().trim().max(64).optional(),
});

const updateRoleBaseSchema = z.object({
  role: z.enum([
    USER_ROLE.staff,
    USER_ROLE.baker,
    USER_ROLE.manager,
  ]),
  full_name: z.string().trim().max(255).optional(),
  phone: z.string().trim().max(50).optional().nullable(),
  employee_code: z.string().trim().max(64).optional().nullable(),
});

export const updateRoleSchema = updateRoleBaseSchema.superRefine((input, ctx) => {
  const isStaffLike =
    input.role === USER_ROLE.staff ||
    input.role === USER_ROLE.baker ||
    input.role === USER_ROLE.manager;

  if (isStaffLike) {
    if (!input.employee_code?.trim()) {
      ctx.addIssue({
        code: z.ZodIssueCode.custom,
        message: "Employee code is required for staff roles",
        path: ["employee_code"],
      });
    }
  }
});

export const listFilterSchema = z.object({
  page: z.coerce.number().int().min(1).default(1),
  page_size: z.coerce.number().int().min(1).max(100).default(20),
  search: z.string().trim().default(""),
});

export const paginationMetaSchema = z.object({
  page: z.number().int().min(1),
  page_size: z.number().int().min(1),
  total: z.number().int().min(0),
  total_pages: z.number().int().min(1),
});

export const usersListResultSchema = z.object({
  users: z.array(adminUserSchema),
  pagination: paginationMetaSchema,
  request_id: z.string().optional(),
});

export type AdminUser = z.infer<typeof adminUserSchema>;
export type CreateOperationalInput = z.infer<typeof createOperationalSchema>;
export type CreateOperationalResponse = z.infer<
  typeof createOperationalResponseSchema
>;
export type UpdateOperationalProfileInput = z.infer<
  typeof updateOperationalProfileSchema
>;
export type AdminUserProfileFormValues = z.infer<typeof adminUserProfileFormSchema>;
export type UpdateRoleInput = z.infer<typeof updateRoleSchema>;
export type ListFilterInput = z.infer<typeof listFilterSchema>;
export type UsersListResult = z.infer<typeof usersListResultSchema>;
