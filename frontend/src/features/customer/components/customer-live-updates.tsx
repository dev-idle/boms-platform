"use client";

import { useLiveQueries } from "@/lib/realtime";

import {
  customerLiveQueryKeys,
  customerQueryKeysForEvent,
} from "../hooks/query-options";

/** Keeps this tab's customer pages current with the order changes the API pushes. */
export function CustomerLiveUpdates() {
  useLiveQueries(customerQueryKeysForEvent, customerLiveQueryKeys);
  return null;
}
