/**
 * Data access layer (server outbound adapters → Fiber).
 * Only import from RSC, Server Actions, or route handlers — never from Client Components.
 */
export { dalFiberHealth } from "./health.dal";
