"use client";

import { useEffect } from "react";

import {
  COLOR_MODE_ATTR,
  type ColorMode,
} from "@/lib/theme/color-mode";

function readColorMode(): ColorMode {
  return window.matchMedia("(prefers-color-scheme: dark)").matches
    ? "dark"
    : "light";
}

/** Keeps `data-mode` in sync when the OS color scheme changes. */
export function ColorModeListener() {
  useEffect(() => {
    const media = window.matchMedia("(prefers-color-scheme: dark)");

    function apply() {
      document.documentElement.setAttribute(COLOR_MODE_ATTR, readColorMode());
    }

    media.addEventListener("change", apply);
    return () => media.removeEventListener("change", apply);
  }, []);

  return null;
}
