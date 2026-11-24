"use client";

import { USER_ROLE } from "@/constants/roles";
import { useLiveQueries } from "@/lib/realtime";
import { useAuthStore } from "@/stores/auth-store";

import {
  customerLiveQueryKeys,
  customerQueryKeysForEvent,
} from "../hooks/query-options";

/**
 * Keeps this tab's storefront current with the changes the API pushes to a
 * signed-in customer — their cart, orders and messages. A guest has no
 * channel to listen on.
 */
export function CustomerLiveUpdates() {
  const customer = useAuthStore(
    (state) => state.status === "authenticated" && state.user?.role === USER_ROLE.customer,
  );
  return customer ? <CustomerLiveQueries /> : null;
}

function CustomerLiveQueries() {
  useLiveQueries(customerQueryKeysForEvent, customerLiveQueryKeys);
  return null;
}
