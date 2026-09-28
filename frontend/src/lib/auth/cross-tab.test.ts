import { afterEach, describe, expect, it, vi } from "vitest";

import {
  accessTokenSubject,
  createTokenChannel,
  withRefreshLock,
  type SharedToken,
  type TokenChannel,
} from "./cross-tab";

const USER_ID = "7c9e6679-7425-40de-944b-e07fc1f90ae7";

function unsignedToken(claims: object): string {
  const encode = (value: object) =>
    btoa(JSON.stringify(value)).replace(/\+/g, "-").replace(/\//g, "_").replace(/=+$/, "");
  return `${encode({ alg: "EdDSA" })}.${encode(claims)}.signature`;
}

describe("accessTokenSubject", () => {
  it("reads the account a token was issued to", () => {
    expect(accessTokenSubject(unsignedToken({ sub: USER_ID, role: "staff" }))).toBe(USER_ID);
  });

  it("returns null for anything that is not a token with a user subject", () => {
    expect(accessTokenSubject("")).toBeNull();
    expect(accessTokenSubject("only-one-part")).toBeNull();
    expect(accessTokenSubject("a.%%%.c")).toBeNull();
    expect(accessTokenSubject(unsignedToken({ sub: "not-a-uuid" }))).toBeNull();
  });
});

describe("withRefreshLock", () => {
  it("never lets two refreshes overlap", async () => {
    const order: string[] = [];
    let releaseFirst: () => void = () => {};
    const first = withRefreshLock(async () => {
      order.push("first:start");
      await new Promise<void>((resolve) => (releaseFirst = resolve));
      order.push("first:end");
    });
    const second = withRefreshLock(async () => {
      order.push("second:start");
    });

    await vi.waitFor(() => expect(order).toEqual(["first:start"]));
    releaseFirst();
    await Promise.all([first, second]);

    expect(order).toEqual(["first:start", "first:end", "second:start"]);
  });
});

describe("createTokenChannel", () => {
  const channels: TokenChannel[] = [];

  afterEach(() => {
    for (const channel of channels.splice(0)) {
      channel.close();
    }
  });

  function tab(onToken: (token: SharedToken) => void = () => {}): TokenChannel {
    const channel = createTokenChannel(onToken);
    if (channel === null) {
      throw new Error("BroadcastChannel is unavailable");
    }
    channels.push(channel);
    return channel;
  }

  it("hands a shared token to the other tabs", async () => {
    const received = vi.fn();
    const sender = tab();
    tab(received);
    const token = { access_token: unsignedToken({ sub: USER_ID }), expires_at: 1_900_000_000_000 };

    sender.share(token);

    await vi.waitFor(() => expect(received).toHaveBeenCalledWith(token));
  });

  it("ignores messages of any other shape", async () => {
    const received = vi.fn();
    tab(received);
    const stranger = new BroadcastChannel("boms:auth-token");

    stranger.postMessage({ access_token: "", expires_at: -1 });
    stranger.postMessage("hello");
    const probe = tab();
    probe.share({ access_token: "valid", expires_at: 1 });

    await vi.waitFor(() => expect(received).toHaveBeenCalledTimes(1));
    expect(received).toHaveBeenCalledWith({ access_token: "valid", expires_at: 1 });
    stranger.close();
  });
});
