import type { MouseEventHandler, ReactNode } from "react";

import { cn } from "@/lib/utils";

type DashboardIconActionProps = {
  children: ReactNode;
  className?: string;
  disabled?: boolean;
  /** Accessible name — the icon carries no text. */
  label: string;
  onClick: () => void;
  onMouseDown?: MouseEventHandler<HTMLButtonElement>;
  tone?: "default" | "danger";
};

/**
 * Icon-only control for form composers (image lists, combo item rows).
 * Table rows use text links instead — an icon inside a data row needs a legend.
 */
export function DashboardIconAction({
  children,
  className,
  disabled = false,
  label,
  onClick,
  onMouseDown,
  tone = "default",
}: DashboardIconActionProps) {
  return (
    <button
      aria-label={label}
      className={cn(
        "db-icon-action",
        tone === "danger" && "db-icon-action--danger",
        className,
      )}
      disabled={disabled}
      onClick={onClick}
      onMouseDown={onMouseDown}
      title={label}
      type="button"
    >
      {children}
    </button>
  );
}
