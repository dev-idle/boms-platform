import { useCallback, useState } from "react";

import { useDebouncedValue } from "./use-debounced-value";

const DEFAULT_DELAY_MS = 280;

/** Local input + debounced query for paginated dashboard tables. */
export function useDebouncedTableSearch(delayMs = DEFAULT_DELAY_MS) {
  const [input, setInput] = useState("");
  const [page, setPage] = useState(1);
  const debouncedInput = useDebouncedValue(input, delayMs);
  const search = debouncedInput.trim();

  const [trackedSearch, setTrackedSearch] = useState(search);
  if (search !== trackedSearch) {
    setTrackedSearch(search);
    setPage(1);
  }

  const clear = useCallback(() => {
    setInput("");
  }, []);

  return {
    clear,
    hasInput: input.length > 0,
    input,
    isDebouncing: input.trim() !== search,
    page,
    search,
    setInput,
    setPage,
  };
}
