import {
  browserRequest,
  browserRequestVoid,
} from "@/lib/browser-api-client";

import {
  loginSchema,
  registerSchema,
  tokenResponseSchema,
  userSchema,
  type LoginInput,
  type RegisterInput,
  type TokenResponse,
  type User,
} from "../schemas";

export async function register(body: RegisterInput): Promise<User> {
  const parsed = registerSchema.parse(body);
  return browserRequest<User>("/api/v1/auth/register", {
    method: "POST",
    json: parsed,
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

export async function getMe(): Promise<User> {
  return browserRequest<User>("/api/v1/auth/me", {
    method: "GET",
    schema: userSchema,
  });
}

export async function logout(): Promise<void> {
  await browserRequestVoid("/api/v1/auth/logout", {
    method: "POST",
    skipRefreshRetry: true,
  });
}
