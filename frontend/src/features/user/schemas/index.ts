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

const selfProfilePhoneFormField = z
  .string()
  .trim()
  .max(50, "Phone must be at most 50 characters")
  .optional();

/** Admin + operational self-service profile forms (PATCH /me: full_name, phone). */
export const fullNamePhoneSelfProfileFormSchema = z.object({
  full_name: z.string().trim().min(1, "Full name is required").max(255),
  phone: selfProfilePhoneFormField,
});

export const customerSelfProfileFormSchema = z.object({
  display_name: z
    .string()
    .trim()
    .max(255, "Display name must be at most 255 characters")
    .optional(),
  phone: selfProfilePhoneFormField,
});

export const changePasswordSchema = z.object({
  old_password: z.string().min(1, "Current password is required"),
  new_password: newPasswordZodString(),
});

/** API payload + confirm field for the change-password form. */
export const changePasswordFormSchema = changePasswordSchema
  .extend({
    confirm_password: z.string().min(1, "Confirm your new password"),
  })
  .superRefine((input, ctx) => {
    if (input.new_password !== input.confirm_password) {
      ctx.addIssue({
        code: z.ZodIssueCode.custom,
        message: "Passwords do not match",
        path: ["confirm_password"],
      });
    }
  });

export type UpdateSelfProfileInput = z.infer<typeof updateSelfProfileSchema>;
export type FullNamePhoneSelfProfileFormValues = z.infer<
  typeof fullNamePhoneSelfProfileFormSchema
>;
export type CustomerSelfProfileFormValues = z.infer<
  typeof customerSelfProfileFormSchema
>;
export type ChangePasswordInput = z.infer<typeof changePasswordSchema>;
export type ChangePasswordFormInput = z.infer<typeof changePasswordFormSchema>;
