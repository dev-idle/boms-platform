"use client";

import { useEffect, useRef, useState } from "react";

import { ApiErrorCode, isApiError } from "@/lib/errors";
import type { Message, MessageInput } from "@/lib/schemas/message";

import { askAssistant } from "../api";
import type { AssistantAnswer } from "../schemas";

/** Why a question got no answer, as the visitor reads it. */
function failureText(error: unknown): string {
  if (isApiError(error) && error.isRateLimited()) {
    return "You are asking quickly. Wait a moment, then ask again.";
  }
  if (isApiError(error) && error.code === ApiErrorCode.AssistantUnavailable) {
    return "The assistant cannot answer right now. Our staff can help you instead.";
  }
  if (isApiError(error) && error.hasValidationDetails()) {
    return "Questions take plain text: remove tabs or hidden characters, then ask again.";
  }
  return "We could not get an answer. Please ask again.";
}

/**
 * A conversation with the assistant, kept while the page is open: the turns
 * as messages (the visitor's as "customer", the assistant's as "staff"), the
 * answer streaming into the last one, and what the last answer offers.
 * onGeneralHandOff runs once a question went to the customer's general
 * conversation with staff.
 */
export function useAssistantChat(onGeneralHandOff: () => void) {
  const [turns, setTurns] = useState<Message[]>([]);
  const [sessionId, setSessionId] = useState<string | null>(null);
  const [pending, setPending] = useState(false);
  const [failure, setFailure] = useState<string | null>(null);
  const [lastAnswer, setLastAnswer] = useState<AssistantAnswer | null>(null);
  const controller = useRef<AbortController | null>(null);

  // A visitor who leaves mid-answer stops it.
  useEffect(() => () => controller.current?.abort(), []);

  async function ask(input: MessageInput, sent: () => void): Promise<void> {
    const now = new Date().toISOString();
    const answerId = crypto.randomUUID();
    const question: Message = { id: crypto.randomUUID(), from: "customer", author_name: null, body: input.body.trim(), created_at: now };
    const placeholder: Message = { id: answerId, from: "staff", author_name: null, body: "…", created_at: now };
    const writeAnswer = (body: (current: string) => string) =>
      setTurns((all) => all.map((turn) => (turn.id === answerId ? { ...turn, body: body(turn.body) } : turn)));

    controller.current = new AbortController();
    setTurns((all) => [...all, question, placeholder]);
    setPending(true);
    setFailure(null);
    setLastAnswer(null);
    sent();
    let streamed = "";
    try {
      const answer = await askAssistant(input, sessionId, {
        onSession: setSessionId,
        onText: (text) => {
          streamed += text;
          writeAnswer(() => streamed);
        },
      }, controller.current.signal);
      writeAnswer(() => answer.reply);
      setLastAnswer(answer);
      if (answer.hand_off && !answer.hand_off.order_id) {
        onGeneralHandOff();
      }
    } catch (error) {
      if (controller.current.signal.aborted) {
        return;
      }
      setTurns((all) => all.filter((turn) => turn.id !== answerId));
      setFailure(failureText(error));
    } finally {
      setPending(false);
    }
  }

  return { turns, pending, failure, lastAnswer, ask };
}

export type AssistantChatState = ReturnType<typeof useAssistantChat>;
