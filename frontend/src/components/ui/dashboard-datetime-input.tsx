"use client";

import * as React from "react";
import { useEffect, useId, useRef, useState } from "react";

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
  value: string;
} & Pick<React.ButtonHTMLAttributes<HTMLButtonElement>, "aria-invalid">;

function isPortaledSelectTarget(target: EventTarget | null): boolean {
  return (
    target instanceof Element &&
    Boolean(target.closest(".field-select-content, [data-radix-popper-content-wrapper]"))
  );
}

/** Dashboard datetime field — themed popover; draft commits on Set only. */
export const DashboardDatetimeInput = React.forwardRef<
  HTMLButtonElement,
  DashboardDatetimeInputProps
>(function DashboardDatetimeInput(
  { className, disabled = false, id, onChange, value, ...props },
  ref,
) {
  const rootRef = useRef<HTMLDivElement>(null);
  const panelId = useId();
  const [open, setOpen] = useState(false);
  const [draft, setDraft] = useState<LocalDatetimeParts>(() =>
    parseIsoToLocalParts(value),
  );
  const [viewMonth, setViewMonth] = useState(() =>
    parseIsoToLocalParts(value).month,
  );
  const [viewYear, setViewYear] = useState(() =>
    parseIsoToLocalParts(value).year,
  );

  useEffect(() => {
    const next = parseIsoToLocalParts(value);
    if (!open) {
      setDraft(next);
      setViewMonth(next.month);
      setViewYear(next.year);
    }
  }, [open, value]);

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
          className="dashboard-datetime__panel"
          id={panelId}
          role="dialog"
          aria-label="Choose date and time"
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
});
DashboardDatetimeInput.displayName = "DashboardDatetimeInput";
