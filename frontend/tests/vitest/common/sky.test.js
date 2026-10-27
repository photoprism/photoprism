import { afterEach, describe, it, expect, vi } from "vitest";
import * as sky from "common/sky";

// fakeGl returns a WebGL2 stub that records draws and context loss.
function fakeGl(maxTextureSize = 8192) {
  const loseContext = vi.fn();
  return new Proxy(
    { loseContext, drawArrays: vi.fn(), createProgram: () => ({}), getParameter: () => maxTextureSize, getExtension: () => ({ loseContext }), isContextLost: () => false, getProgramParameter: () => true },
    { get: (target, key) => (key in target ? target[key] : typeof key === "string" && key === key.toUpperCase() ? 0 : vi.fn()) }
  );
}

// FakeImage stands in for a decoded 4096x2048 star map.
class FakeImage {
  naturalWidth = 4096;
  decode() {
    return Promise.resolve();
  }
}

function dot(a, b) {
  return a[0] * b[0] + a[1] * b[1] + a[2] * b[2];
}

function cross(a, b) {
  return [a[1] * b[2] - a[2] * b[1], a[2] * b[0] - a[0] * b[2], a[0] * b[1] - a[1] * b[0]];
}

describe("common/sky", () => {
  describe("siderealTime", () => {
    it("returns the sidereal time at the J2000 epoch", () => {
      expect(sky.siderealTime(new Date("2000-01-01T12:00:00Z"))).toBeCloseTo(280.4606, 3);
    });
    it("advances by about one degree more than a full turn per day", () => {
      const a = sky.siderealTime(new Date("2026-10-26T00:00:00Z"));
      const b = sky.siderealTime(new Date("2026-10-27T00:00:00Z"));
      expect((b - a + 360) % 360).toBeCloseTo(0.9856, 3);
    });
  });
  describe("atmosphere", () => {
    it("draws only the atmosphere and fades it out before the max zoom", () => {
      const result = sky.atmosphere();
      expect(result["sky-color"]).toBe("transparent");
      expect(result["horizon-color"]).toBe("transparent");
      expect(result["atmosphere-blend"].slice(-2)).toEqual([sky.AtmosphereMaxZoom, 0]);
      expect(result["atmosphere-blend"][4]).toBe(sky.AtmosphereBlend);
    });
  });
  describe("styleSky", () => {
    it("returns undefined for a style without sky", () => {
      expect(sky.styleSky(undefined)).toBeUndefined();
      expect(sky.styleSky(null)).toBeUndefined();
      expect(sky.styleSky({})).toBeUndefined();
    });
    it("fills the properties set by the atmosphere with their defaults", () => {
      const result = sky.styleSky({ "sky-color": "#000000" });
      expect(result["sky-color"]).toBe("#000000");
      expect(result["atmosphere-blend"]).toBe(0.8);
      Object.keys(sky.atmosphere()).forEach((key) => expect(result).toHaveProperty(key));
    });
  });
  describe("skyScale", () => {
    it("uses the narrowest field of view on small canvases", () => {
      expect(sky.skyScale(900, 4096)).toBeCloseTo(Math.tan((sky.SkyFovMin * Math.PI) / 360), 9);
    });
    it("widens the field of view to limit the magnification", () => {
      const scale = sky.skyScale(1400, 4096);
      expect(scale).toBeGreaterThan(Math.tan((sky.SkyFovMin * Math.PI) / 360));
      // Pixels per radian at the center divided by texels per radian.
      expect(700 / scale / (4096 / (2 * Math.PI))).toBeCloseTo(sky.MaxMagnification, 9);
    });
    it("limits the field of view on very large canvases", () => {
      expect(sky.skyScale(4320, 4096)).toBeCloseTo(Math.tan((sky.SkyFovMax * Math.PI) / 360), 9);
    });
  });
  describe("cameraBasis", () => {
    it("returns an orthonormal right-handed basis", () => {
      const { forward, right, up } = sky.cameraBasis(12.5, 48.1, 35, 40, 123);
      [forward, right, up].forEach((v) => expect(dot(v, v)).toBeCloseTo(1, 9));
      expect(dot(forward, right)).toBeCloseTo(0, 9);
      expect(dot(forward, up)).toBeCloseTo(0, 9);
      expect(dot(right, up)).toBeCloseTo(0, 9);
      cross(right, up).forEach((c, i) => expect(c).toBeCloseTo(-forward[i], 9));
    });
    it("looks down at the center without pitch", () => {
      const { forward, up } = sky.cameraBasis(0, 90, 0, 0, 0);
      expect(forward[2]).toBeCloseTo(-1, 9);
      expect(up[2]).toBeCloseTo(0, 9);
    });
    it("turns the top of the screen toward the bearing", () => {
      const { up, right } = sky.cameraBasis(0, 0, 90, 0, 0);
      [0, 1, 0].forEach((c, i) => expect(up[i]).toBeCloseTo(c, 9));
      [0, 0, -1].forEach((c, i) => expect(right[i]).toBeCloseTo(c, 9));
    });
    it("looks at the horizon toward the bearing with full pitch", () => {
      const { forward } = sky.cameraBasis(0, 0, 0, 90, 0);
      expect(forward[2]).toBeCloseTo(1, 9);
    });
    it("rotates with the sidereal time", () => {
      const a = sky.cameraBasis(30, 0, 0, 0, 0).forward;
      const b = sky.cameraBasis(0, 0, 0, 0, 30).forward;
      a.forEach((c, i) => expect(c).toBeCloseTo(b[i], 9));
    });
  });
  describe("Sky", () => {
    afterEach(() => {
      vi.unstubAllGlobals();
      vi.restoreAllMocks();
    });
    it("follows the attached map and stops when destroyed", () => {
      const map = { on: vi.fn(), off: vi.fn() };
      const renderer = new sky.Sky(document.createElement("div"));
      renderer.attach(map);
      renderer.attach(map);
      expect(map.on).toHaveBeenCalledTimes(2);
      expect(map.on).toHaveBeenCalledWith("render", renderer.onRender);
      renderer.destroy();
      expect(map.off).toHaveBeenCalledWith("render", renderer.onRender);
      expect(map.off).toHaveBeenCalledWith("resize", renderer.onResize);
      expect(renderer.map).toBeNull();
      expect(renderer.destroyed).toBe(true);
    });
    it("ignores attach after it was destroyed", () => {
      const map = { on: vi.fn(), off: vi.fn() };
      const renderer = new sky.Sky(document.createElement("div"));
      renderer.destroy();
      renderer.attach(map);
      expect(map.on).not.toHaveBeenCalled();
    });
    it("does not create a canvas when destroyed while the image loads", async () => {
      vi.stubGlobal("Image", FakeImage);
      const getContext = vi.spyOn(HTMLCanvasElement.prototype, "getContext");
      const el = document.createElement("div");
      const renderer = new sky.Sky(el);
      const loading = renderer.load("/static/img/sky/stars.avif");
      renderer.destroy();
      await expect(loading).rejects.toThrow("destroyed");
      expect(getContext).not.toHaveBeenCalled();
      expect(el.children).toHaveLength(0);
    });
    it("releases the context when the texture is too large", async () => {
      vi.stubGlobal("Image", FakeImage);
      const gl = fakeGl(2048);
      vi.spyOn(HTMLCanvasElement.prototype, "getContext").mockReturnValue(gl);
      const el = document.createElement("div");
      const renderer = new sky.Sky(el);
      await expect(renderer.load("/static/img/sky/stars.avif")).rejects.toThrow("too large");
      expect(gl.loseContext).toHaveBeenCalledOnce();
      expect(el.children).toHaveLength(0);
    });
    it("rejects when the image cannot be decoded", async () => {
      vi.stubGlobal(
        "Image",
        class extends FakeImage {
          decode() {
            return Promise.reject(new Error("EncodingError"));
          }
        }
      );
      await expect(new sky.Sky(document.createElement("div")).load("/missing.avif")).rejects.toThrow("failed to load image");
    });
    it("draws in globe projection only when the camera changed", async () => {
      vi.stubGlobal("Image", FakeImage);
      const gl = fakeGl();
      vi.spyOn(HTMLCanvasElement.prototype, "getContext").mockReturnValue(gl);
      const renderer = new sky.Sky(document.createElement("div"));
      await renderer.load("/static/img/sky/stars.avif");
      let projection = "mercator";
      let lng = 10;
      const map = { on: vi.fn(), off: vi.fn(), getProjection: () => ({ type: projection }), getCenter: () => ({ lng, lat: 20 }), getBearing: () => 0, getPitch: () => 0 };
      renderer.attach(map);
      expect(gl.drawArrays).not.toHaveBeenCalled();
      projection = "globe";
      renderer.draw();
      renderer.draw();
      expect(gl.drawArrays).toHaveBeenCalledOnce();
      lng = 11;
      renderer.draw();
      expect(gl.drawArrays).toHaveBeenCalledTimes(2);
      renderer.destroy();
      expect(gl.loseContext).toHaveBeenCalledOnce();
    });
    it("hands over to the owner when the context is lost", async () => {
      vi.stubGlobal("Image", FakeImage);
      vi.spyOn(HTMLCanvasElement.prototype, "getContext").mockReturnValue(fakeGl());
      const onLost = vi.fn();
      const el = document.createElement("div");
      const renderer = new sky.Sky(el, { onLost });
      await renderer.load("/static/img/sky/stars.avif");
      el.querySelector(".sky-canvas").dispatchEvent(new Event("webglcontextlost"));
      expect(onLost).toHaveBeenCalledOnce();
      expect(renderer.destroyed).toBe(true);
      expect(el.children).toHaveLength(0);
    });
    it("stays quiet when its own destroy releases the context", () => {
      const onLost = vi.fn();
      const renderer = new sky.Sky(document.createElement("div"), { onLost });
      renderer.destroy();
      renderer.lost();
      expect(onLost).not.toHaveBeenCalled();
    });
    it("does not draw before WebGL is initialized", () => {
      const map = { on: vi.fn(), off: vi.fn(), getCenter: vi.fn() };
      const renderer = new sky.Sky(document.createElement("div"));
      renderer.attach(map);
      renderer.draw();
      expect(map.getCenter).not.toHaveBeenCalled();
    });
  });
  describe("render", () => {
    it("adds the specified number of stars", () => {
      const el = document.createElement("div");
      sky.render(el, 12);
      expect(el.querySelectorAll(".star")).toHaveLength(12);
    });
    it("ignores a missing element", () => {
      expect(() => sky.render(null, 12)).not.toThrow();
    });
  });
});
