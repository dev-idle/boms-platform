"use client";

import type { RefObject } from "react";
import { useEffect } from "react";

type UseStorefrontHeaderPanelOptions = {
  open: boolean;
  onClose: () => void;
  panelRef: RefObject<HTMLDivElement | null>;
  toggleRef: RefObject<HTMLButtonElement | null>;
  /** Focused when the panel opens; omit for panels that are only links. */
  focusRef?: RefObject<HTMLElement | null>;
  /** Second element that counts as inside the panel, e.g. the suggestion list. */
  satelliteRef?: RefObject<HTMLElement | null>;
};

/** Escape, outside-click and open-focus behaviour shared by the header panels. */
export function useStorefrontHeaderPanel({
  open,
  onClose,
  panelRef,
  toggleRef,
  focusRef,
  satelliteRef,
}: UseStorefrontHeaderPanelOptions) {
  useEffect(() => {
    if (!open) {
      return;
    }

    const frame = window.requestAnimationFrame(() => {
      focusRef?.current?.focus();
    });

    const onKeyDown = (event: KeyboardEvent) => {
      if (event.key === "Escape") {
        onClose();
      }
    };

    document.addEventListener("keydown", onKeyDown);
    return () => {
      window.cancelAnimationFrame(frame);
      document.removeEventListener("keydown", onKeyDown);
    };
  }, [open, onClose, focusRef]);

  useEffect(() => {
    if (!open) {
      return;
    }

    const onPointerDown = (event: PointerEvent) => {
      const target = event.target;
      if (!(target instanceof Node)) {
        return;
      }

      if (panelRef.current?.contains(target)) {
        return;
      }

      if (toggleRef.current?.contains(target)) {
        return;
      }

      if (satelliteRef?.current?.contains(target)) {
        return;
      }

      onClose();
    };

    const timer = window.setTimeout(() => {
      document.addEventListener("pointerdown", onPointerDown);
    }, 0);

    return () => {
      window.clearTimeout(timer);
      document.removeEventListener("pointerdown", onPointerDown);
    };
  }, [open, onClose, panelRef, satelliteRef, toggleRef]);
}
