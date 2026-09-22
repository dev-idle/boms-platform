import { z } from "zod";

import { newPasswordZodString } from "@/lib/validation/password";
import { vietnamPhoneZodString } from "@/lib/validation/phone";

export const updateSelfProfileSchema = z
  .object({
    display_name: z
      .string()
      .trim()
      .max(255, "Display name must be at most 255 characters")
      .optional(),
    phone: vietnamPhoneZodString(),
    full_name: z
      .string()
      .trim()
      .max(255, "Full name must be at most 255 characters")
      .optional(),
  })
  .strict();

/** Admin + operational self-service profile forms (PATCH /me: full_name, phone). */
export const fullNamePhoneSelfProfileFormSchema = z.object({
  full_name: z.string().trim().min(1, "Full name is required").max(255),
  phone: vietnamPhoneZodString(),
});

export const customerSelfProfileFormSchema = z.object({
  display_name: z
    .string()
    .trim()
    .max(255, "Display name must be at most 255 characters")
    .optional(),
  phone: vietnamPhoneZodString(),
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
