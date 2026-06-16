"use client";

import {
  useCallback,
  useLayoutEffect,
  useRef,
  type RefObject,
} from "react";

import {
  clampSelection,
  rememberTrustedInputSelection,
  restoreInputSelection,
  type TextSelection,
} from "@/lib/input-selection";

type UseInputSelectionRestoreOptions = {
  /** When false, the hook is a no-op (CSS-mask path). */
  enabled: boolean;
  inputRef: RefObject<HTMLInputElement | null>;
  /** Changing this value restores the last captured selection (e.g. visibility toggle). */
  restoreKey: unknown;
};

type UseInputSelectionRestoreResult = {
  captureSelectionForToggle: () => void;
  rememberSelection: (input: HTMLInputElement) => void;
};

export function useInputSelectionRestore({
  enabled,
  inputRef,
  restoreKey,
}: UseInputSelectionRestoreOptions): UseInputSelectionRestoreResult {
  const selectionRef = useRef<TextSelection>({ end: 0, start: 0 });
  const pendingRestoreRef = useRef<TextSelection | null>(null);

  const rememberSelection = useCallback((input: HTMLInputElement) => {
    selectionRef.current = rememberTrustedInputSelection(
      input,
      selectionRef.current,
    );
  }, []);

  const captureSelectionForToggle = useCallback(() => {
    const input = inputRef.current;
    if (!input) {
      return;
    }
    rememberSelection(input);
    pendingRestoreRef.current = clampSelection(
      selectionRef.current,
      input.value.length,
    );
  }, [inputRef, rememberSelection]);

  useLayoutEffect(() => {
    if (!enabled) {
      return;
    }

    const pending = pendingRestoreRef.current;
    const input = inputRef.current;
    if (!pending || !input) {
      return;
    }

    const applyRestore = (): void => {
      const currentInput = inputRef.current;
      if (!currentInput) {
        return;
      }
      selectionRef.current = restoreInputSelection(currentInput, pending);
    };

    applyRestore();
    const frameOne = requestAnimationFrame(() => {
      applyRestore();
      requestAnimationFrame(applyRestore);
    });

    pendingRestoreRef.current = null;

    return () => {
      cancelAnimationFrame(frameOne);
    };
  }, [enabled, inputRef, restoreKey]);

  return { captureSelectionForToggle, rememberSelection };
}
