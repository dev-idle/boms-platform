import { browserRequest, browserRequestVoid } from "@/lib/browser-api-client";

import { meSchema, type Me } from "@/lib/schemas/me";

import {
  changePasswordSchema,
  updateSelfProfileSchema,
  type ChangePasswordInput,
  type UpdateSelfProfileInput,
} from "../schemas/index";

export async function getMe(): Promise<Me> {
  return browserRequest<Me>("/api/v1/me", {
    method: "GET",
    schema: meSchema,
  });
}

export async function updateProfile(body: UpdateSelfProfileInput): Promise<Me> {
  const parsed = updateSelfProfileSchema.parse(body);
  return browserRequest<Me>("/api/v1/me", {
    method: "PATCH",
    json: parsed,
    schema: meSchema,
  });
}

export async function changePassword(body: ChangePasswordInput): Promise<void> {
  const parsed = changePasswordSchema.parse(body);
  await browserRequestVoid("/api/v1/me/password", {
    method: "PATCH",
    json: parsed,
  });
}

export async function deleteAccount(): Promise<void> {
  await browserRequestVoid("/api/v1/me", {
    method: "DELETE",
  });
}
