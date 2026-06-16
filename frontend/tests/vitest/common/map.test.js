import { describe, it, expect } from "vitest";
import { groupGeoFeatures } from "common/map";

const feature = (uid, lng, lat, id) => ({
  id: id ?? uid,
  geometry: { coordinates: [lng, lat] },
  properties: { UID: uid, Hash: `${uid}hash` },
});

describe("common/map.groupGeoFeatures", () => {
  it("should return an empty array for invalid input", () => {
    expect(groupGeoFeatures(undefined)).toEqual([]);
    expect(groupGeoFeatures(null)).toEqual([]);
    expect(groupGeoFeatures([])).toEqual([]);
  });

  it("should keep distinct locations in separate groups", () => {
    const groups = groupGeoFeatures([feature("p1", 10, 20), feature("p2", 11, 21)]);
    expect(groups.length).toBe(2);
    expect(groups[0].features.length).toBe(1);
    expect(groups[1].features.length).toBe(1);
  });

  it("should group pictures sharing the exact same coordinates", () => {
    const groups = groupGeoFeatures([feature("p1", 16.5, 47.5), feature("p2", 16.5, 47.5), feature("p3", 16.5, 47.5)]);
    expect(groups.length).toBe(1);
    expect(groups[0].features.length).toBe(3);
    expect(groups[0].coords).toEqual([16.5, 47.5]);
    expect(groups[0].key).toBe("16.5,47.5");
  });

  it("should deduplicate features that reference the same picture", () => {
    const groups = groupGeoFeatures([feature("p1", 16.5, 47.5), feature("p1", 16.5, 47.5), feature("p2", 16.5, 47.5)]);
    expect(groups.length).toBe(1);
    expect(groups[0].features.length).toBe(2);
    expect(Object.keys(groups[0].uids)).toEqual(["p1", "p2"]);
  });

  it("should skip features without valid coordinates", () => {
    const groups = groupGeoFeatures([feature("p1", 16.5, 47.5), { id: "x", properties: { UID: "x" } }, { geometry: { coordinates: [1] }, properties: { UID: "y" } }]);
    expect(groups.length).toBe(1);
    expect(groups[0].features.length).toBe(1);
  });

  it("should preserve features missing a UID without deduplicating them", () => {
    const a = { id: "a", geometry: { coordinates: [0, 0] }, properties: {} };
    const b = { id: "b", geometry: { coordinates: [0, 0] }, properties: {} };
    const groups = groupGeoFeatures([a, b]);
    expect(groups.length).toBe(1);
    expect(groups[0].features.length).toBe(2);
  });
});
