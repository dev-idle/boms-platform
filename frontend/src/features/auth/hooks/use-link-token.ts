"use client";

import { useEffect, useRef, useState } from "react";

import { takeLinkToken } from "../lib/link-token";

/**
 * The token of the emailed link this page was opened from: undefined until
 * read, null when the page has none or a malformed one. The fragment exists
 * only in the browser, so it is read once after mount.
 */
export function useLinkToken(): string | null | undefined {
  const [token, setToken] = useState<string | null | undefined>(undefined);
  // Reading clears the fragment, so a second effect run (Strict Mode) would
  // find it gone.
  const read = useRef(false);

  useEffect(() => {
    if (read.current) {
      return;
    }
    read.current = true;
    setToken(takeLinkToken(window.location, window.history));
  }, []);

  return token;
}
