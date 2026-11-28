"use client";

import { useEffect, useId, useRef, useState, type ButtonHTMLAttributes } from "react";

import { DashboardCalendarIcon } from "@/components/icons/dashboard-ui-icons";
import { Button } from "@/components/ui/button";
import { CalendarMonth } from "@/components/ui/calendar-month";
import { cn } from "@/lib/utils";
import type { CalendarDay } from "@/lib/validation/datetime-calendar";
import { bakeryDayOf } from "@/lib/validation/pickup";

type DateInputProps = {
  /** A day as YYYY-MM-DD, or "" while none is chosen. */
  value: string;
  onChange: (value: string) => void;
  /** The first and the last day that can be chosen, as YYYY-MM-DD. */
  min?: string;
  max?: string;
  /** What the empty field reads. */
  placeholder: string;
  /** Names the field where no visible label does, e.g. in a toolbar. */
  label?: string;
  /** A day must stay chosen: the panel offers no Clear, and `onChange` never gets "". */
  required?: boolean;
  /** Which edge of the field the calendar hangs from. */
  align?: "start" | "end";
  // The id the label beside it points at, or what the form field around it
  // hands down: its id, its hint and whether it is invalid.
} & Pick<ButtonHTMLAttributes<HTMLButtonElement>, "id" | "aria-describedby" | "aria-invalid">;

const DAY_LABEL = new Intl.DateTimeFormat("en-US", { month: "short", day: "numeric", year: "numeric" });

function parseDay(value: string): CalendarDay | null {
  const [year, month, day] = value.split("-").map(Number);
  return year && month && day ? { year, month: month - 1, day } : null;
}

function formatDay(day: CalendarDay): string {
  return `${day.year}-${String(day.month + 1).padStart(2, "0")}-${String(day.day).padStart(2, "0")}`;
}

/** Today at the bakery, which is the day every range the field is given is counted in. */
function bakeryToday(): CalendarDay {
  const [year, month, day] = bakeryDayOf(new Date()).split("-").map(Number);
  return { year, month: month - 1, day };
}

/** A day field with the themed calendar: pressing a day chooses it and closes the panel. */
export function DateInput({
  value,
  onChange,
  min,
  max,
  placeholder,
  label,
  required = false,
  align = "start",
  ...field
}: DateInputProps) {
  const rootRef = useRef<HTMLDivElement>(null);
  const triggerRef = useRef<HTMLButtonElement>(null);
  const panelId = useId();
  const selected = parseDay(value);
  const shown = selected ? DAY_LABEL.format(new Date(selected.year, selected.month, selected.day)) : placeholder;
  const [today, setToday] = useState(bakeryToday);
  const [open, setOpen] = useState(false);
  const [cursor, setCursor] = useState<CalendarDay>(today);
  const [view, setView] = useState({ month: today.month, year: today.year });

  useEffect(() => {
    if (!open) {
      return;
    }

    function handlePointerDown(event: MouseEvent): void {
      if (!rootRef.current?.contains(event.target as Node)) {
        setOpen(false);
      }
    }

    function handleKeyDown(event: KeyboardEvent): void {
      if (event.key === "Escape") {
        setOpen(false);
        triggerRef.current?.focus();
      }
    }

    document.addEventListener("mousedown", handlePointerDown);
    document.addEventListener("keydown", handleKeyDown);
    return () => {
      document.removeEventListener("mousedown", handlePointerDown);
      document.removeEventListener("keydown", handleKeyDown);
    };
  }, [open]);

  function outOfRange(day: CalendarDay): boolean {
    const key = formatDay(day);
    return (min !== undefined && key < min) || (max !== undefined && key > max);
  }

  function openPanel(): void {
    // Read as the panel opens: a tab left open overnight still offers the right day.
    const now = bakeryToday();
    setToday(now);
    const start = selected ?? now;
    setCursor(start);
    setView({ month: start.month, year: start.year });
    setOpen(true);
  }

  function choose(next: string): void {
    onChange(next);
    setOpen(false);
    triggerRef.current?.focus();
  }

  return (
    <div className="dashboard-datetime" ref={rootRef}>
      <button
        {...field}
        aria-controls={panelId}
        aria-expanded={open}
        aria-haspopup="dialog"
        aria-label={label ? `${label}: ${shown}` : undefined}
        className={cn(
          "dashboard-datetime__trigger field-chrome text-form-input",
          open && "dashboard-datetime__trigger--open",
        )}
        onClick={() => (open ? setOpen(false) : openPanel())}
        ref={triggerRef}
        type="button"
      >
        <span
          className={cn(
            "dashboard-datetime__trigger-value",
            !selected && "dashboard-datetime__trigger-value--empty",
          )}
        >
          {shown}
        </span>
      </button>
      <span aria-hidden className="dashboard-datetime__trigger-icon-wrap">
        <DashboardCalendarIcon className="dashboard-datetime__trigger-icon size-4" />
      </span>

      {open ? (
        <div
          aria-label="Choose a day"
          className={cn("dashboard-datetime__panel", align === "end" && "dashboard-datetime__panel--end")}
          id={panelId}
          role="dialog"
        >
          <div className="dashboard-datetime__panel-body">
            <CalendarMonth
              cursor={cursor}
              isDisabled={outOfRange}
              onCursorChange={setCursor}
              onPick={(day) => choose(formatDay(day))}
              onViewMonthChange={(month, year) => setView({ month, year })}
              selected={selected}
              viewMonth={view.month}
              viewYear={view.year}
            />
            <div className="dashboard-datetime__actions">
              {required ? null : (
                <Button
                  disabled={!selected}
                  onClick={() => choose("")}
                  size="sm"
                  title={selected ? undefined : "No day is chosen"}
                  type="button"
                  variant="ghost"
                >
                  Clear
                </Button>
              )}
              <Button
                disabled={outOfRange(today)}
                onClick={() => choose(formatDay(today))}
                size="sm"
                title={outOfRange(today) ? "Today is outside the range" : undefined}
                type="button"
                variant="outline"
              >
                Today
              </Button>
            </div>
          </div>
        </div>
      ) : null}
    </div>
  );
}
