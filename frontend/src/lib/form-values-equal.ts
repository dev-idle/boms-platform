type ShallowEqualOptions<T extends Record<string, unknown>> = {
  /** Treat `null` and `""` as equivalent for these keys (optional nullable form fields). */
  nullableKeys?: (keyof T)[];
};

/** Shallow strict equality for small react-hook-form value objects. */
export function shallowFormValuesEqual<T extends Record<string, unknown>>(
  current: T,
  baseline: T,
  options?: ShallowEqualOptions<T>,
): boolean {
  const nullableKeys = new Set(options?.nullableKeys ?? []);
  const keys = Object.keys(baseline) as (keyof T)[];

  return keys.every((key) => {
    if (nullableKeys.has(key)) {
      return (current[key] ?? "") === (baseline[key] ?? "");
    }
    return current[key] === baseline[key];
  });
}
