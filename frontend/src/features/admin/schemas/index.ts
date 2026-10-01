import { z } from "zod";

import { USER_ROLE } from "@/constants/roles";
import { clockTimeSchema, clockToMinutes } from "@/lib/validation/clock";
import { apiDateTimeSchema } from "@/lib/validation/datetime";
import { reasonSchema } from "@/lib/validation/reason";
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
  id: z.uuid(),
  email: z.email(),
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
  erased: z.boolean(),
  created_at: z.string(),
  updated_at: z.string(),
  display_name: optionalNullableTrimmedString(255, "Display name"),
  full_name: optionalNullableTrimmedString(255, "Full name"),
  /* Read side stays permissive: rows written before the phone rule existed still have to render. */
  phone: optionalNullableTrimmedString(50, "Phone"),
  employee_code: optionalNullableTrimmedString(64, "Employee code"),
});

const createOperationalBaseSchema = z.object({
  email: z
    .string()
    .trim()
    .pipe(z.email("Enter a valid email").max(255, "Email must be at most 255 characters")),
  full_name: z
    .string()
    .trim()
    .min(1, "Full name is required")
    .max(255, "Full name must be at most 255 characters"),
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
  // Required here as it is on create and on the person's own profile: the API
  // refuses an empty operational name rather than dropping it silently.
  full_name: z
    .string()
    .trim()
    .min(1, "Full name is required")
    .max(255, "Full name must be at most 255 characters"),
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
  id: z.uuid(),
  action: z.string().min(1),
  summary: z.string().min(1),
  actor_id: z.uuid(),
  actor_email: z.email(),
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

/** GET/PATCH /api/v1/admin/settings — the pickup rules an admin edits. */
export const storeSettingsSchema = z.object({
  opens_at: clockTimeSchema,
  closes_at: clockTimeSchema,
  preorder_min_lead_minutes: z.number().int().min(0),
  max_advance_days: z.number().int().min(1),
  slot_minutes: z.number().int().min(1),
  slot_capacity: z.number().int().min(1),
  instant_prep_minutes: z.number().int().min(0),
  payment_hold_minutes: z.number().int().min(1),
  updated_at: apiDateTimeSchema,
});

/**
 * Bounds of backend `store.Settings`: at most seven days' notice, 1 to 90 days
 * ahead, slots that divide an hour, 1 to 200 orders a slot, at most four
 * hours to pack an instant order, and 5 to 120 minutes to pay.
 */
export const MAX_PREORDER_LEAD_MINUTES = 7 * 24 * 60;
export const MIN_ADVANCE_DAYS = 1;
export const MAX_ADVANCE_DAYS = 90;
export const SLOT_MINUTE_OPTIONS = [10, 15, 20, 30, 60] as const;
export const MAX_SLOT_CAPACITY = 200;
export const MAX_INSTANT_PREP_MINUTES = 4 * 60;
export const MIN_PAYMENT_HOLD_MINUTES = 5;
export const MAX_PAYMENT_HOLD_MINUTES = 120;

const storeSettingsFieldsSchema = z.object({
  opens_at: clockTimeSchema,
  closes_at: clockTimeSchema,
  preorder_min_lead_minutes: z
    .number({ error: "Enter the notice in minutes" })
    .int()
    .min(0)
    .max(MAX_PREORDER_LEAD_MINUTES, `At most ${MAX_PREORDER_LEAD_MINUTES} minutes (7 days)`),
  max_advance_days: z
    .number({ error: "Enter the number of days" })
    .int()
    .min(MIN_ADVANCE_DAYS, `At least ${MIN_ADVANCE_DAYS} day`)
    .max(MAX_ADVANCE_DAYS, `At most ${MAX_ADVANCE_DAYS} days`),
  slot_minutes: z
    .number({ error: "Choose a slot length" })
    .refine((minutes) => SLOT_MINUTE_OPTIONS.some((option) => option === minutes), "Choose a slot length"),
  slot_capacity: z
    .number({ error: "Enter the number of orders" })
    .int()
    .min(1, "At least 1 order")
    .max(MAX_SLOT_CAPACITY, `At most ${MAX_SLOT_CAPACITY} orders`),
  instant_prep_minutes: z
    .number({ error: "Enter the time in minutes" })
    .int()
    .min(0)
    .max(MAX_INSTANT_PREP_MINUTES, `At most ${MAX_INSTANT_PREP_MINUTES} minutes (4 hours)`),
  payment_hold_minutes: z
    .number({ error: "Enter the time in minutes" })
    .int()
    .min(MIN_PAYMENT_HOLD_MINUTES, `At least ${MIN_PAYMENT_HOLD_MINUTES} minutes`)
    .max(MAX_PAYMENT_HOLD_MINUTES, `At most ${MAX_PAYMENT_HOLD_MINUTES} minutes (2 hours)`),
});

/**
 * PATCH /api/v1/admin/settings — only the fields the admin changed, so an edit
 * never overwrites a setting someone else saved meanwhile. The server checks
 * the merged result.
 */
export const storeSettingsPatchSchema = storeSettingsFieldsSchema.partial();

/** The settings form; mirrors backend `store.Settings.Validate`. */
export const storeSettingsFormSchema = storeSettingsFieldsSchema
  .refine((v) => clockToMinutes(v.opens_at) < clockToMinutes(v.closes_at), {
    path: ["closes_at"],
    message: "Closing time must be after opening time",
  })
  .refine((v) => v.preorder_min_lead_minutes < v.max_advance_days * 24 * 60, {
    path: ["preorder_min_lead_minutes"],
    message: "Notice must be shorter than the booking window",
  })
  // Only once the hours are in order: an inverted pair is one mistake, already named.
  .refine((v) => clockToMinutes(v.closes_at) <= clockToMinutes(v.opens_at) ||
    clockToMinutes(v.closes_at) - clockToMinutes(v.opens_at) >= v.slot_minutes, {
    path: ["slot_minutes"],
    message: "The opening hours must hold at least one slot",
  });

/** GET /api/v1/admin/closed-dates — a day the bakery takes no pickups. */
export const closedDateSchema = z.object({
  id: z.uuid(),
  date: z.iso.date(),
  reason: z.string().min(1),
  created_at: apiDateTimeSchema,
});

export const closedDateFormSchema = z.object({
  date: z.iso.date("Choose a day"),
  reason: reasonSchema("Tell customers why the bakery is closed"),
});

export type StoreSettings = z.infer<typeof storeSettingsSchema>;
export type StoreSettingsFormInput = z.infer<typeof storeSettingsFormSchema>;
export type StoreSettingsPatch = z.infer<typeof storeSettingsPatchSchema>;
export type ClosedDate = z.infer<typeof closedDateSchema>;
export type ClosedDateFormInput = z.infer<typeof closedDateFormSchema>;
