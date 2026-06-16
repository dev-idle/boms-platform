export type LocalDatetimeParts = {
  day: number;
  hour: number;
  minute: number;
  month: number;
  year: number;
};

export type CalendarCell = {
  day: number;
  inCurrentMonth: boolean;
  month: number;
  year: number;
};

const CALENDAR_CELL_COUNT = 42;

export function parseIsoToLocalParts(iso: string): LocalDatetimeParts {
  const date = new Date(iso);
  return {
    year: date.getFullYear(),
    month: date.getMonth(),
    day: date.getDate(),
    hour: date.getHours(),
    minute: date.getMinutes(),
  };
}

export function partsToIso(parts: LocalDatetimeParts): string {
  return new Date(
    parts.year,
    parts.month,
    parts.day,
    parts.hour,
    parts.minute,
    0,
    0,
  ).toISOString();
}

export function buildCalendarMonthGrid(
  year: number,
  month: number,
): CalendarCell[] {
  const firstDay = new Date(year, month, 1);
  const startWeekday = firstDay.getDay();
  const daysInMonth = new Date(year, month + 1, 0).getDate();
  const daysInPrevMonth = new Date(year, month, 0).getDate();
  const cells: CalendarCell[] = [];

  for (let index = startWeekday - 1; index >= 0; index -= 1) {
    const day = daysInPrevMonth - index;
    const prevMonth = month === 0 ? 11 : month - 1;
    const prevYear = month === 0 ? year - 1 : year;
    cells.push({
      year: prevYear,
      month: prevMonth,
      day,
      inCurrentMonth: false,
    });
  }

  for (let day = 1; day <= daysInMonth; day += 1) {
    cells.push({ year, month, day, inCurrentMonth: true });
  }

  let nextDay = 1;
  while (cells.length < CALENDAR_CELL_COUNT) {
    const nextMonth = month === 11 ? 0 : month + 1;
    const nextYear = month === 11 ? year + 1 : year;
    cells.push({
      year: nextYear,
      month: nextMonth,
      day: nextDay,
      inCurrentMonth: false,
    });
    nextDay += 1;
  }

  return cells;
}

export function isSameLocalDay(
  parts: LocalDatetimeParts,
  cell: CalendarCell,
): boolean {
  return (
    parts.year === cell.year &&
    parts.month === cell.month &&
    parts.day === cell.day
  );
}

export function formatDashboardDatetime(iso: string): string {
  return new Intl.DateTimeFormat("en-US", {
    month: "short",
    day: "numeric",
    year: "numeric",
    hour: "numeric",
    minute: "2-digit",
  }).format(new Date(iso));
}

export function formatMonthLabel(year: number, month: number): string {
  return new Intl.DateTimeFormat("en-US", {
    month: "long",
    year: "numeric",
  }).format(new Date(year, month, 1));
}

export const CALENDAR_WEEKDAY_LABELS = [
  "Su",
  "Mo",
  "Tu",
  "We",
  "Th",
  "Fr",
  "Sa",
] as const;

export function buildMinuteOptions(): number[] {
  return Array.from({ length: 60 }, (_, minute) => minute);
}

export type Meridiem = "AM" | "PM";

export function hour24To12(hour24: number): { hour12: number; meridiem: Meridiem } {
  const meridiem: Meridiem = hour24 >= 12 ? "PM" : "AM";
  const hour12 = hour24 % 12 || 12;
  return { hour12, meridiem };
}

export function hour12To24(hour12: number, meridiem: Meridiem): number {
  if (meridiem === "AM") {
    return hour12 === 12 ? 0 : hour12;
  }
  return hour12 === 12 ? 12 : hour12 + 12;
}

export function buildHour12Options(): number[] {
  return Array.from({ length: 12 }, (_, index) => index + 1);
}

export function formatMinuteLabel(minute: number): string {
  return String(minute).padStart(2, "0");
}
