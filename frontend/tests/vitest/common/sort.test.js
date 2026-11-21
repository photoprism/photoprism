import { describe, it, expect } from "vitest";
import { albumOptions, photoOptions, normalizeSort, selectSort, sortAscending, sortLabel, photoSortKeys, resolveSort, sortParams } from "common/sort";

// storage creates a browser-like store with string values.
function storage(initial = {}) {
  const values = new Map(Object.entries(initial));
  return { getItem: (key) => values.get(key) ?? null, setItem: (key, value) => values.set(key, String(value)) };
}

const photos = photoOptions("browse");
const albums = albumOptions("album");

describe("sort options", () => {
  it.each(["favorites", "videos", "live", "animated", "audio", "review", "hidden", "private", "unsorted", "media", "places_browse"])("limits special options on %s", (route) => {
    expect(photoOptions(route).map((item) => item.value)).toEqual(["newest", "title", "name", "size", "duration", "added", "edited"]);
  });
  it("offers search and archive options only on the matching pages", () => {
    expect(photos.map((item) => item.value)).toEqual(["newest", "title", "name", "size", "duration", "added", "edited", "similar", "relevance"]);
    expect(photoOptions("all")).toEqual(photos);
    expect(photoOptions("archive").map((item) => item.value)).toEqual(["newest", "title", "name", "size", "duration", "added", "edited", "archived"]);
  });
  it.each([
    ["album", ["name", "added", "place", "edited", "favorites"]],
    ["folder", ["name", "added", "place", "favorites"]],
    ["state", ["name", "added", "place", "favorites"]],
    ["moment", ["name", "added", "newest", "place", "favorites"]],
    ["month", ["name", "added", "newest", "place", "favorites"]],
  ])("offers the exact options for %s", (type, values) => {
    expect(albumOptions(type).map((item) => item.value)).toEqual(values);
    expect(albumOptions(type).find((item) => item.value === "place").direction).toBeUndefined();
  });
  it("labels moment years and calendar dates separately", () => {
    expect(albumOptions("moment")[2].text).toBe("Year");
    expect(albumOptions("month")[2].text).toBe("Date");
  });
});

describe("sort transitions and wire values", () => {
  it.each([
    ["newest", "oldest", false], ["title", "title", true], ["name", "name", true],
    ["size", "size", true], ["duration", "duration", true], ["added", "added", false],
    ["edited", "edited", false], ["similar", "similar", false], ["relevance", "relevance", false],
  ])("toggles %s and returns to its default direction", (order, reversedOrder, reverse) => {
    const item = photos.find((entry) => entry.value === order);
    const first = selectSort({ order: "archived", reverse: true }, item);
    expect(first).toEqual({ order, reverse: false });
    const toggled = selectSort(first, item);
    expect(toggled).toEqual({ order: reversedOrder, reverse });
    expect(selectSort(toggled, item)).toEqual(first);
  });
  it.each(["title", "name", "size", "duration", "newest"])("selects both arrows of %s directly", (order) => {
    const item = photos.find((entry) => entry.value === order);
    for (const up of [true, false]) {
      const next = selectSort({ order: "added", reverse: false }, item, up);
      expect(sortAscending(next, item)).toBe(up);
      expect(selectSort(next, item, up)).toEqual(next);
    }
  });
  it("reverses album creation dates without changing the order value", () => {
    expect(selectSort({ order: "added", reverse: false }, albums[1])).toEqual({ order: "added", reverse: true });
  });
  it.each(["newest", "oldest"])("never serializes reverse with %s", (order) => {
    expect(sortParams({ order, reverse: true, q: "cats" })).toEqual({ order, q: "cats" });
  });
  it("omits false reverse without mutating filter state", () => {
    const state = { order: "name", reverse: false };
    expect(sortParams(state)).toEqual({ order: "name" });
    expect(state.reverse).toBe(false);
    expect(sortParams({ order: "name", reverse: true })).toEqual({ order: "name", reverse: true });
  });
});

