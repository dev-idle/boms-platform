import { z } from "zod";

import { ApiError } from "@/lib/errors";
import { apiEnvelopeSchema } from "@/lib/api-envelope";

export async function readApiEnvelope(response: Response): Promise<{
  status: number;
  envelope: z.infer<typeof apiEnvelopeSchema>;
  raw: unknown;
}> {
  const raw = await response.json().catch(() => null);
  const parsed = apiEnvelopeSchema.safeParse(raw);

  if (!parsed.success) {
    throw new ApiError(response.status, {
      code: "INVALID_RESPONSE",
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
    code: "UNKNOWN_ERROR",
    message: `Request failed with HTTP ${status}`,
  });
}

export function isAuthSessionError(error: unknown): boolean {
  if (!(error instanceof ApiError)) {
    return false;
  }
  if (error.status === 401) {
    return true;
  }
  return (
    error.code === "INVALID_REFRESH_TOKEN" ||
    error.code === "SESSION_REVOKED" ||
    error.code === "MISSING_REFRESH_TOKEN"
  );
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
