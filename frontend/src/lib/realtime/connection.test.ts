import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";

import {
  createRealtimeConnection,
  type RealtimeListener,
  type RealtimeSocket,
} from "./connection";
import type { RealtimeTicket } from "./events";

class FakeSocket implements RealtimeSocket {
  onopen: WebSocket["onopen"] = null;
  onmessage: WebSocket["onmessage"] = null;
  onclose: WebSocket["onclose"] = null;
  closedWith: number | undefined;
  closeCalls = 0;

  constructor(readonly url: string) {}

  close(code?: number) {
    this.closedWith = code;
    this.closeCalls += 1;
  }

  open() {
    this.onopen?.call(this as unknown as WebSocket, new Event("open"));
  }

  receive(data: unknown) {
    this.onmessage?.call(this as unknown as WebSocket, { data } as MessageEvent);
  }

  drop(code: number) {
    this.onclose?.call(this as unknown as WebSocket, { code } as CloseEvent);
  }
}

class FatalError extends Error {}

function setup() {
  const sockets: FakeSocket[] = [];
  let issued = 0;
  let failTickets: Error | null = null;
  let refuseSockets = false;
  const requestTicket = vi.fn(async (): Promise<RealtimeTicket> => {
    if (failTickets) {
      throw failTickets;
    }
    issued += 1;
    return { ticket: `ticket-${issued}`, url: "ws://localhost:8081/ws" };
  });
  const connection = createRealtimeConnection({
    requestTicket,
    openSocket: (url) => {
      if (refuseSockets) {
        throw new DOMException("insecure", "SecurityError");
      }
      const socket = new FakeSocket(url);
      sockets.push(socket);
      return socket;
    },
    isFatal: (error) => error instanceof FatalError,
    random: () => 0,
  });
  const listener: RealtimeListener = { onEvent: vi.fn(), onResync: vi.fn() };
  return {
    connection,
    listener,
    requestTicket,
    sockets,
    latest: () => sockets[sockets.length - 1],
    failTickets: (error: Error | null) => {
      failTickets = error;
    },
    refuseSockets: (refuse: boolean) => {
      refuseSockets = refuse;
    },
  };
}

/** Lets pending ticket requests settle. */
const settle = () => vi.advanceTimersByTimeAsync(0);

beforeEach(() => {
  vi.useFakeTimers();
});

afterEach(() => {
  vi.useRealTimers();
});

