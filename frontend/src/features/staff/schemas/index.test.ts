import { describe, expect, it } from "vitest";

import {
  INCIDENT_NOTE_MAX_LENGTH,
  createStaffOrderInputSchema,
  patchStaffOrderStatusInputSchema,
  reportIncidentInputSchema,
  staffOrderSchema,
  staffPickupsSchema,
} from "./index";

const order = {
  id: "00000000-0000-4000-8000-000000000001",
  code: "CH-260928-001",
  status: "confirmed",
  channel: "online",
  order_type: "pre_order",
  subtotal_cents: 4500,
  discount_cents: 0,
  total_cents: 4500,
  items: [],
  tickets: [],
  customer: { user_id: "00000000-0000-4000-8000-000000000002", email: "mai@example.com" },
  payment: { provider: "paypal", status: "captured", captured_at: "2026-09-28T09:05:00+07:00", refund_requested_at: null, refunded_at: null },
  created_at: "2026-09-28T09:00:00+07:00",
  updated_at: "2026-09-28T09:05:00+07:00",
};

// The counter's order history names the role behind every move.
describe("staff order timeline", () => {
  it("reads each move with the role that made it", () => {
    const result = staffOrderSchema.safeParse({
      ...order,
      timeline: [
        { status: "pending", actor_role: "customer", reason: null, at: "2026-09-28T09:00:00+07:00" },
        { status: "confirmed", actor_role: "staff", reason: null, at: "2026-09-28T09:05:00+07:00" },
      ],
    });
    expect(result.success).toBe(true);
  });

  it("reads a move the system made", () => {
    const result = staffOrderSchema.safeParse({
      ...order,
      status: "expired",
      timeline: [
        { status: "awaiting_payment", actor_role: "customer", reason: null, at: "2026-09-28T09:00:00+07:00" },
        { status: "expired", actor_role: null, reason: null, at: "2026-09-28T09:15:00+07:00" },
      ],
    });
    expect(result.success).toBe(true);
  });

  it("rejects a move without a known role", () => {
    const withoutRole = staffOrderSchema.safeParse({
      ...order,
      timeline: [{ status: "pending", reason: null, at: "2026-09-28T09:00:00+07:00" }],
    });
    const unknownRole = staffOrderSchema.safeParse({
      ...order,
      timeline: [{ status: "pending", actor_role: "system", reason: null, at: "2026-09-28T09:00:00+07:00" }],
    });
    expect(withoutRole.success).toBe(false);
    expect(unknownRole.success).toBe(false);
  });
});

// The counter sees what each station makes, to move a ticket nobody started.
describe("staff order tickets", () => {
  it("reads each station's ticket with its products", () => {
    const result = staffOrderSchema.parse({
      ...order,
      timeline: [],
      tickets: [
        {
          id: "00000000-0000-4000-8000-000000000003",
          station: "counter",
          status: "queued",
          items: [{ name: "Croissant", quantity: 2, customization: null }],
        },
      ],
    });
    expect(result.tickets[0]?.items).toEqual([{ name: "Croissant", quantity: 2, customization: null }]);
  });

  it("rejects an order without its tickets", () => {
    expect(staffOrderSchema.safeParse({ ...order, tickets: undefined, timeline: [] }).success).toBe(false);
  });
});

// A cancellation at the counter tells the customer why; a handover carries the customer's pickup code.
describe("staff order status move", () => {
  it("cancels only with a reason", () => {
    expect(patchStaffOrderStatusInputSchema.parse({ status: "cancelled", reason: " Out of matcha " })).toEqual({
      status: "cancelled",
      reason: "Out of matcha",
    });
    expect(patchStaffOrderStatusInputSchema.safeParse({ status: "cancelled" }).success).toBe(false);
    expect(patchStaffOrderStatusInputSchema.safeParse({ status: "cancelled", reason: "  " }).success).toBe(false);
  });

  it("moves forward without a reason", () => {
    expect(patchStaffOrderStatusInputSchema.parse({ status: "confirmed", reason: "x" })).toEqual({ status: "confirmed" });
  });

  it("hands over only with a 4-digit pickup code", () => {
    expect(patchStaffOrderStatusInputSchema.parse({ status: "fulfilled", pickup_code: "0427" })).toEqual({
      status: "fulfilled",
      pickup_code: "0427",
    });
    for (const pickup_code of [undefined, "427", "04270", "04a7"]) {
      expect(patchStaffOrderStatusInputSchema.safeParse({ status: "fulfilled", pickup_code }).success).toBe(false);
    }
  });
});

