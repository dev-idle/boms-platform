"use client";

import { useQueryClient, type QueryKey } from "@tanstack/react-query";
import { useEffect, useEffectEvent, useSyncExternalStore } from "react";

import { browserRequest } from "@/lib/browser-api-client";
import { isApiError } from "@/lib/errors";
import { useAuthStore } from "@/stores/auth-store";

import {
  createRealtimeConnection,
  type RealtimeConnection,
  type RealtimeListener,
  type RealtimeStatus,
} from "./connection";
import {
  realtimeTicketSchema,
  type RealtimeEvent,
  type RealtimeTicket,
} from "./events";

export type { RealtimeStatus } from "./connection";

const TICKET_TIMEOUT_MS = 10_000;

let connection: RealtimeConnection | null = null;

/** The tab's one socket, created on first use. */
function tabConnection(): RealtimeConnection {
  if (connection !== null) {
    return connection;
  }
  const created = createRealtimeConnection({
    requestTicket: (signal) =>
      browserRequest<RealtimeTicket>("/api/v1/realtime/tickets", {
        method: "POST",
        schema: realtimeTicketSchema,
        signal: AbortSignal.any([signal, AbortSignal.timeout(TICKET_TIMEOUT_MS)]),
      }),
    openSocket: (url) => new WebSocket(url),
    // Signed out, or not allowed a ticket (a password change is due): retrying
    // cannot help until the tab's session changes and the page remounts.
    isFatal: (error) =>
      isApiError(error) && (error.isAuthError() || error.isForbidden()),
  });
  // Back online or back on the tab: reconnect now instead of waiting out the backoff.
  window.addEventListener("online", created.resume);
  document.addEventListener("visibilitychange", () => {
    if (document.visibilityState === "visible") {
      created.resume();
    }
  });
  connection = created;
  return created;
}

function useRealtimeListener(listener: RealtimeListener): void {
  const onEvent = useEffectEvent(listener.onEvent);
  const onResync = useEffectEvent(listener.onResync);
  // The API issues no ticket until a required password change is done.
  const passwordChangeDue = useAuthStore(
    (state) => state.user?.must_change_password === true,
  );
  useEffect(() => {
    if (passwordChangeDue) {
      return;
    }
    return tabConnection().retain({
      onEvent: (event) => onEvent(event),
      onResync: () => onResync(),
    });
  }, [passwordChangeDue]);
}

/**
 * Keeps the tab's socket open while mounted and marks stale the queries each
 * pushed event names; after any gap in the connection, all of `resyncKeys`.
 */
export function useLiveQueries(
  keysForEvent: (event: RealtimeEvent) => readonly QueryKey[],
  resyncKeys: readonly QueryKey[],
): void {
  const queryClient = useQueryClient();
  const invalidate = (keys: readonly QueryKey[]) => {
    for (const queryKey of keys) {
      void queryClient.invalidateQueries({ queryKey });
    }
  };
  useRealtimeListener({
    onEvent: (event) => invalidate(keysForEvent(event)),
    onResync: () => invalidate(resyncKeys),
  });
}

export function useRealtimeStatus(): RealtimeStatus {
  return useSyncExternalStore(
    (onChange) => tabConnection().subscribeStatus(onChange),
    () => tabConnection().getStatus(),
    () => "idle",
  );
}
