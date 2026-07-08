// Targets the onUpdate switch and the uid-scoped refetch helpers in
// page/albums.vue. Calling the Options API methods directly with a
// stub `this` exercises the dispatch logic in isolation, pinning the
// contract that albums.created and albums.updated are UID-only signals
// answered by one uid-filtered query instead of a full result refresh.
import { describe, it, expect, beforeEach, afterEach, vi } from "vitest";

import PPageAlbums from "page/albums.vue";
import Album from "model/album";

// Captures the surface of `this` that the handlers touch.
function newStub() {
  return {
    listen: true,
    dirty: false,
    results: [],
    staticFilter: { type: "album" },
    refresh: vi.fn(),
    refetchResults: vi.fn(),
    insertCreated: vi.fn(),
    removeSelection: vi.fn(),
  };
}

describe("page/albums.vue onUpdate", () => {
  const onUpdate = PPageAlbums.methods.onUpdate;

  let warnSpy;
  beforeEach(() => {
    vi.spyOn(console, "log").mockImplementation(() => {});
    warnSpy = vi.spyOn(console, "warn").mockImplementation(() => {});
  });

  it("delegates albums.updated to the uid-scoped refetch", () => {
    const stub = newStub();

    onUpdate.call(stub, "albums.updated", { entities: ["album-1"] });

    expect(stub.refetchResults).toHaveBeenCalledWith(["album-1"]);
    expect(stub.refresh).not.toHaveBeenCalled();
    expect(warnSpy).not.toHaveBeenCalled();
  });

  it("inserts created albums without a full refresh", () => {
    const stub = newStub();

    onUpdate.call(stub, "albums.created", { entities: ["album-2"] });

    expect(stub.dirty).toBe(true);
    expect(stub.insertCreated).toHaveBeenCalledWith(["album-2"]);
    expect(stub.refresh).not.toHaveBeenCalled();
  });

  it("removes deleted albums from results and selection", () => {
    const stub = newStub();
    stub.results = [{ UID: "album-1" }, { UID: "album-2" }];

    onUpdate.call(stub, "albums.deleted", { entities: ["album-1"] });

    expect(stub.dirty).toBe(true);
    expect(stub.results.map((m) => m.UID)).toEqual(["album-2"]);
    expect(stub.removeSelection).toHaveBeenCalledWith("album-1");
  });

  it("ignores events when listen=false", () => {
    const stub = newStub();
    stub.listen = false;

    onUpdate.call(stub, "albums.updated", { entities: ["album-1"] });

    expect(stub.refetchResults).not.toHaveBeenCalled();
  });
});

describe("page/albums.vue refetchResults", () => {
  const refetchResults = PPageAlbums.methods.refetchResults;

  let searchSpy;
  afterEach(() => {
    searchSpy?.mockRestore();
  });

  it("patches affected loaded albums through one uid-scoped search", async () => {
    const stub = newStub();
    stub.results = [{ UID: "album-1", Title: "Old" }];
    searchSpy = vi.spyOn(Album, "search").mockResolvedValue({ models: [{ UID: "album-1", Title: "New" }] });

    refetchResults.call(stub, ["album-1", "album-not-loaded"]);
    await Promise.resolve();
    await Promise.resolve();

    expect(searchSpy).toHaveBeenCalledTimes(1);
    expect(searchSpy).toHaveBeenCalledWith({ uid: "album-1", count: 1 });
    expect(stub.results[0].Title).toBe("New");
    expect(stub.dirty).toBe(false);
  });

  it("removes albums the scoped search no longer returns", async () => {
    const stub = newStub();
    stub.results = [{ UID: "album-gone" }];
    searchSpy = vi.spyOn(Album, "search").mockResolvedValue({ models: [] });

    refetchResults.call(stub, ["album-gone"]);
    await Promise.resolve();
    await Promise.resolve();

    expect(stub.results).toEqual([]);
    expect(stub.removeSelection).toHaveBeenCalledWith("album-gone");
  });

  it("does nothing when no affected album is loaded", () => {
    const stub = newStub();
    searchSpy = vi.spyOn(Album, "search").mockResolvedValue({ models: [] });

    refetchResults.call(stub, ["album-1"]);

    expect(searchSpy).not.toHaveBeenCalled();
  });
});

describe("page/albums.vue insertCreated", () => {
  const insertCreated = PPageAlbums.methods.insertCreated;

  let searchSpy;
  afterEach(() => {
    searchSpy?.mockRestore();
  });

  it("prepends new albums matching the current view type", async () => {
    const stub = newStub();
    searchSpy = vi.spyOn(Album, "search").mockResolvedValue({ models: [{ UID: "album-new", Title: "New Album" }] });

    insertCreated.call(stub, ["album-new"]);
    await Promise.resolve();
    await Promise.resolve();

    expect(searchSpy).toHaveBeenCalledWith({ uid: "album-new", count: 1, type: "album" });
    expect(stub.results.map((m) => m.UID)).toEqual(["album-new"]);
  });

  it("does not duplicate albums that are already loaded", async () => {
    const stub = newStub();
    stub.results = [{ UID: "album-new" }];
    searchSpy = vi.spyOn(Album, "search").mockResolvedValue({ models: [{ UID: "album-new" }] });

    insertCreated.call(stub, ["album-new"]);
    await Promise.resolve();
    await Promise.resolve();

    expect(stub.results).toHaveLength(1);
  });
});
