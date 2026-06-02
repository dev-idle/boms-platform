"use client";

import { useSyncExternalStore } from "react";

import { useAuthStore } from "@/stores/auth-store";

export function useAuthHydrated(): boolean {
  return useSyncExternalStore(
    (onStoreChange) => useAuthStore.persist.onFinishHydration(onStoreChange),
    () => useAuthStore.persist.hasHydrated(),
    () => false,
  );
}
