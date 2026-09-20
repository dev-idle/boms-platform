import type { OrderStatus } from "@/lib/schemas/order";
import { cn } from "@/lib/utils";

export type StatusPillVariant =
  | "pending"
  | "confirmed"
  | "in_progress"
  | "ready"
  | "completed"
  | "cancelled";

type StatusPillProps = {
  variant: StatusPillVariant;
  label: string;
  className?: string;
};

/**
 * Semantic order-status indicator — dot plus micro-label. Colour is never the
 * only channel: the label always accompanies the dot.
 */
export function StatusPill({ variant, label, className }: StatusPillProps) {
  return (
    <span className={cn("status-pill", `status-pill--${variant}`, className)}>
      <span aria-hidden className="status-pill__dot" />
      {label}
    </span>
  );
}

/** Map API order status strings to semantic pill variants. */
export function orderStatusToPillVariant(status: OrderStatus): StatusPillVariant {
  switch (status) {
    case "pending":
      return "pending";
    case "confirmed":
      return "confirmed";
    case "in_production":
      return "in_progress";
    case "ready":
      return "ready";
    case "fulfilled":
      return "completed";
    case "cancelled":
      return "cancelled";
  }
}

export function formatOrderStatusLabel(status: string): string {
  return status.replace(/_/g, " ");
}

/** Active / disabled account state — same pill system as order status. */
export function CatalogAvailabilityPill({ available }: { available: boolean }) {
  return (
    <StatusPill
      label={available ? "Available" : "Unavailable"}
      variant={available ? "ready" : "completed"}
    />
  );
}

/** Generic active/inactive entity state for manager catalog tables. */
export function EntityActivePill({ active }: { active: boolean }) {
  return (
    <StatusPill
      label={active ? "Active" : "Inactive"}
      variant={active ? "ready" : "completed"}
    />
  );
}

