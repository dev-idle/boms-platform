"use client";

import { useEffect, useRef, useState, type KeyboardEvent } from "react";

import {
  DashboardChevronLeftIcon,
  DashboardChevronRightIcon,
} from "@/components/icons/dashboard-ui-icons";
import { cn } from "@/lib/utils";
import {
  buildCalendarMonthGrid,
  calendarDayKey,
  CALENDAR_WEEKDAY_LABELS,
  formatCalendarDayLabel,
  formatMonthLabel,
  isSameLocalDay,
  shiftDraftDays,
  type CalendarDay,
} from "@/lib/validation/datetime-calendar";

/** Arrow / Home / End / PageUp / PageDown → day offset from the focused day. */
function keyDayOffset(key: string, day: CalendarDay): number | null {
  const weekday = new Date(day.year, day.month, day.day).getDay();
  switch (key) {
    case "ArrowLeft":
      return -1;
    case "ArrowRight":
      return 1;
    case "ArrowUp":
      return -7;
    case "ArrowDown":
      return 7;
    case "Home":
      return -weekday;
    case "End":
      return 6 - weekday;
    case "PageUp":
      return -new Date(day.year, day.month, 0).getDate();
    case "PageDown":
      return new Date(day.year, day.month + 1, 0).getDate();
    default:
      return null;
  }
}

function localToday(): CalendarDay {
  const now = new Date();
  return { year: now.getFullYear(), month: now.getMonth(), day: now.getDate() };
}

type CalendarMonthProps = {
  /** The chosen day, filled; null while none is chosen. */
  selected: CalendarDay | null;
  /** The day the arrow keys stand on, and the one Tab lands on. */
  cursor: CalendarDay;
  viewMonth: number;
  viewYear: number;
  onViewMonthChange: (month: number, year: number) => void;
  /** An arrow key moved the cursor. */
  onCursorChange: (day: CalendarDay) => void;
  /** A day was pressed. */
  onPick: (day: CalendarDay) => void;
  /** Days outside the allowed range: shown, never chosen. */
  isDisabled?: (day: CalendarDay) => boolean;
};

/** One month of days with its header — the calendar every date field opens. */
export function CalendarMonth({
  selected,
  cursor,
  viewMonth,
  viewYear,
  onViewMonthChange,
  onCursorChange,
  onPick,
  isDisabled,
}: CalendarMonthProps) {
  const [today] = useState(localToday);
  const gridRef = useRef<HTMLDivElement>(null);
  const focusAfterMove = useRef(false);
  const cells = buildCalendarMonthGrid(viewYear, viewMonth);
  const monthLabel = formatMonthLabel(viewYear, viewMonth);
  const cursorKey = calendarDayKey(cursor);
  const rovingKey = cells.some((cell) => calendarDayKey(cell) === cursorKey)
    ? cursorKey
    : calendarDayKey({ year: viewYear, month: viewMonth, day: 1 });

  useEffect(() => {
    if (!focusAfterMove.current) {
      return;
    }
    focusAfterMove.current = false;
    gridRef.current
      ?.querySelector<HTMLButtonElement>(`[data-day="${cursorKey}"]`)
      ?.focus();
  }, [cursorKey]);

  function shiftMonth(delta: number): void {
    const next = new Date(viewYear, viewMonth + delta, 1);
    onViewMonthChange(next.getMonth(), next.getFullYear());
  }

  function handleDayKeyDown(event: KeyboardEvent<HTMLButtonElement>, day: CalendarDay): void {
    const offset = keyDayOffset(event.key, day);
    if (offset === null) {
      return;
    }
    event.preventDefault();
    const next = shiftDraftDays(day, offset);
    focusAfterMove.current = true;
    onCursorChange(next);
    if (next.month !== viewMonth || next.year !== viewYear) {
      onViewMonthChange(next.month, next.year);
    }
  }

  return (
    <div className="dashboard-datetime__calendar">
      <div className="dashboard-datetime__nav">
        <button
          aria-label="Previous month"
          className="dashboard-datetime__nav-btn"
          onClick={() => shiftMonth(-1)}
          type="button"
        >
          <DashboardChevronLeftIcon className="size-4" />
        </button>
        <span aria-live="polite" className="dashboard-datetime__month-label">
          {monthLabel}
        </span>
        <button
          aria-label="Next month"
          className="dashboard-datetime__nav-btn"
          onClick={() => shiftMonth(1)}
          type="button"
        >
          <DashboardChevronRightIcon className="size-4" />
        </button>
      </div>

      <div aria-hidden className="dashboard-datetime__weekdays">
        {CALENDAR_WEEKDAY_LABELS.map((label) => (
          <span key={label}>{label}</span>
        ))}
      </div>

      <div aria-label={monthLabel} className="dashboard-datetime__grid" ref={gridRef} role="group">
        {cells.map((cell) => {
          const key = calendarDayKey(cell);
          const day: CalendarDay = { year: cell.year, month: cell.month, day: cell.day };
          const isSelected = selected !== null && isSameLocalDay(selected, cell);
          const isToday = isSameLocalDay(today, cell);
          const disabled = isDisabled?.(day) ?? false;
          return (
            <button
              aria-current={isToday ? "date" : undefined}
              // Not `disabled`: the arrow keys still cross a day that cannot be chosen.
              aria-disabled={disabled || undefined}
              aria-label={formatCalendarDayLabel(cell)}
              aria-pressed={isSelected}
              className={cn(
                "dashboard-datetime__day",
                !cell.inCurrentMonth && "dashboard-datetime__day--outside",
                isToday && "dashboard-datetime__day--today",
                isSelected && "dashboard-datetime__day--selected",
              )}
              data-day={key}
              key={key}
              onClick={() => {
                if (!disabled) {
                  onPick(day);
                }
              }}
              onKeyDown={(event) => handleDayKeyDown(event, day)}
              tabIndex={key === rovingKey ? 0 : -1}
              type="button"
            >
              {cell.day}
            </button>
          );
        })}
      </div>
    </div>
  );
}
