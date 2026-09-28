"use client";

import { useRealtimeStatus, type RealtimeStatus } from "@/lib/realtime";

import { StatusPill, type StatusPillVariant } from "./status-pill";

const LIVE_STATE: Record<
  Exclude<RealtimeStatus, "idle">,
  { label: string; variant: StatusPillVariant }
> = {
  connecting: { label: "Connecting", variant: "completed" },
  live: { label: "Live", variant: "ready" },
  reconnecting: { label: "Reconnecting", variant: "pending" },
};

/**
 * Whether changes made elsewhere reach this tab as they happen. A status
 * region, so assistive tech announces a lost or restored connection. Hidden
 * while the tab is not trying to receive pushes at all.
 */
export function LiveIndicator() {
  const status = useRealtimeStatus();
  if (status === "idle") {
    return null;
  }
  const { label, variant } = LIVE_STATE[status];
  return (
    <div role="status">
      <StatusPill label={label} variant={variant} />
    </div>
  );
}
