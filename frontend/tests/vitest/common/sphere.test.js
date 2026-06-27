import "../fixtures";
import { expect, describe, it, vi, beforeEach } from "vitest";

const { viewerCtor, destroySpy } = vi.hoisted(() => ({
  viewerCtor: vi.fn(),
  destroySpy: vi.fn(),
}));

vi.mock("@photo-sphere-viewer/core", () => ({
  Viewer: function Viewer(cfg) {
    viewerCtor(cfg);
    this.destroy = destroySpy;
  },
}));

vi.mock("@photo-sphere-viewer/video-plugin", () => ({
  VideoPlugin: function VideoPlugin() {},
}));

vi.mock("@photo-sphere-viewer/equirectangular-video-adapter", () => ({
  EquirectangularVideoAdapter: function EquirectangularVideoAdapter() {},
}));

vi.mock("@photo-sphere-viewer/core/index.css", () => ({}));
vi.mock("@photo-sphere-viewer/video-plugin/index.css", () => ({}));

import { createSphereViewer, destroySphereViewer, is360Equirectangular } from "common/sphere";

describe("common/sphere is360Equirectangular", () => {
  it("is true for an explicit equirectangular projection without known dimensions", () => {
    expect(is360Equirectangular({ Type: "image", Projection: "equirectangular" })).toBe(true);
    expect(is360Equirectangular({ Type: "video", Projection: "equirectangular" })).toBe(true);
  });
  it("is true for an explicit equirectangular projection with ~2:1 dimensions", () => {
    expect(is360Equirectangular({ Type: "image", Projection: "equirectangular", Width: 6656, Height: 3328 })).toBe(true);
    expect(is360Equirectangular({ Type: "image", Projection: "equirectangular", Width: 15520, Height: 7760 })).toBe(true);
  });
  it("is false for an equirectangular-tagged frame that is clearly not 2:1 (partial/cylindrical panorama)", () => {
    expect(is360Equirectangular({ Type: "image", Projection: "equirectangular", Width: 8192, Height: 1024 })).toBe(false);
    expect(is360Equirectangular({ Type: "video", Projection: "equirectangular", Width: 8192, Height: 1024 })).toBe(false);
  });
  it("is false for any other non-empty projection", () => {
    expect(is360Equirectangular({ Type: "video", Projection: "cubemap", Panorama: true, Width: 3840, Height: 1920 })).toBe(false);
    expect(is360Equirectangular({ Type: "image", Projection: "cubestrip" })).toBe(false);
  });
  it("falls back to the 2:1 frame for a panorama video without a projection tag", () => {
    expect(is360Equirectangular({ Type: "video", Panorama: true, Width: 3840, Height: 1920 })).toBe(true);
    expect(is360Equirectangular({ Type: "video", Panorama: true, Width: 4096, Height: 2048 })).toBe(true);
  });
  it("is false for a wide non-2:1 video (ultrawide) even when panorama-flagged", () => {
    expect(is360Equirectangular({ Type: "video", Panorama: true, Width: 3840, Height: 1632 })).toBe(false);
  });
  it("is false for a 2:1 video that is not panorama-flagged", () => {
    expect(is360Equirectangular({ Type: "video", Panorama: false, Width: 3840, Height: 1920 })).toBe(false);
  });
  it("does not apply the 2:1 fallback to photos (projection only)", () => {
    expect(is360Equirectangular({ Type: "image", Panorama: true, Width: 3840, Height: 1920 })).toBe(false);
  });
  it("is false for empty, null, or dimensionless input", () => {
    expect(is360Equirectangular(null)).toBe(false);
    expect(is360Equirectangular({})).toBe(false);
    expect(is360Equirectangular({ Type: "video", Panorama: true })).toBe(false);
  });
});

describe("common/sphere", () => {
  beforeEach(() => {
    viewerCtor.mockClear();
    destroySpy.mockClear();
  });

  it("photo path constructs Viewer with panorama option", async () => {
    const container = document.createElement("div");
    const viewer = await createSphereViewer(container, "/photo.jpg");
    expect(viewerCtor).toHaveBeenCalledOnce();
    expect(viewerCtor.mock.calls[0][0]).toMatchObject({
      container,
      panorama: "/photo.jpg",
      keyboard: "always",
      defaultYaw: 0,
      defaultPitch: 0,
    });
    expect(viewer).toBeDefined();
  });

  it("video path wires VideoPlugin and equirectangular adapter", async () => {
    const container = document.createElement("div");
    await createSphereViewer(container, "/video.mp4", { isVideo: true });
    expect(viewerCtor).toHaveBeenCalledOnce();
    const cfg = viewerCtor.mock.calls[0][0];
    expect(cfg.panorama).toEqual({ source: "/video.mp4" });
    expect(cfg.plugins).toHaveLength(1);
    expect(cfg.adapter).toHaveLength(2);
  });

  it("destroySphereViewer calls viewer.destroy exactly once", async () => {
    const container = document.createElement("div");
    const viewer = await createSphereViewer(container, "/photo.jpg");
    destroySphereViewer(viewer);
    expect(destroySpy).toHaveBeenCalledOnce();
  });

  it("destroySphereViewer is safe on null", () => {
    expect(() => destroySphereViewer(null)).not.toThrow();
    expect(() => destroySphereViewer(undefined)).not.toThrow();
    expect(destroySpy).not.toHaveBeenCalled();
  });

  it("destroySphereViewer is safe on viewer without destroy method", () => {
    expect(() => destroySphereViewer({})).not.toThrow();
    expect(destroySpy).not.toHaveBeenCalled();
  });
});
