import { describe, it, expect } from "vitest";
import { AlbumSortOptions, albumOrder, nextAlbumOrder, albumCoverFrame, albumCoverTone, albumCollapseProgress } from "common/album-view";

describe("album view ordering", () => {
  it("uses saved ordering and normalizes URL directions", () => {
    expect(albumOrder({}, "oldest")).toEqual({ order: "oldest", reverse: false });
    expect(albumOrder({ order: "added", reverse: "true" })).toEqual({ order: "added", reverse: false });
    expect(albumOrder({ order: "newest", reverse: "true" })).toEqual({ order: "newest", reverse: false });
    expect(albumOrder({ order: "size", reverse: "true" })).toEqual({ order: "size", reverse: true });
    expect(albumOrder({ order: "unknown" })).toEqual({ order: "newest", reverse: false });
  });
  it("uses separate date orders and reverses other two-direction options", () => {
    expect(nextAlbumOrder("newest", { order: "newest" })).toEqual({ order: "oldest", reverse: false });
    expect(nextAlbumOrder("newest", { order: "oldest" })).toEqual({ order: "newest", reverse: false });
    for (const order of ["title", "name", "size", "duration"]) {
      expect(nextAlbumOrder(order, { order, reverse: false })).toEqual({ order, reverse: true });
      expect(nextAlbumOrder(order, { order, reverse: true })).toEqual({ order, reverse: false });
    }
    for (const order of ["added", "edited", "relevance"]) {
      expect(nextAlbumOrder(order, { order, reverse: true })).toEqual({ order, reverse: false });
      expect(AlbumSortOptions.find((item) => item.value === order).reversible).toBeUndefined();
    }
  });
});

describe("album cover framing", () => {
  it("keeps a face near the image edge visible on desktop and phone", () => {
    const faces = { X: 0.7, Y: 0.02, W: 0.15, H: 0.2 };
    for (const [width, height] of [[1440, 350], [768, 320], [390, 320]]) {
      const frame = albumCoverFrame(width, height, 1600, 1200, faces);
      expect(frame.left + faces.X * frame.width).toBeGreaterThanOrEqual(0);
      expect(frame.top + faces.Y * frame.height).toBeGreaterThanOrEqual(0);
      expect(frame.left + (faces.X + faces.W) * frame.width).toBeLessThanOrEqual(width);
      expect(frame.top + (faces.Y + faces.H) * frame.height).toBeLessThanOrEqual(height);
    }
  });
  it("keeps faces outside the toolbar and caption", () => {
    const faces = { X: 0.7, Y: 0.02, W: 0.15, H: 0.2 };
    for (const [width, height, top, bottom] of [[1440, 350, 64, 100], [390, 320, 48, 140]]) {
      const frame = albumCoverFrame(width, height, 1600, 1200, faces, { top, bottom });
      expect(frame.top + faces.Y * frame.height).toBeGreaterThanOrEqual(top);
      expect(frame.top + (faces.Y + faces.H) * frame.height).toBeLessThanOrEqual(height - bottom);
    }
  });
  it("reduces the image for a wide group and centers covers without faces", () => {
    const frame = albumCoverFrame(390, 320, 2000, 1000, { X: 0, Y: 0.3, W: 1, H: 0.2 });
    expect(frame.width).toBe(390);
    expect(frame.height).toBeLessThan(320);
    expect(frame.top).toBeGreaterThan(0);
    expect(albumCoverFrame(100, 100, 200, 100)).toEqual({ width: 200, height: 100, left: -50, top: 0 });
    expect(albumCoverFrame(0, 100, 200, 100).width).toBe(0);
  });
  it("uses no shade on a uniformly light or dark cover", () => {
    expect(albumCoverTone([255, 255, 255, 255])).toMatchObject({
      "--album-cover-ink": "#000",
      "--album-cover-tint": "255, 255, 255",
      "--album-cover-shade": "0",
      "--album-toolbar-shade": "0",
    });
    expect(albumCoverTone([0, 0, 0, 255])).toMatchObject({
      "--album-cover-ink": "#fff",
      "--album-cover-tint": "0, 0, 0",
      "--album-cover-shade": "0",
      "--album-toolbar-shade": "0",
    });
  });
  it("shades only the region that needs contrast for the cover's text color", () => {
    const white = [255, 255, 255, 255];
    const black = [0, 0, 0, 255];
    const mostlyWhite = [...white, ...white, ...black];
    expect(albumCoverTone(mostlyWhite, black, white)).toMatchObject({
      "--album-cover-ink": "#000",
      "--album-cover-shade": "0.46",
      "--album-toolbar-shade": "0",
    });
    expect(albumCoverTone(black, black, white)).toMatchObject({
      "--album-cover-ink": "#fff",
      "--album-cover-shade": "0",
      "--album-toolbar-shade": "0.54",
    });
  });
  it("uses a contrast-safe shade where the cover is unavailable", () => {
    expect(albumCoverTone([])["--album-cover-shade"]).toBe("0.54");
    expect(albumCoverTone([255, 255, 255, 255], [0, 0, 0, 0])["--album-cover-shade"]).toBe("0.46");
  });
});

describe("album toolbar transition", () => {
  it("is transparent at the top and fades continuously before the cover disappears", () => {
    expect(albumCollapseProgress(350)).toBe(0);
    expect(albumCollapseProgress(168)).toBe(0.5);
    expect(albumCollapseProgress(48)).toBe(1);
    expect(albumCollapseProgress(-100)).toBe(1);
  });
});
