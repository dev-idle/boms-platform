import { describe, expect, it } from "vitest";

import { ASSISTANT_QUESTION_MAX_LENGTH, assistantAnswerSchema, assistantQuestionSchema } from "./index";

describe("assistantQuestionSchema", () => {
  it("takes a question of up to the API's length, counted as the API counts", () => {
    expect(assistantQuestionSchema.safeParse({ body: "🎂".repeat(ASSISTANT_QUESTION_MAX_LENGTH) }).success).toBe(true);
    expect(assistantQuestionSchema.safeParse({ body: "a".repeat(ASSISTANT_QUESTION_MAX_LENGTH + 1) }).success).toBe(false);
    expect(assistantQuestionSchema.safeParse({ body: "   " }).success).toBe(false);
  });
});

describe("assistantAnswerSchema", () => {
  const answer = { reply: "We open at 07:00.", offer_staff: false, hand_off: null };

  it("reads an answer, and one passed to staff on an order or in general", () => {
    expect(assistantAnswerSchema.safeParse(answer).success).toBe(true);
    expect(
      assistantAnswerSchema.safeParse({
        ...answer,
        hand_off: { order_id: "0b6c8f5e-3c1d-4a8e-9f0a-2d7e6c5b4a39", order_code: "CH-261125-004" },
      }).success,
    ).toBe(true);
    expect(assistantAnswerSchema.safeParse({ ...answer, hand_off: { order_id: null, order_code: null } }).success).toBe(true);
  });

  it("refuses an answer with no reply", () => {
    expect(assistantAnswerSchema.safeParse({ ...answer, reply: "" }).success).toBe(false);
  });
});
