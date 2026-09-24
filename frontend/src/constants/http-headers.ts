/**
 * Headers the browser must never set. The hosting edge writes the forwarding
 * headers and the BFF re-stamps them: a visitor that could set them would pick
 * its own rate-limit bucket at the API, and the address written into audit rows.
 */
export const FORWARDING_HEADERS = [
  "forwarded",
  "x-client-ip",
  "x-forwarded-for",
  "x-real-ip",
] as const;

/** The address of the visitor, as the BFF reports it to the API. */
export const CLIENT_IP_HEADER = "X-Client-IP";
