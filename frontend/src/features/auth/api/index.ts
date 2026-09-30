import { TERMS_VERSION } from "@/constants/policies";
import {
  browserRequest,
  browserRequestVoid,
} from "@/lib/browser-api-client";

import {
  forgotPasswordSchema,
  loginSchema,
  passwordResetConfirmSchema,
  registerSchema,
  tokenResponseSchema,
  userSchema,
  verifyEmailSchema,
  type ForgotPasswordInput,
  type LoginInput,
  type PasswordResetConfirmInput,
  type RegisterInput,
  type TokenResponse,
  type User,
} from "../schemas";

/** Signs up with the policies the form showed: the API records their version. */
export async function register(body: RegisterInput): Promise<User> {
  const { email, password } = registerSchema.parse(body);
  return browserRequest<User>("/api/v1/auth/register", {
    method: "POST",
    json: { email, password, terms_version: TERMS_VERSION },
    schema: userSchema,
    skipAuth: true,
    skipRefreshRetry: true,
  });
}

export async function login(body: LoginInput): Promise<TokenResponse> {
  const parsed = loginSchema.parse(body);
  return browserRequest<TokenResponse>("/api/v1/auth/login", {
    method: "POST",
    json: parsed,
    schema: tokenResponseSchema,
    skipAuth: true,
    skipRefreshRetry: true,
  });
}

export async function logout(): Promise<void> {
  await browserRequestVoid("/api/v1/auth/logout", {
    method: "POST",
    skipRefreshRetry: true,
  });
}

/** Answers 202 whether or not the address has an account. */
export async function requestPasswordReset(body: ForgotPasswordInput): Promise<void> {
  const parsed = forgotPasswordSchema.parse(body);
  await browserRequestVoid("/api/v1/auth/password-reset/request", {
    method: "POST",
    json: parsed,
    skipAuth: true,
    skipRefreshRetry: true,
  });
}

export async function confirmPasswordReset(body: PasswordResetConfirmInput): Promise<void> {
  const parsed = passwordResetConfirmSchema.parse(body);
  await browserRequestVoid("/api/v1/auth/password-reset/confirm", {
    method: "POST",
    json: parsed,
    skipAuth: true,
    skipRefreshRetry: true,
  });
}

/** Works without a session: the link is often opened on another device. */
export async function verifyEmail(token: string): Promise<void> {
  const parsed = verifyEmailSchema.parse({ token });
  await browserRequestVoid("/api/v1/auth/verify-email", {
    method: "POST",
    json: parsed,
    skipAuth: true,
    skipRefreshRetry: true,
  });
}
