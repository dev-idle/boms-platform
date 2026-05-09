import "server-only";

import { z } from "zod";

import { getBomsApiClient } from "@/lib/api-client";

const healthSchema = z.object({
  status: z.string().optional(),
  ok: z.boolean().optional(),
});

/**
 * Example DAL call. Replace with real domain modules under this directory.
 */
export async function dalFiberHealth(): Promise<z.infer<typeof healthSchema>> {
  return getBomsApiClient().request("/health", {
    method: "GET",
    schema: healthSchema,
  });
}
