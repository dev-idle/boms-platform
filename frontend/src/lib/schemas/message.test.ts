import { describe, expect, it } from "vitest";

import { MESSAGE_MAX_LENGTH, messageInputSchema, messageSchema } from "./message";

describe("messageInputSchema", () => {
  it("counts characters as the API does, an emoji as one", () => {
    const longest = "🎂".repeat(MESSAGE_MAX_LENGTH);
    expect(messageInputSchema.safeParse({ body: longest }).success).toBe(true);
    expect(messageInputSchema.safeParse({ body: `${longest}a` }).success).toBe(false);
  });

  it("trims what is written and refuses an empty message", () => {
    expect(messageInputSchema.parse({ body: "  Is it nut free?\n" }).body).toBe("Is it nut free?");
    expect(messageInputSchema.safeParse({ body: " \n " }).success).toBe(false);
  });
});

describe("messageSchema", () => {
  it("reads any message the API accepted", () => {
    const message = {
      id: "0b6c8f5e-3c1d-4a8e-9f0a-2d7e6c5b4a39",
      from: "staff",
      author_name: null,
      body: "🎂".repeat(MESSAGE_MAX_LENGTH),
      created_at: "2026-10-02T09:00:00+07:00",
    };
    expect(messageSchema.safeParse(message).success).toBe(true);
  });
});
