import { useEffect, useState } from "react";

/**
 * `value` after it stops changing for `delayMs`, and whether it is still
 * catching up. The wait is part of the answer: a field that reports results for
 * the settled value must say it is working until the two agree, or it shows the
 * previous answer — "nothing matches" — for what the user has already typed.
 */
export function useDebouncedValue<T>(
  value: T,
  delayMs: number,
): [debounced: T, settling: boolean] {
  const [debounced, setDebounced] = useState(value);

  useEffect(() => {
    const timerId = window.setTimeout(() => setDebounced(value), delayMs);
    return () => window.clearTimeout(timerId);
  }, [delayMs, value]);

  return [debounced, debounced !== value];
}
