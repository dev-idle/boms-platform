/**
 * Headers the browser must never set. The hosting edge writes the forwarding
 * headers and the BFF re-stamps them: a visitor that could set them would pick
 * its own rate-limit bucket at the API, and the address written into audit rows.
 */
/**
 * Where the hosting edge reports the visitor: one address, or a chain whose
 * first entry is the client it saw. Read through these names only — they are the
 * same strings the strip list below removes, so a header the BFF reads can never
 * be one a browser was allowed to set.
 */
export const EDGE_CLIENT_IP_HEADER = {
  single: "x-real-ip",
  chain: "x-forwarded-for",
} as const;

export const FORWARDING_HEADERS = [
  "forwarded",
  "x-client-ip",
  EDGE_CLIENT_IP_HEADER.chain,
  EDGE_CLIENT_IP_HEADER.single,
] as const;

/** The address of the visitor, as the BFF reports it to the API. */
export const CLIENT_IP_HEADER = "X-Client-IP";
