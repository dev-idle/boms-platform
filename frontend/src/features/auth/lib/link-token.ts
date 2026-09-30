import { linkTokenSchema } from "../schemas";

/**
 * Reads the token an emailed link carries in its fragment (`#token=…`) and
 * takes it out of the address bar, so it does not linger in history or get
 * copied along with the URL. The fragment is never sent to a server.
 */
export function takeLinkToken(location: Pick<Location, "hash" | "pathname" | "search">, history: Pick<History, "replaceState">): string | null {
  const token = new URLSearchParams(location.hash.replace(/^#/, "")).get("token");
  if (location.hash) {
    history.replaceState(null, "", `${location.pathname}${location.search}`);
  }
  return token && linkTokenSchema.safeParse(token).success ? token : null;
}
