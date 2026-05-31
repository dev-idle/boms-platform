import { z } from "zod";

import { ApiError, ApiErrorCode, isApiError } from "@/lib/errors";
import { apiEnvelopeSchema, parseResponseBody } from "@/lib/api-envelope";

export async function readApiEnvelope(response: Response): Promise<{
  status: number;
  envelope: z.infer<typeof apiEnvelopeSchema>;
  raw: unknown;
}> {
  const raw = await parseResponseBody(response);
  const parsed = apiEnvelopeSchema.safeParse(raw);

  if (!parsed.success) {
    throw new ApiError(response.status, {
      code: ApiErrorCode.InvalidResponse,
      message: "Response failed envelope validation",
    });
  }

  return { status: response.status, envelope: parsed.data, raw };
}

export function throwApiErrorFromEnvelope(
  status: number,
  envelope: z.infer<typeof apiEnvelopeSchema>,
): never {
  if (envelope.error) {
    throw new ApiError(status, envelope.error);
  }
  throw new ApiError(status, {
    code: ApiErrorCode.Unknown,
    message: `Request failed with HTTP ${status}`,
  });
}

export function isAuthSessionError(error: unknown): boolean {
  return isApiError(error) && error.isAuthError();
}

/** Clears HttpOnly refresh cookie via hybrid logout (idempotent). */
export async function clearStaleSession(): Promise<void> {
  try {
    await fetch("/api/v1/auth/logout", {
      method: "POST",
      credentials: "include",
      cache: "no-store",
    });
  } catch {
    // Best-effort; cookie may already be cleared by backend.
  }
}
