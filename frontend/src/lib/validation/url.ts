import { z } from "zod";

/** HTTP(S) URLs only — blocks javascript:, data:, and other schemes at parse time. */
export const optionalHttpUrlSchema = z
  .string()
  .trim()
  .max(2048)
  .url("Enter a valid URL")
  .refine(
    (value) => value.startsWith("http://") || value.startsWith("https://"),
    "URL must use HTTP or HTTPS",
  )
  .optional()
  .nullable();
