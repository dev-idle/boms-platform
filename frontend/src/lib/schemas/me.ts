import { z } from "zod";

import { USER_ROLE } from "@/constants/roles";

const optionalNullableStringSchema = z.string().optional().nullable();

/** GET /api/v1/me profile shapes (shared contract; used by auth store + user feature). */
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

export type Me = z.infer<typeof meSchema>;
export type CustomerProfile = z.infer<typeof customerProfileSchema>;
export type StaffProfile = z.infer<typeof staffProfileSchema>;
export type AdminProfile = z.infer<typeof adminProfileSchema>;
