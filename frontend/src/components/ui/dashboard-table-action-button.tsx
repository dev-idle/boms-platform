import { cn } from "@/lib/utils";

type DashboardTableActionButtonProps = {
  /** Accessible name — carries the row identity (e.g. "Disable jane@example.com"). */
  label: string;
  onClick: () => void;
  text: string;
  tone: "accent" | "danger" | "warning";
  /** When set, the action exists but cannot apply: greyed, with the reason as its title. */
  blockedReason?: string;
};

const TONE_CLASS = {
  accent: "db-table-action--edit",
  danger: "db-table-action--delete",
  warning: "db-table-action--revoke",
} as const;

/**
 * Text row action with a fixed slot. A blocked action stays in place, greyed —
 * `disabled` blocks activation, `aria-disabled` announces it, and
 * `tabIndex={-1}` keeps it out of the tab order.
 */
export function DashboardTableActionButton({
  blockedReason,
  label,
  onClick,
  text,
  tone,
}: DashboardTableActionButtonProps) {
  const blocked = Boolean(blockedReason);

  return (
    <button
      aria-disabled={blocked || undefined}
      aria-label={label}
      className={cn("db-table-action", TONE_CLASS[tone])}
      disabled={blocked}
      onClick={blocked ? undefined : onClick}
      tabIndex={blocked ? -1 : undefined}
      title={blockedReason}
      type="button"
    >
      {text}
    </button>
  );
}
