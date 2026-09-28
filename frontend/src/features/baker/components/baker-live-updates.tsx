"use client";

import { useLiveQueries } from "@/lib/realtime";

import {
  bakerLiveQueryKeys,
  bakerQueryKeysForEvent,
} from "../hooks/query-options";

/** Keeps this tab's baker pages current with the order changes the API pushes. */
export function BakerLiveUpdates() {
  useLiveQueries(bakerQueryKeysForEvent, bakerLiveQueryKeys);
  return null;
}
