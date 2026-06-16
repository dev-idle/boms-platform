"use client";

import { useLayoutEffect, useState } from "react";

import {
  resolvePasswordMaskStrategy,
  type PasswordMaskStrategy,
} from "@/lib/password-mask";

/**
 * Resolves after hydration in a layout effect — keeps SSR and the first client render
 * on `pending` (native password type), then switches before paint on Chromium/WebKit.
 */
export function usePasswordMaskStrategy(): PasswordMaskStrategy | "pending" {
  const [strategy, setStrategy] = useState<PasswordMaskStrategy | "pending">(
    "pending",
  );

  useLayoutEffect(() => {
    // One-shot feature detect; layout effect avoids a visible type flash.
    // eslint-disable-next-line react-hooks/set-state-in-effect -- intentional pre-paint strategy resolve
    setStrategy(resolvePasswordMaskStrategy());
  }, []);

  return strategy;
}
