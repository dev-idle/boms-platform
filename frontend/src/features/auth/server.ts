/**
 * Server-only auth exports for RSC (app layout).
 * Do not import from `@/features/auth` barrel in Server Components — it re-exports client modules.
 */
export { AuthBootstrap } from "./provider/auth-bootstrap";
