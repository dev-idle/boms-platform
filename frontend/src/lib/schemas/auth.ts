import { z } from "zod";

/** POST /auth/refresh response body (used by lib/auth; not feature-owned). */
export const refreshResponseSchema = z.object({
  access_token: z.string().min(1),
  token_type: z.string(),
  expires_in: z.number().int().positive(),
  must_change_password: z.boolean().optional(),
});

export type RefreshResponse = z.infer<typeof refreshResponseSchema>;
