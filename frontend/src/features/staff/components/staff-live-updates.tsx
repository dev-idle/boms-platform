"use client";

import { useLiveQueries } from "@/lib/realtime";

import {
  staffLiveQueryKeys,
  staffQueryKeysForEvent,
} from "../hooks/query-options";

/** Keeps this tab's staff pages current with the order changes the API pushes. */
export function StaffLiveUpdates() {
  useLiveQueries(staffQueryKeysForEvent, staffLiveQueryKeys);
  return null;
}
