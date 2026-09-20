/** Baker production page eyebrow — shift label plus calendar date. */
export function formatBakerShiftEyebrow(date: Date): string {
  const hour = date.getHours();
  const shift =
    hour < 12 ? "Morning shift" : hour < 17 ? "Afternoon shift" : "Evening shift";
  const formatted = date.toLocaleDateString("en-US", {
    weekday: "short",
    day: "numeric",
    month: "short",
  });

  return `${shift} · ${formatted}`;
}
