import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";

vi.mock("common/maplibregl", () => ({ default: { Map: class {} } }));

describe("map runtime helpers", () => {
  beforeEach(() => vi.resetModules());
  afterEach(() => vi.restoreAllMocks());

  it("rejects WebGL1-only clients", async () => {
    const getContext = vi.spyOn(HTMLCanvasElement.prototype, "getContext").mockImplementation((type) => type === "webgl" ? {} : null);
    const { supportsWebGL2 } = await import("common/map");
    expect(supportsWebGL2()).toBe(false);
    expect(getContext).toHaveBeenCalledExactlyOnceWith("webgl2");
  });

  it("releases its WebGL2 probe context", async () => {
    const loseContext = vi.fn();
    vi.spyOn(HTMLCanvasElement.prototype, "getContext").mockReturnValue({
      getParameter: vi.fn(),
      getExtension: vi.fn(() => ({ loseContext })),
    });
    const { supportsWebGL2 } = await import("common/map");
    expect(supportsWebGL2()).toBe(true);
    expect(loseContext).toHaveBeenCalledOnce();
  });

  it("supports contexts without an optional release extension", async () => {
    vi.spyOn(HTMLCanvasElement.prototype, "getContext").mockReturnValue({ getParameter: vi.fn(), getExtension: () => null });
    const { supportsWebGL2 } = await import("common/map");
    expect(supportsWebGL2()).toBe(true);
  });

  it("handles blocked context creation", async () => {
    vi.spyOn(HTMLCanvasElement.prototype, "getContext").mockImplementation(() => {
      throw new Error("disabled");
    });
    const { supportsWebGL2 } = await import("common/map");
    expect(supportsWebGL2()).toBe(false);
  });

  it("shares the renderer with concurrent and subsequent mounts", async () => {
    const { load } = await import("common/map");
    const first = load();
    const second = load();
    expect(first).toBe(second);
    const [a, b] = await Promise.all([first, second]);
    expect(a.Map).toBeTypeOf("function");
    expect(a).toBe(b);
    expect(await load()).toBe(a);
  });
});
