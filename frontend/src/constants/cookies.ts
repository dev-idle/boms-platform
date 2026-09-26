/** HttpOnly refresh session cookie issued by Fiber (name configurable on backend). */
export const AUTH_REFRESH_COOKIE = "boms_refresh";

/**
 * The signed-in role, issued beside the session cookie (backend COOKIE_ROLE_NAME).
 * The proxy routes on it so a returning visitor never watches a page they are not
 * allowed to keep. A navigation hint only — never read as authorization.
 */
export const AUTH_ROLE_COOKIE = "boms_role";
