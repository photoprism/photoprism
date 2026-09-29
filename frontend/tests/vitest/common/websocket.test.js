// Pins the reconnect backoff of common/websocket.js; Sockette is mocked so no connection is opened.
import { describe, it, expect, vi, beforeAll } from "vitest";

const sockette = vi.hoisted(() => ({ opts: null }));

vi.mock("sockette", () => ({
  default: vi.fn(function (url, opts) {
    sockette.opts = opts;
  }),
}));

vi.mock("app/session", () => ({
  $config: { disconnected: { value: false } },
}));

// The setup file already loaded the module with the real Sockette, so it is imported again here.
let nextReconnectTimeout, reconnectTimeout, maxReconnectTimeout;

beforeAll(async () => {
  vi.resetModules();
  ({ nextReconnectTimeout, reconnectTimeout, maxReconnectTimeout } = await import("common/websocket"));
});

describe("common/websocket.js", () => {
  describe("nextReconnectTimeout", () => {
    it("doubles the delay up to the maximum while rate limited", () => {
      expect(nextReconnectTimeout(reconnectTimeout, "websocket.rate-limited")).toBe(10e3);
      expect(nextReconnectTimeout(40e3, "websocket.rate-limited")).toBe(maxReconnectTimeout);
      expect(nextReconnectTimeout(maxReconnectTimeout, "websocket.rate-limited")).toBe(maxReconnectTimeout);
      expect(nextReconnectTimeout(0, "websocket.rate-limited")).toBe(10e3);
    });

    it("restores the default after any other message", () => {
      expect(nextReconnectTimeout(maxReconnectTimeout, "config.updated")).toBe(reconnectTimeout);
      expect(nextReconnectTimeout(reconnectTimeout, undefined)).toBe(reconnectTimeout);
    });
  });

  describe("Sockette options", () => {
    it("updates the reconnect timeout from server messages", () => {
      const opts = sockette.opts;
      expect(opts.timeout).toBe(reconnectTimeout);

      opts.onmessage({ data: JSON.stringify({ event: "websocket.rate-limited", data: { code: 429 } }) });
      expect(opts.timeout).toBe(10e3);

      opts.onmessage({ data: JSON.stringify({ event: "websocket.rate-limited", data: { code: 429 } }) });
      expect(opts.timeout).toBe(20e3);

      opts.onmessage({ data: JSON.stringify({ event: "config.updated", data: {} }) });
      expect(opts.timeout).toBe(reconnectTimeout);
    });
  });
});
