"use client";

import { useCallback, useState } from "react";

/**
 * Keeps a saved form snapshot and bumps a remount key after server success so
 * react-hook-form gets a fresh instance (avoids `reset()` breaking controlled inputs).
 */
export function useRemountingFormSnapshot<T>(initialSnapshot: T) {
  const [formKey, setFormKey] = useState(0);
  const [snapshot, setSnapshot] = useState(initialSnapshot);

  const commitSnapshot = useCallback((next: T) => {
    setSnapshot(next);
    setFormKey((current) => current + 1);
  }, []);

  return { commitSnapshot, formKey, snapshot };
}
