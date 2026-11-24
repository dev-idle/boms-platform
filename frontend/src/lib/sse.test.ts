import { describe, expect, it } from "vitest";

import { readServerSentEvents, type ServerSentEvent } from "./sse";

/** A stream that delivers the chunks as they are cut, the way a network does. */
function streamOf(...chunks: string[]): ReadableStream<Uint8Array> {
  const encoder = new TextEncoder();
  return new ReadableStream({
    start(controller) {
      for (const chunk of chunks) {
        controller.enqueue(encoder.encode(chunk));
      }
      controller.close();
    },
  });
}

async function collect(stream: ReadableStream<Uint8Array>): Promise<ServerSentEvent[]> {
  const events: ServerSentEvent[] = [];
  for await (const event of readServerSentEvents(stream)) {
    events.push(event);
  }
  return events;
}

describe("readServerSentEvents", () => {
  it("reads named events, whatever chunks they arrive in", async () => {
    const events = await collect(
      streamOf('event: session\ndata: {"session_id":"1"}\n\nevent: del', 'ta\ndata: {"text":"Hi"}\n', "\n"),
    );

    expect(events).toEqual([
      { event: "session", data: '{"session_id":"1"}' },
      { event: "delta", data: '{"text":"Hi"}' },
    ]);
  });

  it("joins data lines, names an unnamed event and skips comments and empty events", async () => {
    const events = await collect(streamOf(": keep-alive\n\ndata: a\ndata: b\n\nevent: ping\n\n"));

    expect(events).toEqual([{ event: "message", data: "a\nb" }]);
  });

  it("drops an event the stream cut off", async () => {
    const events = await collect(streamOf("event: answer\ndata: {"));

    expect(events).toEqual([]);
  });
});
