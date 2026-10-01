import { z } from "zod";

import { apiDateTimeSchema } from "@/lib/validation/datetime";

/**
 * The most characters one message holds — mirrors backend
 * `conversation.MaxBodyLength`, which counts characters (code points), not the
 * UTF-16 units `String.length` counts: an emoji is one.
 */
export const MESSAGE_MAX_LENGTH = 2000;

/**
 * One message about an order. `author_name`, who at the counter wrote it,
 * reaches staff only; customers read the counter's messages as the bakery's.
 * The API holds the length limit; reading a message never refuses it.
 */
export const messageSchema = z.object({
  id: z.uuid(),
  from: z.enum(["customer", "staff"]),
  author_name: z.string().nullable(),
  body: z.string().min(1),
  created_at: apiDateTimeSchema,
});

export type Message = z.infer<typeof messageSchema>;

/** A message as written; the API trims it and checks its characters again. */
export const messageInputSchema = z.object({
  body: z
    .string()
    .trim()
    .min(1, "Write a message first")
    .refine(
      (body) => [...body].length <= MESSAGE_MAX_LENGTH,
      `A message holds up to ${MESSAGE_MAX_LENGTH} characters`,
    ),
});

export type MessageInput = z.infer<typeof messageInputSchema>;

/** Why the API refused a message the composer let through: a character plain text does not hold. */
export const MESSAGE_REFUSED_TEXT = "Messages take plain text: remove tabs or hidden characters, then send again.";
