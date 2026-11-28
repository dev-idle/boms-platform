"use client";

import { DateInput } from "@/components/ui/date-input";
import { Label } from "@/components/ui/label";
import { Select } from "@/components/ui/select";
import { orderStatusSchema, type OrderStatus } from "@/lib/schemas/order";
import { bakeryDayOf } from "@/lib/validation/pickup";

import { orderHeadline } from "../lib/order-progress";

const ALL_STATUSES = "all";

/** Every status a row can show, labelled as its status dot reads. */
const ORDER_HISTORY_STATUS_OPTIONS: Array<{ value: OrderStatus | typeof ALL_STATUSES; label: string }> = [
  { value: ALL_STATUSES, label: "All orders" },
  ...orderStatusSchema.options.map((status) => ({ value: status, label: orderHeadline(status) })),
];

export type OrderHistoryFilterValue = {
  status: OrderStatus | undefined;
  /** Bakery days (YYYY-MM-DD) the orders were placed on; empty when open. */
  from: string;
  to: string;
};

type OrderHistoryFiltersProps = {
  value: OrderHistoryFilterValue;
  onChange: (value: OrderHistoryFilterValue) => void;
};

/**
 * Status and placed-on range for the order history. A day picked past the
 * other end of the range moves that end with it, so the range is never empty
 * by construction.
 */
export function OrderHistoryFilters({ value, onChange }: OrderHistoryFiltersProps) {
  const today = bakeryDayOf(new Date());

  return (
    <div className="storefront-orders__filters" role="group" aria-label="Filter orders">
      <div className="storefront-orders__filter storefront-orders__filter--status">
        <Label htmlFor="orders-status">Status</Label>
        <Select
          contentClassName="field-select-content--storefront"
          id="orders-status"
          value={value.status ?? ALL_STATUSES}
          onChange={(event) => {
            const parsed = orderStatusSchema.safeParse(event.target.value);
            onChange({ ...value, status: parsed.success ? parsed.data : undefined });
          }}
        >
          {ORDER_HISTORY_STATUS_OPTIONS.map((option) => (
            <option key={option.value} value={option.value}>
              {option.label}
            </option>
          ))}
        </Select>
      </div>
      <div className="storefront-orders__filter">
        <Label htmlFor="orders-placed-from">Placed from</Label>
        <DateInput
          id="orders-placed-from"
          max={value.to || today}
          placeholder="Any day"
          value={value.from}
          onChange={(from) => {
            const to = from && value.to && value.to < from ? from : value.to;
            onChange({ ...value, from, to });
          }}
        />
      </div>
      <div className="storefront-orders__filter">
        <Label htmlFor="orders-placed-to">Placed to</Label>
        <DateInput
          align="end"
          id="orders-placed-to"
          max={today}
          min={value.from || undefined}
          placeholder="Any day"
          value={value.to}
          onChange={(to) => {
            const from = to && value.from && value.from > to ? to : value.from;
            onChange({ ...value, from, to });
          }}
        />
      </div>
    </div>
  );
}
