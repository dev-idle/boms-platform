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
