import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";

import type { Me } from "@/lib/schemas/me";
import { useAuthStore } from "@/stores/auth-store";

import { createTokenChannel, type SharedToken, type TokenChannel } from "./cross-tab";
import { refreshNow, resetRefreshManager, scheduleRefresh } from "./refresh-manager";

const USER_ID = "7c9e6679-7425-40de-944b-e07fc1f90ae7";
const OTHER_USER_ID = "16fd2706-8baf-433b-82eb-8c7fada847da";
const MINUTE = 60_000;

function unsignedToken(sub: string, label: string): string {
  const encode = (value: object) =>
    btoa(JSON.stringify(value)).replace(/\+/g, "-").replace(/\//g, "_").replace(/=+$/, "");
  return `${encode({ alg: "EdDSA" })}.${encode({ sub, jti: label })}.signature`;
}

const T0 = unsignedToken(USER_ID, "t0");

/** Every access token this tab's store has held since the test began. */
let held: string[] = [];
let unsubscribe: () => void = () => {};
/** Another tab of the same site. */
let otherTab: TokenChannel;

function share(token: SharedToken) {
  otherTab.share(token);
}

async function waitForToken(token: string) {
  await vi.waitFor(() => expect(useAuthStore.getState().accessToken).toBe(token));
}

beforeEach(() => {
  const expiresAt = Date.now() + 15 * MINUTE;
  useAuthStore.setState({
    status: "authenticated",
    user: { id: USER_ID, must_change_password: false } as Me,
    accessToken: T0,
    expiresAt,
  });
  held = [];
  unsubscribe = useAuthStore.subscribe((state) => {
    if (state.accessToken !== null && held.at(-1) !== state.accessToken) {
      held.push(state.accessToken);
    }
  });
  // Opens this tab's side of the channel, as a signed-in tab does.
  scheduleRefresh(expiresAt);
  const channel = createTokenChannel(() => {});
  if (channel === null) {
    throw new Error("BroadcastChannel is unavailable");
  }
  otherTab = channel;
});

afterEach(() => {
  unsubscribe();
  otherTab.close();
  resetRefreshManager();
  vi.unstubAllGlobals();
});

describe("tokens shared by other tabs", () => {
  it("adopts a newer token for this tab's account", async () => {
    const t1 = unsignedToken(USER_ID, "t1");
    share({ access_token: t1, expires_at: Date.now() + 30 * MINUTE });

    await waitForToken(t1);
  });

  it("ignores another account's token and an older one", async () => {
    const foreign = unsignedToken(OTHER_USER_ID, "other");
    const older = unsignedToken(USER_ID, "older");
    const newest = unsignedToken(USER_ID, "newest");
    share({ access_token: foreign, expires_at: Date.now() + 30 * MINUTE });
    share({ access_token: older, expires_at: Date.now() + MINUTE });
    share({ access_token: newest, expires_at: Date.now() + 30 * MINUTE });

    await waitForToken(newest);
    expect(held).toEqual([newest]);
  });

  it("ignores tokens while signed out", async () => {
    useAuthStore.setState({ status: "unauthenticated", user: null, accessToken: null, expiresAt: null });
    share({ access_token: unsignedToken(USER_ID, "t1"), expires_at: Date.now() + 30 * MINUTE });

    await new Promise((resolve) => setTimeout(resolve, 50));
    expect(useAuthStore.getState().accessToken).toBeNull();
  });
});

describe("refreshNow", () => {
  const refreshResponse = (token: string) =>
    new Response(
      JSON.stringify({
        success: true,
        data: { access_token: token, token_type: "Bearer", expires_in: 900 },
      }),
      { status: 200, headers: { "Content-Type": "application/json" } },
    );

  it("skips its refresh when another tab refreshed while it waited", async () => {
    const fetchMock = vi.fn();
    vi.stubGlobal("fetch", fetchMock);
    let releaseLock: () => void = () => {};
    const lockHeld = new Promise<void>((resolve) => (releaseLock = resolve));
    let lockTaken: () => void = () => {};
    const taken = new Promise<void>((resolve) => (lockTaken = resolve));
    const otherRefresh = navigator.locks.request("boms:auth-refresh", () => {
      lockTaken();
      return lockHeld;
    });
    await taken;

    const refreshing = refreshNow({ staleToken: T0 });
    const t1 = unsignedToken(USER_ID, "t1");
    share({ access_token: t1, expires_at: Date.now() + 30 * MINUTE });
    releaseLock();
    await otherRefresh;
    await refreshing;

    expect(fetchMock).not.toHaveBeenCalled();
    expect(useAuthStore.getState().accessToken).toBe(t1);
  });

  it("refreshes a token nobody replaced and hands it to the other tabs", async () => {
    const t2 = unsignedToken(USER_ID, "t2");
    const fetchMock = vi.fn(async () => refreshResponse(t2));
    vi.stubGlobal("fetch", fetchMock);
    const received = vi.fn();
    const listener = createTokenChannel(received);

    await refreshNow({ staleToken: T0 });

    expect(fetchMock).toHaveBeenCalledTimes(1);
    expect(useAuthStore.getState().accessToken).toBe(t2);
    await vi.waitFor(() =>
      expect(received).toHaveBeenCalledWith(expect.objectContaining({ access_token: t2 })),
    );
    listener?.close();
  });
});
