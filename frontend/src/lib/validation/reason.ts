import { z } from "zod";

/**
 * A reason customers read — why the bakery is closed, why it cancelled an
 * order: 1 to 200 characters of plain text, as the backend checks it. No
 * control or invisible formatting characters that would make it read
 * differently from what was typed.
 */
export function reasonSchema(requiredMessage: string) {
  return z
    .string()
    .trim()
    .min(1, requiredMessage)
    .max(200, "At most 200 characters")
    .refine((value) => !/[\p{Cc}\p{Cf}]/u.test(value), "Use plain text only");
}
