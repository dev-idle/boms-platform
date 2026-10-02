"use client";

import { useLiveQueries } from "@/lib/realtime";

import {
  managerLiveQueryKeys,
  managerQueryKeysForEvent,
} from "../hooks/query-options";

/** Keeps this tab's manager pages current with the reviews the API pushes. */
export function ManagerLiveUpdates() {
  useLiveQueries(managerQueryKeysForEvent, managerLiveQueryKeys);
  return null;
}
