/** HTML attribute for light/dark overrides in `globals.css`. */
export const COLOR_MODE_ATTR = "data-mode";

export type ColorMode = "light" | "dark";

/**
 * Blocking init script (root layout, `beforeInteractive`).
 * Avoids FOUC before React hydrates; no user-controlled input.
 */
export const COLOR_MODE_INIT_SCRIPT = `(function(){try{var m=window.matchMedia("(prefers-color-scheme: dark)").matches?"dark":"light";document.documentElement.setAttribute("${COLOR_MODE_ATTR}",m)}catch(e){}})();`;
