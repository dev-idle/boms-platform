import { reconnectDelay } from "./backoff";
import {
  realtimeEventSchema,
  type RealtimeEvent,
  type RealtimeTicket,
} from "./events";

/**
 * What the tab tells the user. "idle" means nothing asked for pushes, or the
 * server refused this tab for good; a planned re-ticket keeps showing "live".
 */
export type RealtimeStatus = "idle" | "connecting" | "live" | "reconnecting";

export type RealtimeListener = {
  onEvent: (event: RealtimeEvent) => void;
  /** Events may have been missed: refetch everything this listener shows. */
  onResync: () => void;
};

export type RealtimeSocket = Pick<
  WebSocket,
  "onopen" | "onmessage" | "onclose" | "close"
>;

type RealtimeConnectionDeps = {
  requestTicket: (signal: AbortSignal) => Promise<RealtimeTicket>;
  openSocket: (url: string) => RealtimeSocket;
  /** A ticket error no retry can fix, e.g. a session that is gone. */
  isFatal: (error: unknown) => boolean;
  random?: () => number;
};

export type RealtimeConnection = {
  /** Keeps the socket open until the returned release is called. */
  retain: (listener: RealtimeListener) => () => void;
  /** Skips a pending reconnect wait, e.g. when the tab comes back. */
  resume: () => void;
  getStatus: () => RealtimeStatus;
  subscribeStatus: (onChange: () => void) => () => void;
};

/** The server wants a new ticket: session rotated or ended, or lifetime reached. */
const CLOSE_REAUTHENTICATE = 4001;
const CLOSE_NORMAL = 1000;
/** Keeps the socket through a remount instead of reopening it. */
const RELEASE_GRACE_MS = 1_000;
/** A socket that has not opened by then is abandoned and retried. */
const OPEN_TIMEOUT_MS = 10_000;

function parseEvent(data: unknown): RealtimeEvent | null {
  if (typeof data !== "string") {
    return null;
  }
  let json: unknown;
  try {
    json = JSON.parse(data);
  } catch {
    return null;
  }
  const parsed = realtimeEventSchema.safeParse(json);
  return parsed.success ? parsed.data : null;
}

/**
 * One push socket shared by every listener in the tab. It opens on the first
 * listener, closes shortly after the last one leaves, and reconnects on its own:
 * at once with a fresh ticket when the server asks for one, otherwise after a
 * growing, jittered wait.
 */
export function createRealtimeConnection({
  requestTicket,
  openSocket,
  isFatal,
  random = Math.random,
}: RealtimeConnectionDeps): RealtimeConnection {
  const listeners = new Set<RealtimeListener>();
  const statusListeners = new Set<() => void>();
  let status: RealtimeStatus = "idle";
  let running = false;
  let attempt = 0;
  let socket: RealtimeSocket | null = null;
  let ticketRequest: AbortController | null = null;
  let retryTimer: ReturnType<typeof setTimeout> | null = null;
  let releaseTimer: ReturnType<typeof setTimeout> | null = null;
  let openTimer: ReturnType<typeof setTimeout> | null = null;

  function setStatus(next: RealtimeStatus) {
    if (status === next) {
      return;
    }
    status = next;
    for (const onChange of statusListeners) {
      onChange();
    }
  }

  function resyncAll() {
    for (const listener of listeners) {
      listener.onResync();
    }
  }

  function handleMessage(data: unknown) {
    const event = parseEvent(data);
    if (event === null) {
      // A message this tab cannot read still says something changed.
      resyncAll();
      return;
    }
    for (const listener of listeners) {
      listener.onEvent(event);
    }
  }

  function scheduleRetry() {
    setStatus("reconnecting");
    retryTimer = setTimeout(() => void connect(), reconnectDelay(attempt, random));
    attempt += 1;
  }

  function clearOpenTimer() {
    if (openTimer !== null) {
      clearTimeout(openTimer);
      openTimer = null;
    }
  }

  function open(ticket: RealtimeTicket) {
    const url = new URL(ticket.url);
    url.searchParams.set("ticket", ticket.ticket);
    let ws: RealtimeSocket;
    try {
      ws = openSocket(url.toString());
    } catch {
      // The browser refused the URL outright (e.g. ws:// from an https page).
      scheduleRetry();
      return;
    }
    socket = ws;
    // Closing a socket that never opened fires onclose, which retries.
    openTimer = setTimeout(() => ws.close(), OPEN_TIMEOUT_MS);
    ws.onopen = () => {
      if (socket !== ws) {
        return;
      }
      clearOpenTimer();
      attempt = 0;
      setStatus("live");
      // What changed while no socket was open is only known by asking again.
      resyncAll();
    };
    ws.onmessage = (message) => {
      if (socket === ws) {
        handleMessage(message.data);
      }
    };
    ws.onclose = (event) => {
      if (socket !== ws) {
        return;
      }
      clearOpenTimer();
      socket = null;
      if (event.code === CLOSE_REAUTHENTICATE) {
        void connect();
        return;
      }
      scheduleRetry();
    };
  }

  async function connect(): Promise<void> {
    retryTimer = null;
    const request = new AbortController();
    ticketRequest = request;
    let ticket: RealtimeTicket;
    try {
      ticket = await requestTicket(request.signal);
    } catch (error) {
      if (request.signal.aborted) {
        return;
      }
      if (isFatal(error)) {
        // Stays down until a listener retains the socket again.
        halt();
        return;
      }
      scheduleRetry();
      return;
    } finally {
      if (ticketRequest === request) {
        ticketRequest = null;
      }
    }
    if (!request.signal.aborted) {
      open(ticket);
    }
  }

  function start() {
    running = true;
    attempt = 0;
    setStatus("connecting");
    void connect();
  }

  /** Drops the socket and every pending attempt, keeping the listeners. */
  function halt() {
    running = false;
    ticketRequest?.abort();
    ticketRequest = null;
    if (retryTimer !== null) {
      clearTimeout(retryTimer);
      retryTimer = null;
    }
    clearOpenTimer();
    const ws = socket;
    socket = null;
    ws?.close(CLOSE_NORMAL);
    setStatus("idle");
  }

  function stop() {
    releaseTimer = null;
    halt();
  }

  return {
    retain(listener) {
      listeners.add(listener);
      if (releaseTimer !== null) {
        clearTimeout(releaseTimer);
        releaseTimer = null;
      }
      if (!running) {
        start();
      }
      return () => {
        listeners.delete(listener);
        if (listeners.size === 0 && releaseTimer === null) {
          releaseTimer = setTimeout(stop, RELEASE_GRACE_MS);
        }
      };
    },
    resume() {
      if (retryTimer === null) {
        return;
      }
      clearTimeout(retryTimer);
      void connect();
    },
    getStatus: () => status,
    subscribeStatus(onChange) {
      statusListeners.add(onChange);
      return () => {
        statusListeners.delete(onChange);
      };
    },
  };
}
