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
  id: z.string().uuid(),
  email: z.string().email(),
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
  email: z.string().trim().email("Enter a valid email address").max(255),
  password: newPasswordZodString(),
});

export const loginSchema = z.object({
  email: z.string().trim().email("Enter a valid email address"),
  password: z.string().min(1, "Password is required"),
});

export type User = z.infer<typeof userSchema>;
export type TokenResponse = z.infer<typeof tokenResponseSchema>;
export type RegisterInput = z.infer<typeof registerSchema>;
export type LoginInput = z.infer<typeof loginSchema>;
