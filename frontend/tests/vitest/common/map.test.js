import { describe, it, expect } from "vitest";
import { groupGeoFeatures } from "common/map";

const feature = (uid, lng, lat, id) => ({
  id: id ?? uid,
  geometry: { coordinates: [lng, lat] },
  properties: { UID: uid, Hash: `${uid}hash` },
});

// Identity projection treats raw coordinates as screen pixels so the proximity grouping can
// be exercised with a small, explicit tolerance and no live map.
const identity = (coords) => [coords[0], coords[1]];

describe("common/map.groupGeoFeatures", () => {
  it("should return an empty array for invalid input", () => {
    expect(groupGeoFeatures(undefined)).toEqual([]);
    expect(groupGeoFeatures(null)).toEqual([]);
    expect(groupGeoFeatures([])).toEqual([]);
  });

  it("should keep distinct locations in separate groups", () => {
    const groups = groupGeoFeatures([feature("p1", 10, 20), feature("p2", 11, 21)], identity, 0.5);
    expect(groups.length).toBe(2);
    expect(groups[0].features.length).toBe(1);
    expect(groups[1].features.length).toBe(1);
  });

  it("should group pictures sharing the exact same coordinates", () => {
    const groups = groupGeoFeatures([feature("p1", 16.5, 47.5), feature("p2", 16.5, 47.5), feature("p3", 16.5, 47.5)], identity, 0.5);
    expect(groups.length).toBe(1);
    expect(groups[0].features.length).toBe(3);
  });

  it("should group near-coincident pictures within the tolerance", () => {
    // Pictures a fraction apart (GPS jitter or tile quantization) must still stack.
    const groups = groupGeoFeatures([feature("p1", 16.5, 47.5), feature("p2", 16.5001, 47.5001), feature("p3", 16.4999, 47.4999)], identity, 0.5);
    expect(groups.length).toBe(1);
    expect(groups[0].features.length).toBe(3);
  });

  it("should not group pictures beyond the tolerance", () => {
    const groups = groupGeoFeatures([feature("p1", 16.5, 47.5), feature("p2", 17.5, 47.5)], identity, 0.5);
    expect(groups.length).toBe(2);
  });

  it("should fall back to raw coordinate (degree) grouping when no projection is given", () => {
    // Without a projection the tolerance is applied in raw [lng, lat] degrees rather than
    // screen pixels, so grouping is not zoom-aware. The default 40-degree tolerance merges
    // coordinates that are far apart on screen — which is why the live map always passes a
    // projection. Two points ~0.7 degrees apart merge; two points 100 degrees apart do not.
    const near = groupGeoFeatures([feature("p1", 10, 20), feature("p2", 10.5, 20.5)]);
    expect(near.length).toBe(1);
    expect(near[0].features.length).toBe(2);
    const far = groupGeoFeatures([feature("p1", 0, 0), feature("p2", 100, 0)]);
    expect(far.length).toBe(2);
  });

  it("should group by pixel distance when given a maplibregl Point projection", () => {
    const project = (coords) => ({ x: coords[0] * 1000, y: coords[1] * 1000 });
    // 0.001 deg => 1px apart with this projection => within the default tolerance.
    const near = groupGeoFeatures([feature("p1", 16.5, 47.5), feature("p2", 16.501, 47.5)], project);
    expect(near.length).toBe(1);
    // 0.1 deg => 100px apart => beyond the default tolerance.
    const far = groupGeoFeatures([feature("p1", 16.5, 47.5), feature("p2", 16.6, 47.5)], project);
    expect(far.length).toBe(2);
  });

  it("should deduplicate features that reference the same picture", () => {
    const groups = groupGeoFeatures([feature("p1", 16.5, 47.5), feature("p1", 16.5, 47.5), feature("p2", 16.5, 47.5)], identity, 0.5);
    expect(groups.length).toBe(1);
    expect(groups[0].features.length).toBe(2);
    expect(Object.keys(groups[0].uids).sort()).toEqual(["p1", "p2"]);
  });

  it("should skip features without valid coordinates", () => {
    const groups = groupGeoFeatures([feature("p1", 16.5, 47.5), { id: "x", properties: { UID: "x" } }, { geometry: { coordinates: [1] }, properties: { UID: "y" } }], identity, 0.5);
    expect(groups.length).toBe(1);
    expect(groups[0].features.length).toBe(1);
  });

  it("should preserve features missing a UID without deduplicating them", () => {
    const a = { id: "a", geometry: { coordinates: [0, 0] }, properties: {} };
    const b = { id: "b", geometry: { coordinates: [0, 0] }, properties: {} };
    const groups = groupGeoFeatures([a, b], identity, 0.5);
    expect(groups.length).toBe(1);
    expect(groups[0].features.length).toBe(2);
  });

  it("should expose the lat/lng bounding box of each group", () => {
    const groups = groupGeoFeatures([feature("p2", 16.6, 47.6), feature("p1", 16.4, 47.4)], identity, 0.5);
    expect(groups.length).toBe(1);
    expect(groups[0].bounds).toEqual({ latN: 47.6, lngE: 16.6, latS: 47.4, lngW: 16.4 });
  });

  it("should derive a stable key from the smallest UID", () => {
    const groups = groupGeoFeatures([feature("p3", 16.5, 47.5), feature("p1", 16.5, 47.5), feature("p2", 16.5, 47.5)], identity, 0.5);
    expect(groups.length).toBe(1);
    expect(groups[0].key).toBe("p1");
  });
});
