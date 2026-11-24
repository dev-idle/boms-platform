import type { ZodType } from "zod";

import { browserEventStream } from "@/lib/browser-api-client";
import { ApiError, ApiErrorCode } from "@/lib/errors";
import type { MessageInput } from "@/lib/schemas/message";
import { readServerSentEvents } from "@/lib/sse";

import {
  assistantAnswerSchema,
  assistantDeltaEventSchema,
  assistantErrorEventSchema,
  assistantQuestionSchema,
  assistantSessionEventSchema,
  type AssistantAnswer,
} from "../schemas";

type AskHandlers = {
  /** The session the answer belongs to, sent before its text. */
  onSession: (sessionId: string) => void;
  /** Text of the answer as the model writes it. */
  onText: (text: string) => void;
};

function readEvent<T>(schema: ZodType<T>, data: string): T {
  let parsed: unknown;
  try {
    parsed = JSON.parse(data);
  } catch {
    parsed = undefined;
  }
  const result = schema.safeParse(parsed);
  if (!result.success) {
    throw new ApiError(502, { code: ApiErrorCode.InvalidResponse, message: "The answer could not be read" });
  }
  return result.data;
}

/**
 * Asks the assistant a question, continuing sessionId when there is one. The
 * answer's text streams through handlers as it is written; the promise ends
 * with the answer as kept, or the error that stopped it.
 */
export async function askAssistant(
  input: MessageInput,
  sessionId: string | null,
  handlers: AskHandlers,
  signal: AbortSignal,
): Promise<AssistantAnswer> {
  const { body } = assistantQuestionSchema.parse(input);
  const stream = await browserEventStream("/api/v1/assistant/messages", {
    method: "POST",
    json: { question: body, session_id: sessionId ?? undefined },
    signal,
  });
  for await (const { event, data } of readServerSentEvents(stream)) {
    switch (event) {
      case "session":
        handlers.onSession(readEvent(assistantSessionEventSchema, data).session_id);
        break;
      case "delta":
        handlers.onText(readEvent(assistantDeltaEventSchema, data).text);
        break;
      case "answer":
        return readEvent(assistantAnswerSchema, data);
      case "error": {
        const failure = readEvent(assistantErrorEventSchema, data);
        throw new ApiError(failure.code === ApiErrorCode.AssistantUnavailable ? 503 : 500, failure);
      }
    }
  }
  throw new ApiError(502, { code: ApiErrorCode.InvalidResponse, message: "The answer ended before it was complete" });
}
