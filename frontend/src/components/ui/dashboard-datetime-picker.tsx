"use client";

import {
  DashboardChevronLeftIcon,
  DashboardChevronRightIcon,
} from "@/components/icons/dashboard-ui-icons";
import { Select } from "@/components/ui/select";
import { cn } from "@/lib/utils";
import {
  buildCalendarMonthGrid,
  buildHour12Options,
  buildMinuteOptions,
  CALENDAR_WEEKDAY_LABELS,
  formatMinuteLabel,
  formatMonthLabel,
  hour12To24,
  hour24To12,
  isSameLocalDay,
  type LocalDatetimeParts,
} from "@/lib/validation/datetime-calendar";

const DATETIME_SELECT_MENU_CLASS = "field-select-content--datetime";

type DashboardDatetimePickerProps = {
  draft: LocalDatetimeParts;
  onDraftChange: (parts: LocalDatetimeParts) => void;
  onSet: () => void;
  onToday: () => void;
  viewMonth: number;
  viewYear: number;
  onViewMonthChange: (month: number, year: number) => void;
};

/** Themed calendar + time column; changes stay in draft until Set. */
export function DashboardDatetimePicker({
  draft,
  onDraftChange,
  onSet,
  onToday,
  onViewMonthChange,
  viewMonth,
  viewYear,
}: DashboardDatetimePickerProps) {
  const { hour12, meridiem } = hour24To12(draft.hour);
  const cells = buildCalendarMonthGrid(viewYear, viewMonth);
  const hourOptions = buildHour12Options();
  const minuteOptions = buildMinuteOptions();

  function shiftMonth(delta: number): void {
    const next = new Date(viewYear, viewMonth + delta, 1);
    onViewMonthChange(next.getMonth(), next.getFullYear());
  }

  function updateDraft(patch: Partial<LocalDatetimeParts>): void {
    onDraftChange({ ...draft, ...patch });
  }

  return (
    <div className="dashboard-datetime__panel-body">
      <div className="dashboard-datetime__panel-inner">
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
            <span className="dashboard-datetime__month-label">
              {formatMonthLabel(viewYear, viewMonth)}
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

          <div className="dashboard-datetime__matrix">
            <div aria-hidden className="dashboard-datetime__weekdays">
              {CALENDAR_WEEKDAY_LABELS.map((label) => (
                <span key={label}>{label}</span>
              ))}
            </div>

            <div className="dashboard-datetime__grid" role="grid">
              {cells.map((cell) => {
                const selected = isSameLocalDay(draft, cell);
                return (
                  <button
                    aria-selected={selected}
                    className={cn(
                      "dashboard-datetime__day",
                      !cell.inCurrentMonth && "dashboard-datetime__day--outside",
                      selected && "dashboard-datetime__day--selected",
                    )}
                    key={`${cell.year}-${cell.month}-${cell.day}`}
                    onClick={() =>
                      updateDraft({
                        year: cell.year,
                        month: cell.month,
                        day: cell.day,
                      })
                    }
                    role="gridcell"
                    type="button"
                  >
                    {cell.day}
                  </button>
                );
              })}
            </div>
          </div>
        </div>

        <div className="dashboard-datetime__time-bar">
          <div className="dashboard-datetime__time-header">
            <span className="dashboard-datetime__time-label" id="dashboard-datetime-time">
              Time
            </span>
          </div>
          <div
            aria-labelledby="dashboard-datetime-time"
            className="dashboard-datetime__time-controls"
          >
            <Select
              aria-label="Hour"
              className="dashboard-datetime__time-select"
              contentClassName={DATETIME_SELECT_MENU_CLASS}
              onChange={(event) =>
                updateDraft({
                  hour: hour12To24(Number(event.target.value), meridiem),
                })
              }
              value={String(hour12)}
            >
              {hourOptions.map((hour) => (
                <option key={hour} value={hour}>
                  {String(hour).padStart(2, "0")}
                </option>
              ))}
            </Select>
            <Select
              aria-label="Minute"
              className="dashboard-datetime__time-select"
              contentClassName={DATETIME_SELECT_MENU_CLASS}
              onChange={(event) =>
                updateDraft({ minute: Number(event.target.value) })
              }
              value={String(draft.minute)}
            >
              {minuteOptions.map((minute) => (
                <option key={minute} value={minute}>
                  {formatMinuteLabel(minute)}
                </option>
              ))}
            </Select>
            <div
              className="dashboard-datetime__period-toggle"
              role="group"
              aria-label="AM or PM"
            >
              {(["AM", "PM"] as const).map((option) => (
                <button
                  aria-pressed={meridiem === option}
                  className={cn(
                    "dashboard-datetime__period-btn",
                    meridiem === option && "dashboard-datetime__period-btn--selected",
                  )}
                  key={option}
                  onClick={() =>
                    updateDraft({ hour: hour12To24(hour12, option) })
                  }
                  type="button"
                >
                  {option}
                </button>
              ))}
            </div>
          </div>
        </div>
      </div>

      <div className="dashboard-datetime__panel-actions">
        <button
          className="dashboard-datetime__footer-btn"
          onClick={onToday}
          type="button"
        >
          Today
        </button>
        <button
          className="dashboard-datetime__time-set"
          onClick={onSet}
          type="button"
        >
          Set
        </button>
      </div>
    </div>
  );
}
