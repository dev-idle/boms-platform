import { z } from "zod";

import { USER_ROLE } from "@/constants/roles";
import { newPasswordZodString } from "@/lib/validation/password";

const userRoleSchema = z.enum([
  USER_ROLE.customer,
  USER_ROLE.staff,
  USER_ROLE.baker,
  USER_ROLE.manager,
  USER_ROLE.admin,
]);

export const userSchema = z.object({
  id: z.uuid(),
  email: z.email(),
  role: userRoleSchema,
  email_verified: z.boolean(),
  created_at: z.string(),
});

export const tokenResponseSchema = z.object({
  access_token: z.string().min(1),
  token_type: z.string(),
  expires_in: z.number().int().positive(),
  user: userSchema,
  must_change_password: z.boolean().optional(),
});

export const registerSchema = z.object({
  email: z
    .string()
    .trim()
    .pipe(
      z
        .email("Enter a valid email address")
        .max(255, "Email must be at most 255 characters"),
    ),
  password: newPasswordZodString(),
  accept_terms: z
    .boolean()
    .refine((accepted) => accepted, "Accept the terms and the privacy policy to create an account"),
});

export const forgotPasswordSchema = z.object({
  email: z
    .string()
    .trim()
    .pipe(
      z
        .email("Enter a valid email address")
        .max(255, "Email must be at most 255 characters"),
    ),
});

/** The token an emailed link carries: 256 random bits, base64url. */
export const linkTokenSchema = z.string().regex(/^[A-Za-z0-9_-]{16,128}$/, "This link is not valid");

export const verifyEmailSchema = z.object({ token: linkTokenSchema });

/** POST /auth/password-reset/confirm. */
export const passwordResetConfirmSchema = z.object({
  token: linkTokenSchema,
  new_password: newPasswordZodString(),
});

/** The reset form: the new password twice. */
export const resetPasswordFormSchema = z
  .object({
    new_password: newPasswordZodString(),
    confirm_password: z.string().min(1, "Confirm your new password"),
  })
  .superRefine((input, ctx) => {
    if (input.new_password !== input.confirm_password) {
      ctx.addIssue({ code: "custom", message: "Passwords do not match", path: ["confirm_password"] });
    }
  });

export const loginSchema = z.object({
  email: z.string().trim().pipe(z.email("Enter a valid email address")),
  password: z.string().min(1, "Password is required"),
});

export type User = z.infer<typeof userSchema>;
export type TokenResponse = z.infer<typeof tokenResponseSchema>;
export type RegisterInput = z.infer<typeof registerSchema>;
export type LoginInput = z.infer<typeof loginSchema>;
export type ForgotPasswordInput = z.infer<typeof forgotPasswordSchema>;
export type PasswordResetConfirmInput = z.infer<typeof passwordResetConfirmSchema>;
export type ResetPasswordFormInput = z.infer<typeof resetPasswordFormSchema>;
