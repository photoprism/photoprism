import { beforeEach, describe, expect, it, vi } from "vitest";

const { setWorkerUrl } = vi.hoisted(() => ({ setWorkerUrl: vi.fn() }));
vi.mock("maplibre-gl", () => ({
  setWorkerUrl,
  Map: class {
    // setStyle records the style and preserves MapLibre's fluent contract.
    setStyle(style) {
      this.style = style; return this;
    }
    // getStyle returns the currently loaded style.
    getStyle() {
      return this.style;
    }
    // once records the one-shot listener for explicit style-load delivery.
    once(event, handler) {
      this.listener = handler; return this;
    }
  },
}));

const style = { version: 8, sources: {}, layers: [{ id: "labels", type: "symbol", layout: { "text-field": "{name}" } }] };

describe("MapLibre integration", () => {
  beforeEach(async () => {
    const { default: maps } = await import("common/maplibregl");
    maps.Map.prototype.setLanguageEnabled(true);
  });

  it("configures the module worker and preserves fluent setStyle", async () => {
    const { default: maps } = await import("common/maplibregl");
    expect(setWorkerUrl).toHaveBeenCalledWith(expect.stringContaining("maplibre-gl-worker.mjs"));
    const map = new maps.Map();
    expect(map.setStyle(style)).toBe(map);
  });

  it("decorates language labels without mutating the input style", async () => {
    const { default: maps } = await import("common/maplibregl");
    const map = new maps.Map();
    map.setStyle(style);
    map.setLanguage("de", true);
    expect(map.style.layers.map((layer) => layer.id)).toEqual(["labels", "labels-de"]);
    expect(map.style.layers[1].layout["text-field"]).toBe("{name:de}");
    expect(style.layers).toHaveLength(1);
    expect(style.layers[0].layout["text-field"]).toBe("{name}");
    map.setStyle(style);
    map.listener();
    expect(map.style.layers[1].id).toBe("labels-de");
  });

  it("supports native labels and disabling automatic decoration", async () => {
    const { default: maps } = await import("common/maplibregl");
    const map = new maps.Map();
    map.setStyle(style);
    map.setLanguage("native");
    expect(map.style.layers).toHaveLength(1);
    expect(map.style.layers[0].layout["text-field"]).toBe("{name}");
    map.setLanguageEnabled(false);
    map.listener = null;
    map.setStyle(style);
    expect(map.listener).toBeNull();
  });
});
