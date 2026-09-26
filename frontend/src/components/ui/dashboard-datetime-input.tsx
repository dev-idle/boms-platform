"use client";

import * as React from "react";
import { useEffect, useId, useLayoutEffect, useRef, useState } from "react";

import { DashboardCalendarIcon } from "@/components/icons/dashboard-ui-icons";
import { DashboardDatetimePicker } from "@/components/ui/dashboard-datetime-picker";
import { cn } from "@/lib/utils";
import {
  formatDashboardDatetime,
  parseIsoToLocalParts,
  partsToIso,
  type LocalDatetimeParts,
} from "@/lib/validation/datetime-calendar";

export type DashboardDatetimeInputProps = {
  className?: string;
  disabled?: boolean;
  id?: string;
  onChange: (iso: string) => void;
  ref?: React.Ref<HTMLButtonElement>;
  value: string;
} & Pick<React.ButtonHTMLAttributes<HTMLButtonElement>, "aria-invalid">;

function isPortaledSelectTarget(target: EventTarget | null): boolean {
  return (
    target instanceof Element &&
    Boolean(target.closest(".field-select-content, [data-radix-popper-content-wrapper]"))
  );
}

/** Clearance kept between the panel and the edge of the window. */
const PANEL_VIEWPORT_MARGIN = 12;

type PanelPlacement = {
  /** How much room that side leaves, so a panel taller than it can scroll.
   * Null until measured: an unmeasured panel must render at its natural height,
   * or the first measurement reads a cap of its own making. */
  space: number | null;
  side: "above" | "below";
};

const INITIAL_PLACEMENT: PanelPlacement = { side: "below", space: null };

/** Below the field, or above it when the panel would otherwise run off-screen. */
function measurePlacement(field: DOMRect, panelHeight: number): PanelPlacement {
  const below = window.innerHeight - field.bottom - PANEL_VIEWPORT_MARGIN;
  const above = field.top - PANEL_VIEWPORT_MARGIN;
  const side = panelHeight > below && above > below ? "above" : "below";
  return { side, space: Math.max(side === "above" ? above : below, 0) };
}

/** Dashboard datetime field — themed popover; draft commits on Set only. */
export function DashboardDatetimeInput({
  className,
  disabled = false,
  id,
  onChange,
  value,
  ref,
  ...props
}: DashboardDatetimeInputProps) {
  const rootRef = useRef<HTMLDivElement>(null);
  const panelRef = useRef<HTMLDivElement>(null);
  const panelId = useId();
  const [open, setOpen] = useState(false);
  const [placement, setPlacement] = useState<PanelPlacement>(INITIAL_PLACEMENT);
  const [draft, setDraft] = useState<LocalDatetimeParts>(() =>
    parseIsoToLocalParts(value),
  );
  const [viewMonth, setViewMonth] = useState(() =>
    parseIsoToLocalParts(value).month,
  );
  const [viewYear, setViewYear] = useState(() =>
    parseIsoToLocalParts(value).year,
  );

  // No value → draft sync effect: the panel only renders while open, and openPanel()
  // re-seeds draft and view from the current value every time it opens.
  useEffect(() => {
    if (!open) {
      return;
    }

    function handlePointerDown(event: MouseEvent): void {
      if (rootRef.current?.contains(event.target as Node)) {
        return;
      }
      if (isPortaledSelectTarget(event.target)) {
        return;
      }
      setOpen(false);
    }

    function handleKeyDown(event: KeyboardEvent): void {
      if (event.key === "Escape") {
        setOpen(false);
      }
    }

    document.addEventListener("mousedown", handlePointerDown);
    document.addEventListener("keydown", handleKeyDown);
    return () => {
      document.removeEventListener("mousedown", handlePointerDown);
      document.removeEventListener("keydown", handleKeyDown);
    };
  }, [open]);

  // A field low on the page would otherwise open a panel that runs past the
  // bottom of the window, leaving the time row and Set out of reach.
  useLayoutEffect(() => {
    if (!open) {
      return;
    }

    function place(): void {
      const field = rootRef.current;
      const panel = panelRef.current;
      if (!field || !panel) {
        return;
      }
      // `scrollHeight`, not `offsetHeight`: the panel is already capped by the
      // space this effect wrote last time, so measuring its box would compare
      // the cap with itself and the flip could never fire.
      const next = measurePlacement(
        field.getBoundingClientRect(),
        panel.scrollHeight,
      );
      setPlacement((current) =>
        current.side === next.side &&
        current.space !== null &&
        Math.abs(current.space - (next.space ?? 0)) < 1
          ? current
          : next,
      );
    }

    place();
    window.addEventListener("resize", place);
    window.addEventListener("scroll", place, true);
    return () => {
      window.removeEventListener("resize", place);
      window.removeEventListener("scroll", place, true);
    };
  }, [open]);

  function openPanel(): void {
    const current = parseIsoToLocalParts(value);
    setDraft(current);
    setViewMonth(current.month);
    setViewYear(current.year);
    setOpen(true);
  }

  function handleToday(): void {
    const now = parseIsoToLocalParts(new Date().toISOString());
    setViewMonth(now.month);
    setViewYear(now.year);
    setDraft(now);
  }

  function handleSet(): void {
    onChange(partsToIso(draft));
    setOpen(false);
  }

  return (
    <div className={cn("dashboard-datetime", className)} ref={rootRef}>
      <button
        {...props}
        aria-controls={panelId}
        aria-expanded={open}
        aria-haspopup="dialog"
        className={cn(
          "dashboard-datetime__trigger field-chrome text-form-input",
          open && "dashboard-datetime__trigger--open",
        )}
        disabled={disabled}
        id={id}
        onClick={() => {
          if (open) {
            setOpen(false);
            return;
          }
          openPanel();
        }}
        ref={ref}
        type="button"
      >
        <span className="dashboard-datetime__trigger-value">
          {formatDashboardDatetime(value)}
        </span>
      </button>
      <span aria-hidden className="dashboard-datetime__trigger-icon-wrap">
        <DashboardCalendarIcon className="dashboard-datetime__trigger-icon size-4" />
      </span>

      {open ? (
        <div
          className={cn(
            "dashboard-datetime__panel",
            placement.side === "above" && "dashboard-datetime__panel--above",
          )}
          id={panelId}
          ref={panelRef}
          role="dialog"
          aria-label="Choose date and time"
          style={
            placement.space === null
              ? undefined
              : ({
                  "--datetime-panel-space": `${placement.space}px`,
                } as React.CSSProperties)
          }
        >
          <DashboardDatetimePicker
            draft={draft}
            onDraftChange={setDraft}
            onSet={handleSet}
            onToday={handleToday}
            onViewMonthChange={(month, year) => {
              setViewMonth(month);
              setViewYear(year);
            }}
            viewMonth={viewMonth}
            viewYear={viewYear}
          />
        </div>
      ) : null}
    </div>
  );
}
