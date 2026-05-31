import { z } from "zod";

import { newPasswordZodString } from "@/lib/validation/password";

import { USER_ROLE } from "@/constants/roles";

const dateOnlySchema = z.string().regex(/^\d{4}-\d{2}-\d{2}$/, "Use YYYY-MM-DD");

const optionalNullableStringSchema = z.string().optional().nullable();
const optionalNullablePhoneSchema = z
  .string()
  .trim()
  .max(50, "Phone must be at most 50 characters")
  .optional()
  .nullable();

export const customerProfileSchema = z.object({
  type: z.literal("customer"),
  display_name: optionalNullableStringSchema,
  phone: optionalNullableStringSchema,
});

export const staffProfileSchema = z.object({
  type: z.literal("staff"),
  full_name: z.string(),
  phone: optionalNullableStringSchema,
  employee_code: z.string(),
  hire_date: dateOnlySchema,
  shift: z.string(),
});

export const adminProfileSchema = z.object({
  type: z.literal("admin"),
  full_name: z.string(),
  phone: optionalNullableStringSchema,
});

const meBaseSchema = z.object({
  id: z.string().uuid(),
  email: z.string().email(),
  email_verified: z.boolean(),
  must_change_password: z.boolean(),
  disabled: z.boolean(),
  created_at: z.string(),
});

const customerMeSchema = meBaseSchema.extend({
  role: z.literal(USER_ROLE.customer),
  profile: customerProfileSchema,
});

const staffMeSchema = meBaseSchema.extend({
  role: z.literal(USER_ROLE.staff),
  profile: staffProfileSchema,
});

const bakerMeSchema = meBaseSchema.extend({
  role: z.literal(USER_ROLE.baker),
  profile: staffProfileSchema,
});

const managerMeSchema = meBaseSchema.extend({
  role: z.literal(USER_ROLE.manager),
  profile: staffProfileSchema,
});

const adminMeSchema = meBaseSchema.extend({
  role: z.literal(USER_ROLE.admin),
  profile: adminProfileSchema,
});

export const meSchema = z.discriminatedUnion("role", [
  customerMeSchema,
  staffMeSchema,
  bakerMeSchema,
  managerMeSchema,
  adminMeSchema,
]);

export const updateSelfProfileSchema = z
  .object({
    display_name: z
      .string()
      .trim()
      .max(255, "Display name must be at most 255 characters")
      .optional(),
    phone: optionalNullablePhoneSchema,
    full_name: z
      .string()
      .trim()
      .max(255, "Full name must be at most 255 characters")
      .optional(),
  })
  .strict();

export const changePasswordSchema = z.object({
  old_password: z.string().min(1, "Current password is required"),
  new_password: newPasswordZodString(),
});

export type UpdateSelfProfileInput = z.infer<typeof updateSelfProfileSchema>;
export type ChangePasswordInput = z.infer<typeof changePasswordSchema>;