describe("legacy sort normalization", () => {
  it.each(["album", "folder", "state"])("maps oldest/newest to creation dates for %s", (type) => {
    expect(normalizeSort("oldest", false, albumOptions(type), "name")).toEqual({ order: "added", reverse: true });
    expect(normalizeSort("newest", false, albumOptions(type), "name")).toEqual({ order: "added", reverse: false });
  });
  it.each(["moment", "month"])("keeps the dated sort on %s", (type) => {
    expect(normalizeSort("oldest", false, albumOptions(type), "newest")).toEqual({ order: "oldest", reverse: false });
  });
  it("normalizes reversed date URLs and refuses unsupported menu values", () => {
    expect(normalizeSort("newest", "true", photos, "added")).toEqual({ order: "oldest", reverse: false });
    expect(normalizeSort("oldest", "true", photos, "added")).toEqual({ order: "newest", reverse: false });
    expect(normalizeSort("random", true, photos, "newest")).toEqual({ order: "newest", reverse: false });
    expect(normalizeSort("edited", true, albumOptions("folder"), "name")).toEqual({ order: "name", reverse: false });
    expect(normalizeSort("place", true, albums, "favorites")).toEqual({ order: "place", reverse: false });
  });
});

describe("per-page preference restoration", () => {
  it("lets a URL order reset a stored reverse direction", () => {
    const store = storage({ "photos.order": "size", "photos.reverse": "true" });
    const keys = photoSortKeys("browse");
    expect(resolveSort({ order: "title" }, store, keys, photos, "newest")).toEqual({ order: "title", reverse: false });
    expect(store.getItem(keys.reverse)).toBe("false");
  });
  it("restores directions, supports reverse-only URLs and explicit false", () => {
    const store = storage({ "photos.order": "size", "photos.reverse": "true" });
    const keys = photoSortKeys("browse");
    expect(resolveSort({}, store, keys, photos, "newest")).toEqual({ order: "size", reverse: true });
    expect(resolveSort({ reverse: "false" }, store, keys, photos, "newest")).toEqual({ order: "size", reverse: false });
    expect(resolveSort({ reverse: "true" }, store, keys, photos, "newest")).toEqual({ order: "size", reverse: true });
  });
  it("seeds new route preferences once from the old shared key", () => {
    const store = storage({ "photos.order": "oldest" });
    const videos = photoSortKeys("videos");
    const live = photoSortKeys("live");
    expect(resolveSort({}, store, videos, photoOptions("videos"), "newest").order).toBe("oldest");
    resolveSort({ order: "size", reverse: "true" }, store, videos, photoOptions("videos"), "newest");
    expect(resolveSort({}, store, live, photoOptions("live"), "newest")).toEqual({ order: "oldest", reverse: false });
    expect(resolveSort({}, store, videos, photoOptions("videos"), "newest")).toEqual({ order: "size", reverse: true });
    expect(store.getItem("photos.order")).toBe("oldest");
  });
  it.each(["archive", "favorites", "hidden", "review"])("preserves the %s storage key", (route) => {
    expect(photoSortKeys(route)).toEqual({ order: `${route}.order`, reverse: `${route}.reverse` });
  });
  it("uses the supplied page default when no preference exists", () => {
    expect(resolveSort({}, storage(), photoSortKeys("archive"), photoOptions("archive"), "archived")).toEqual({ order: "archived", reverse: false });
  });
});

describe("translated sort descriptions", () => {
  it.each([
    ["newest", false, "Sorted by Date Taken, descending"],
    ["oldest", false, "Sorted by Date Taken, ascending"],
    ["title", false, "Sorted by Title, ascending"],
    ["title", true, "Sorted by Title, descending"],
    ["size", true, "Sorted by File Size, ascending"],
    ["size", false, "Sorted by File Size, descending"],
    ["duration", true, "Sorted by Video Duration, ascending"],
    ["duration", false, "Sorted by Video Duration, descending"],
    ["added", false, "Sorted by Recently Added"],
  ])("describes %s reverse=%s", (order, reverse, label) => {
    const item = photos.find((entry) => entry.value === (order === "oldest" ? "newest" : order));
    expect(sortLabel({ order, reverse }, item)).toBe(label);
  });
  it("describes Location without a direction", () => {
    const location = albumOptions("album").find((item) => item.value === "place");
    expect(sortLabel({ order: "place", reverse: false }, location)).toBe("Sorted by Location");
  });
  it("has an accessible fallback for a missing option", () => {
    expect(sortLabel({}, undefined)).toBe("Sort By");
  });
});
