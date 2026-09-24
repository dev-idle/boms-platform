import { EDGE_CLIENT_IP_HEADER } from "@/constants/http-headers";

/** Resolving the visitor's address from what the hosting edge reported. */

const IPV4 = /^(\d{1,3}\.){3}\d{1,3}$/;

function isAddress(value: string): boolean {
  if (value.includes(":")) {
    // Loose on IPv6 shapes; the API parses the value with net.ParseIP and falls
    // back to the socket address when it does not hold up.
    return /^[0-9a-f:.]+$/i.test(value);
  }
  if (!IPV4.test(value)) {
    return false;
  }
  return value
    .split(".")
    .every((part) => part.length <= 3 && Number(part) <= 255);
}

/**
 * The address of the visitor, or undefined when the edge did not report one.
 *
 * The API sees only this proxy, so without a value here every visitor shares the
 * BFF's own address: one rate-limit bucket for the whole storefront, and one
 * address in every audit row.
 *
 * **Only trustworthy behind a proxy that overwrites these headers** — Vercel and
 * a standard nginx `proxy_set_header` both do. A directly exposed Next server
 * would let a visitor choose its own address, so the deployment must not do that.
 */
export function resolveClientIp(headers: Headers): string | undefined {
  const candidates = [
    headers.get(EDGE_CLIENT_IP_HEADER.single),
    headers.get(EDGE_CLIENT_IP_HEADER.chain)?.split(",")[0],
  ];

  for (const candidate of candidates) {
    const trimmed = candidate?.trim();
    if (trimmed && isAddress(trimmed)) {
      return trimmed;
    }
  }

  return undefined;
}
