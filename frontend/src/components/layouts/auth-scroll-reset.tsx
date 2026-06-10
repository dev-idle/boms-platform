"use client";

import { useLayoutEffect } from "react";

/** Reset document scroll before paint when entering auth routes. */
export function AuthScrollReset() {
  useLayoutEffect(() => {
    window.scrollTo(0, 0);
  }, []);

  return null;
}
