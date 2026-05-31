import { z } from "zod";

export const apiErrorBodySchema = z.object({
  code: z.string(),
  message: z.string(),
  details: z.record(z.string()).optional(),
});

export const apiEnvelopeSchema = z.object({
  success: z.boolean(),
  data: z.unknown().optional(),
  error: apiErrorBodySchema.optional(),
  meta: z
    .object({
      request_id: z.string().optional(),
      trace_id: z.string().optional(),
    })
    .passthrough()
    .optional(),
});

export type ApiErrorBody = z.infer<typeof apiErrorBodySchema>;
export type ApiEnvelope = z.infer<typeof apiEnvelopeSchema>;

/** Safe JSON body parse — empty or invalid body returns `null` (no throw). */
export async function parseResponseBody(response: Response): Promise<unknown> {
  const text = await response.text();
  if (!text) {
    return null;
  }
  try {
    return JSON.parse(text) as unknown;
  } catch {
    return null;
  }
}

export function parseApiEnvelope(
  payload: unknown,
): z.SafeParseReturnType<unknown, ApiEnvelope> {
  return apiEnvelopeSchema.safeParse(payload);
}
