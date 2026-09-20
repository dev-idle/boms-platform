"use client";

import { useEffect, useId, useRef, useState, type KeyboardEvent } from "react";

import {
  DashboardChevronLeftIcon,
  DashboardChevronRightIcon,
} from "@/components/icons/dashboard-ui-icons";
import { Button } from "@/components/ui/button";
import { Select } from "@/components/ui/select";
import { cn } from "@/lib/utils";
import {
  buildCalendarMonthGrid,
  buildHour12Options,
  buildMinuteOptions,
  calendarDayKey,
  CALENDAR_WEEKDAY_LABELS,
  formatCalendarDayLabel,
  formatMinuteLabel,
  formatMonthLabel,
  hour12To24,
  hour24To12,
  isSameLocalDay,
  parseIsoToLocalParts,
  shiftDraftDays,
  type CalendarCell,
  type LocalDatetimeParts,
} from "@/lib/validation/datetime-calendar";

const DATETIME_SELECT_MENU_CLASS = "field-select-content--datetime";

/** Arrow / Home / End / PageUp / PageDown → day offset from the focused day. */
function keyDayOffset(key: string, parts: LocalDatetimeParts): number | null {
  const weekday = new Date(parts.year, parts.month, parts.day).getDay();
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
      return -new Date(parts.year, parts.month, 0).getDate();
    case "PageDown":
      return new Date(parts.year, parts.month + 1, 0).getDate();
    default:
      return null;
  }
}

type DashboardDatetimePickerProps = {
  draft: LocalDatetimeParts;
  onDraftChange: (parts: LocalDatetimeParts) => void;
  onSet: () => void;
  onToday: () => void;
  viewMonth: number;
  viewYear: number;
  onViewMonthChange: (month: number, year: number) => void;
};

/** Calendar, then a time row; changes stay in the draft until Set. */
export function DashboardDatetimePicker({
  draft,
  onDraftChange,
  onSet,
  onToday,
  onViewMonthChange,
  viewMonth,
  viewYear,
}: DashboardDatetimePickerProps) {
  const [today] = useState(() => parseIsoToLocalParts(new Date().toISOString()));
  const timeLabelId = useId();
  const gridRef = useRef<HTMLDivElement>(null);
  const focusAfterMove = useRef(false);
  const { hour12, meridiem } = hour24To12(draft.hour);
  const cells = buildCalendarMonthGrid(viewYear, viewMonth);
  const monthLabel = formatMonthLabel(viewYear, viewMonth);
  const draftKey = calendarDayKey(draft);
  const rovingKey = cells.some((cell) => calendarDayKey(cell) === draftKey)
    ? draftKey
    : calendarDayKey({ year: viewYear, month: viewMonth, day: 1 });

  useEffect(() => {
    if (!focusAfterMove.current) {
      return;
    }
    focusAfterMove.current = false;
    gridRef.current
      ?.querySelector<HTMLButtonElement>(`[data-day="${draftKey}"]`)
      ?.focus();
  }, [draftKey]);

  function shiftMonth(delta: number): void {
    const next = new Date(viewYear, viewMonth + delta, 1);
    onViewMonthChange(next.getMonth(), next.getFullYear());
  }

  function updateDraft(patch: Partial<LocalDatetimeParts>): void {
    onDraftChange({ ...draft, ...patch });
  }

  function handleDayKeyDown(event: KeyboardEvent<HTMLButtonElement>, cell: CalendarCell): void {
    const from = { ...draft, year: cell.year, month: cell.month, day: cell.day };
    const offset = keyDayOffset(event.key, from);
    if (offset === null) {
      return;
    }
    event.preventDefault();
    const next = shiftDraftDays(from, offset);
    focusAfterMove.current = true;
    onDraftChange(next);
    if (next.month !== viewMonth || next.year !== viewYear) {
      onViewMonthChange(next.month, next.year);
    }
  }

  return (
    <div className="dashboard-datetime__panel-body">
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

        <div
          aria-label={monthLabel}
          className="dashboard-datetime__grid"
          ref={gridRef}
          role="group"
        >
          {cells.map((cell) => {
            const key = calendarDayKey(cell);
            const selected = isSameLocalDay(draft, cell);
            const isToday = isSameLocalDay(today, cell);
            return (
              <button
                aria-current={isToday ? "date" : undefined}
                aria-label={formatCalendarDayLabel(cell)}
                aria-pressed={selected}
                className={cn(
                  "dashboard-datetime__day",
                  !cell.inCurrentMonth && "dashboard-datetime__day--outside",
                  isToday && "dashboard-datetime__day--today",
                  selected && "dashboard-datetime__day--selected",
                )}
                data-day={key}
                key={key}
                onClick={() =>
                  updateDraft({ year: cell.year, month: cell.month, day: cell.day })
                }
                onKeyDown={(event) => handleDayKeyDown(event, cell)}
                tabIndex={key === rovingKey ? 0 : -1}
                type="button"
              >
                {cell.day}
              </button>
            );
          })}
        </div>
      </div>

      <div aria-labelledby={timeLabelId} className="dashboard-datetime__time" role="group">
        <span className="dashboard-datetime__time-label" id={timeLabelId}>
          Time
        </span>
        <div className="dashboard-datetime__time-controls">
          <Select
            aria-label="Hour"
            className="dashboard-datetime__time-select"
            contentClassName={DATETIME_SELECT_MENU_CLASS}
            onChange={(event) =>
              updateDraft({ hour: hour12To24(Number(event.target.value), meridiem) })
            }
            value={String(hour12)}
          >
            {buildHour12Options().map((hour) => (
              <option key={hour} value={hour}>
                {String(hour).padStart(2, "0")}
              </option>
            ))}
          </Select>
          <span aria-hidden className="dashboard-datetime__time-separator">
            :
          </span>
          <Select
            aria-label="Minute"
            className="dashboard-datetime__time-select"
            contentClassName={DATETIME_SELECT_MENU_CLASS}
            onChange={(event) => updateDraft({ minute: Number(event.target.value) })}
            value={String(draft.minute)}
          >
            {buildMinuteOptions().map((minute) => (
              <option key={minute} value={minute}>
                {formatMinuteLabel(minute)}
              </option>
            ))}
          </Select>
          <div aria-label="AM or PM" className="dashboard-datetime__period" role="group">
            {(["AM", "PM"] as const).map((option) => (
              <button
                aria-pressed={meridiem === option}
                className={cn(
                  "dashboard-datetime__period-btn",
                  meridiem === option && "dashboard-datetime__period-btn--selected",
                )}
                key={option}
                onClick={() => updateDraft({ hour: hour12To24(hour12, option) })}
                type="button"
              >
                {option}
              </button>
            ))}
          </div>
        </div>
      </div>

      <div className="dashboard-datetime__actions">
        <Button onClick={onToday} type="button" variant="ghost">
          Today
        </Button>
        <Button onClick={onSet} type="button">
          Set
        </Button>
      </div>
    </div>
  );
}
