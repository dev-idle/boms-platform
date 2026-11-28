"use client";

import { useId } from "react";

import { Button } from "@/components/ui/button";
import { CalendarMonth } from "@/components/ui/calendar-month";
import { Select } from "@/components/ui/select";
import { cn } from "@/lib/utils";
import {
  buildHour12Options,
  buildMinuteOptions,
  formatMinuteLabel,
  hour12To24,
  hour24To12,
  type CalendarDay,
  type LocalDatetimeParts,
} from "@/lib/validation/datetime-calendar";

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
  const timeLabelId = useId();
  const { hour12, meridiem } = hour24To12(draft.hour);

  function updateDraft(patch: Partial<LocalDatetimeParts>): void {
    onDraftChange({ ...draft, ...patch });
  }

  // A pressed day and an arrow key both move the draft: the day follows the focus.
  function setDay(day: CalendarDay): void {
    updateDraft(day);
  }

  return (
    <div className="dashboard-datetime__panel-body">
      <CalendarMonth
        cursor={draft}
        onCursorChange={setDay}
        onPick={setDay}
        onViewMonthChange={onViewMonthChange}
        selected={draft}
        viewMonth={viewMonth}
        viewYear={viewYear}
      />

      <div aria-labelledby={timeLabelId} className="dashboard-datetime__time" role="group">
        <span className="dashboard-datetime__time-label" id={timeLabelId}>
          Time
        </span>
        <div className="dashboard-datetime__time-controls">
          <Select
            aria-label="Hour"
            className="dashboard-datetime__time-select"
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
