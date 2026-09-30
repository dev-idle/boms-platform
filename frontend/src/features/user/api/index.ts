import { browserRequest, browserRequestVoid } from "@/lib/browser-api-client";

import { meSchema, type Me } from "@/lib/schemas/me";

import {
  changePasswordSchema,
  dataExportSchema,
  eraseMyAccountSchema,
  updateSelfProfileSchema,
  type ChangePasswordInput,
  type DataExport,
  type EraseMyAccountInput,
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

export async function deleteAccount(body: EraseMyAccountInput): Promise<void> {
  const parsed = eraseMyAccountSchema.parse(body);
  await browserRequestVoid("/api/v1/me", {
    method: "DELETE",
    json: parsed,
  });
}

/** Asks for a new link confirming the signed-in user's address; the last one stops working. */
export async function resendVerificationEmail(): Promise<void> {
  await browserRequestVoid("/api/v1/me/email-verification", {
    method: "POST",
  });
}

export async function exportMyData(): Promise<DataExport> {
  return browserRequest<DataExport>("/api/v1/me/export", {
    method: "GET",
    schema: dataExportSchema,
  });
}