// The day's schedule names each pickup's time; one without a time is not a pickup.
describe("staff pickups", () => {
  const pickup = {
    id: order.id,
    code: order.code,
    status: "ready",
    channel: "online",
    total_cents: 4500,
    item_count: 2,
    customer: order.customer,
    pickup_at: "2026-10-01T09:00:00+07:00",
    created_at: order.created_at,
  };

  it("reads a page of a day's pickups and the slot length", () => {
    const result = staffPickupsSchema.parse({ slot_minutes: 30, pickups: [pickup] });
    expect(result.pickups[0]?.pickup_at).toBe("2026-10-01T09:00:00+07:00");
  });

  it("rejects a pickup without a time", () => {
    const result = staffPickupsSchema.safeParse({ slot_minutes: 30, pickups: [{ ...pickup, pickup_at: null }] });
    expect(result.success).toBe(false);
  });
});

// The counter takes orders for guests, who have no account, and hands them
// over as their cash is paid.
describe("orders staff take", () => {
  it("reads a guest's order, with no account behind it", () => {
    const guest = staffOrderSchema.parse({
      ...order,
      channel: "counter",
      timeline: [],
      customer: { user_id: null, email: null, display_name: "Lan", phone: "+84901234567" },
      payment: { provider: "cash", status: "created", captured_at: null, refund_requested_at: null, refunded_at: null },
    });
    expect(guest.customer.user_id).toBeNull();
    expect(guest.payment?.provider).toBe("cash");
  });

  it("hands an order over with its cash, or its code, never neither", () => {
    expect(patchStaffOrderStatusInputSchema.safeParse({ status: "fulfilled", cash_collected: true }).success).toBe(true);
    expect(patchStaffOrderStatusInputSchema.safeParse({ status: "fulfilled", cash_collected: false }).success).toBe(false);
    expect(patchStaffOrderStatusInputSchema.safeParse({ status: "fulfilled" }).success).toBe(false);
  });

  it("takes a guest with a name and a Vietnam mobile number", () => {
    const taken = createStaffOrderInputSchema.parse({
      channel: "counter",
      guest: { name: " Lan ", phone: "0901 234 567" },
      pickup_at: "2026-10-02T09:00:00+07:00",
      items: [{ product_id: "00000000-0000-4000-8000-000000000003", quantity: 2 }],
    });
    expect(taken.guest).toEqual({ name: "Lan", phone: "+84901234567" });
    const noPhone = createStaffOrderInputSchema.safeParse({
      channel: "counter",
      guest: { name: "Lan", phone: "" },
      pickup_at: "2026-10-02T09:00:00+07:00",
      items: [{ product_id: "00000000-0000-4000-8000-000000000003", quantity: 2 }],
    });
    expect(noPhone.success).toBe(false);
  });
});

describe("reportIncidentInputSchema", () => {
  it("takes a type staff report and trims the note", () => {
    expect(reportIncidentInputSchema.parse({ type: "custom_mismatch", note: "  Wrote Mai, not May. " })).toEqual({
      type: "custom_mismatch",
      note: "Wrote Mai, not May.",
    });
  });

  it("asks what went wrong and refuses a type the system records", () => {
    const missing = reportIncidentInputSchema.safeParse({ note: "Box crushed" });
    expect(missing.success).toBe(false);
    expect(missing.error?.issues[0]?.message).toBe("Choose what went wrong");
    expect(reportIncidentInputSchema.safeParse({ type: "no_show", note: "Box crushed" }).success).toBe(false);
  });

  it("counts the note's characters as the API does, an emoji as one", () => {
    const longest = "🎂".repeat(INCIDENT_NOTE_MAX_LENGTH);
    expect(reportIncidentInputSchema.safeParse({ type: "other", note: longest }).success).toBe(true);
    expect(reportIncidentInputSchema.safeParse({ type: "other", note: `${longest}a` }).success).toBe(false);
    expect(reportIncidentInputSchema.safeParse({ type: "other", note: " \n " }).success).toBe(false);
  });
});
