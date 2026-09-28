"use client";

import { useEffect, useState } from "react";

/**
 * The current time, refreshed every `intervalMs`. Rules measured from now (a
 * minimum notice, a booking window) read it from here so a page left open does
 * not keep showing a time that has since become too soon.
 */
export function useNow(intervalMs: number): Date {
  const [now, setNow] = useState(() => new Date());
  useEffect(() => {
    const timer = setInterval(() => setNow(new Date()), intervalMs);
    return () => clearInterval(timer);
  }, [intervalMs]);
  return now;
}
