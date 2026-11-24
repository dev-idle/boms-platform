import { z } from "zod";

import { messageInputSchemaOf } from "@/lib/schemas/message";

/** The most characters one question holds — mirrors backend `assistant.MaxQuestionLength`. */
export const ASSISTANT_QUESTION_MAX_LENGTH = 500;

/** A question as written; the API trims it and checks its characters again. */
export const assistantQuestionSchema = messageInputSchemaOf(ASSISTANT_QUESTION_MAX_LENGTH);

/** The first event of an answer: the session to continue with the next question. */
export const assistantSessionEventSchema = z.object({ session_id: z.uuid() });

/** Text of the answer as the model writes it. */
export const assistantDeltaEventSchema = z.object({ text: z.string() });

/**
 * The last event of an answer: the answer as kept, which replaces the text
 * streamed before it. `offer_staff` suggests talking to staff instead;
 * `hand_off` says the question went to the counter — on the order it names,
 * or in the customer's general conversation when the order is null.
 */
export const assistantAnswerSchema = z.object({
  reply: z.string().min(1),
  offer_staff: z.boolean(),
  hand_off: z
    .object({
      order_id: z.uuid().nullable(),
      order_code: z.string().nullable(),
    })
    .nullable(),
});

/** Ends an answer that could not be given. */
export const assistantErrorEventSchema = z.object({
  code: z.string(),
  message: z.string(),
});

export type AssistantAnswer = z.infer<typeof assistantAnswerSchema>;
