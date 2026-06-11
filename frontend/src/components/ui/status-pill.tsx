import { cn } from "@/lib/utils";

export const STATUS_PILL_VARIANTS = [
  "pending",
  "confirmed",
  "in_progress",
  "ready",
  "completed",
  "cancelled",
] as const;

export type StatusPillVariant = (typeof STATUS_PILL_VARIANTS)[number];

type StatusPillProps = {
  variant: StatusPillVariant;
  label: string;
  className?: string;
};

/** Semantic order-status pill — shared across internal roles. */
export function StatusPill({ variant, label, className }: StatusPillProps) {
  return (
    <span
      className={cn("status-pill", `status-pill--${variant}`, className)}
    >
      {label}
    </span>
  );
}

/** Map API order status strings to semantic pill variants. */
export function orderStatusToPillVariant(status: string): StatusPillVariant {
  switch (status) {
    case "pending":
      return "pending";
    case "confirmed":
      return "confirmed";
    case "in_progress":
      return "in_progress";
    case "ready":
      return "ready";
    case "fulfilled":
    case "completed":
      return "completed";
    case "cancelled":
      return "cancelled";
    default:
      return "pending";
  }
}

export function formatOrderStatusLabel(status: string): string {
  return status.replace(/_/g, " ");
}

/** Active / disabled account state — same pill system as order status. */
export function AccountStatusPill({ disabled }: { disabled: boolean }) {
  return (
    <StatusPill
      label={disabled ? "Disabled" : "Active"}
      variant={disabled ? "cancelled" : "completed"}
    />
  );
}