describe("createRealtimeConnection", () => {
  it("redeems a ticket and goes live, then asks listeners to resync", async () => {
    const t = setup();
    t.connection.retain(t.listener);
    await settle();

    expect(t.latest().url).toBe("ws://localhost:8081/ws?ticket=ticket-1");
    expect(t.connection.getStatus()).toBe("connecting");

    t.latest().open();

    expect(t.connection.getStatus()).toBe("live");
    expect(t.listener.onResync).toHaveBeenCalledTimes(1);
  });

  it("hands readable events to listeners and resyncs on anything else", async () => {
    const t = setup();
    t.connection.retain(t.listener);
    await settle();
    t.latest().open();

    t.latest().receive(
      JSON.stringify({ id: "1", type: "order.created", at: "x", data: { order_id: "o-1" } }),
    );
    expect(t.listener.onEvent).toHaveBeenCalledWith({
      type: "order.created",
      data: { order_id: "o-1" },
    });

    t.latest().receive("not json");
    t.latest().receive(JSON.stringify({ data: {} }));
    expect(t.listener.onResync).toHaveBeenCalledTimes(3);
  });

  it("re-tickets at once when the server asks, staying live", async () => {
    const t = setup();
    t.connection.retain(t.listener);
    await settle();
    t.latest().open();

    t.latest().drop(4001);
    await settle();

    expect(t.connection.getStatus()).toBe("live");
    expect(t.sockets).toHaveLength(2);
    expect(t.latest().url).toContain("ticket=ticket-2");
  });

  it("backs off after a lost connection and resets once live again", async () => {
    const t = setup();
    t.connection.retain(t.listener);
    await settle();
    t.latest().open();

    t.failTickets(new Error("ticket refused"));
    t.latest().drop(1006);
    expect(t.connection.getStatus()).toBe("reconnecting");

    await vi.advanceTimersByTimeAsync(499);
    expect(t.requestTicket).toHaveBeenCalledTimes(1);
    await vi.advanceTimersByTimeAsync(1);
    expect(t.requestTicket).toHaveBeenCalledTimes(2);

    // Second failure waits longer: half of 2 s at the least.
    await vi.advanceTimersByTimeAsync(999);
    expect(t.requestTicket).toHaveBeenCalledTimes(2);
    t.failTickets(null);
    await vi.advanceTimersByTimeAsync(1);
    expect(t.requestTicket).toHaveBeenCalledTimes(3);

    t.latest().open();
    expect(t.connection.getStatus()).toBe("live");
    expect(t.listener.onResync).toHaveBeenCalledTimes(2);

    t.latest().drop(1001);
    await vi.advanceTimersByTimeAsync(500);
    expect(t.requestTicket).toHaveBeenCalledTimes(4);
  });

  it("resume skips the wait", async () => {
    const t = setup();
    t.connection.retain(t.listener);
    await settle();
    t.latest().drop(1006);

    t.connection.resume();
    await settle();

    expect(t.requestTicket).toHaveBeenCalledTimes(2);
  });

  it("shares one socket across listeners and closes it after the last leaves", async () => {
    const t = setup();
    const releaseFirst = t.connection.retain(t.listener);
    const releaseSecond = t.connection.retain({ onEvent: vi.fn(), onResync: vi.fn() });
    await settle();
    t.latest().open();
    expect(t.sockets).toHaveLength(1);

    releaseFirst();
    releaseSecond();
    // A remount within the grace keeps the socket.
    const releaseAgain = t.connection.retain(t.listener);
    await vi.advanceTimersByTimeAsync(5_000);
    expect(t.latest().closedWith).toBeUndefined();

    releaseAgain();
    await vi.advanceTimersByTimeAsync(1_000);
    expect(t.latest().closedWith).toBe(1000);
    expect(t.connection.getStatus()).toBe("idle");

    // Its close does not trigger a reconnect.
    t.latest().drop(1000);
    await vi.advanceTimersByTimeAsync(60_000);
    expect(t.requestTicket).toHaveBeenCalledTimes(1);
  });

  it("opens no socket for a ticket that arrives after the last listener left", async () => {
    const t = setup();
    let resolveTicket: (ticket: RealtimeTicket) => void = () => {};
    t.requestTicket.mockImplementationOnce(
      () => new Promise((resolve) => (resolveTicket = resolve)),
    );
    const release = t.connection.retain(t.listener);
    release();
    await vi.advanceTimersByTimeAsync(1_000);

    resolveTicket({ ticket: "late", url: "ws://localhost:8081/ws" });
    await settle();

    expect(t.sockets).toHaveLength(0);
  });

  it("tells status subscribers about every change", async () => {
    const t = setup();
    const onChange = vi.fn();
    const unsubscribe = t.connection.subscribeStatus(onChange);
    t.connection.retain(t.listener);
    await settle();
    t.latest().open();
    t.latest().drop(1006);

    // idle → connecting → live → reconnecting
    expect(onChange).toHaveBeenCalledTimes(3);
    unsubscribe();
    t.connection.resume();
    await settle();
    t.latest().open();
    expect(onChange).toHaveBeenCalledTimes(3);
  });

  it("stays down after a ticket error no retry can fix, until retained again", async () => {
    const t = setup();
    t.failTickets(new FatalError("signed out"));
    const release = t.connection.retain(t.listener);
    await settle();

    expect(t.connection.getStatus()).toBe("idle");
    await vi.advanceTimersByTimeAsync(60_000);
    expect(t.requestTicket).toHaveBeenCalledTimes(1);

    release();
    t.failTickets(null);
    t.connection.retain(t.listener);
    await settle();
    expect(t.requestTicket).toHaveBeenCalledTimes(2);
    expect(t.sockets).toHaveLength(1);
  });

  it("retries when the browser refuses to open the socket", async () => {
    const t = setup();
    t.refuseSockets(true);
    t.connection.retain(t.listener);
    await settle();

    expect(t.connection.getStatus()).toBe("reconnecting");
    t.refuseSockets(false);
    await vi.advanceTimersByTimeAsync(500);
    expect(t.sockets).toHaveLength(1);
  });

  it("abandons a socket that never opens", async () => {
    const t = setup();
    t.connection.retain(t.listener);
    await settle();

    await vi.advanceTimersByTimeAsync(10_000);

    expect(t.latest().closedWith).toBeUndefined();
    expect(t.latest().closeCalls).toBe(1);
  });
});
