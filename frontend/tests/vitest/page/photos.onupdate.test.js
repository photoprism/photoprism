// Targets the onUpdate switch and the refetchResults helper in
// page/photos.vue and page/album/photos.vue. Booting the full SFC via
// @vue/test-utils would require Vuetify, the router, and the page's
// data() initial state — overkill for verifying dispatch on a single
// event type. Calling the Options API methods directly with a stub
// `this` exercises the branch logic in isolation and pins the contract
// that photos.updated patches affected loaded rows through one
// uid-scoped search instead of re-running the full result query.
import { describe, it, expect, beforeEach, afterEach, vi } from "vitest";

import PPagePhotos from "page/photos.vue";
import PAlbumPhotos from "page/album/photos.vue";
import { Photo } from "model/photo";

// Captures the surface of `this` that onUpdate touches. Lets us
// assert dirty/complete transitions without instantiating Vue.
function newStub() {
  return {
    listen: true,
    context: "photos",
    dirty: false,
    complete: true,
    scrollDisabled: true,
    results: [],
    lightbox: { results: [] },
    $clipboard: { removeId: vi.fn() },
    $forceUpdate: vi.fn(),
    removeResult: vi.fn(),
    updateResults: vi.fn(),
    refetchResults: vi.fn(),
    loadMore: vi.fn(),
    refresh: vi.fn(),
  };
}

describe("page/photos.vue onUpdate", () => {
  const onUpdate = PPagePhotos.methods.onUpdate;

  let warnSpy;
  beforeEach(() => {
    warnSpy = vi.spyOn(console, "warn").mockImplementation(() => {});
  });

  it("delegates photos.updated to the uid-scoped refetch", () => {
    const stub = newStub();

    onUpdate.call(stub, "photos.updated", { entities: ["uid-1", "uid-2"] });

    expect(stub.refetchResults).toHaveBeenCalledWith(["uid-1", "uid-2"]);
    // photos.updated must not re-run the full result query.
    expect(stub.refresh).not.toHaveBeenCalled();
    expect(warnSpy).not.toHaveBeenCalled();
  });

  it("ignores photos.updated when listen=false", () => {
    const stub = newStub();
    stub.listen = false;

    onUpdate.call(stub, "photos.updated", { entities: ["uid-1"] });

    expect(stub.refetchResults).not.toHaveBeenCalled();
  });

  it("short-circuits on malformed payloads without warning", () => {
    const stub = newStub();
    onUpdate.call(stub, "photos.updated", null);
    onUpdate.call(stub, "photos.updated", {});
    onUpdate.call(stub, "photos.updated", { entities: "not-an-array" });

    expect(stub.refetchResults).not.toHaveBeenCalled();
    expect(warnSpy).not.toHaveBeenCalled();
  });

  it("still warns on unknown event types", () => {
    const stub = newStub();
    onUpdate.call(stub, "photos.unknown", { entities: ["uid-1"] });

    expect(warnSpy).toHaveBeenCalledTimes(1);
  });
});

describe("page/photos.vue refetchResults", () => {
  const refetchResults = PPagePhotos.methods.refetchResults;

  let searchSpy;
  afterEach(() => {
    searchSpy?.mockRestore();
  });

  it("patches affected loaded photos through one uid-scoped search", async () => {
    const stub = newStub();
    stub.results = [{ UID: "uid-1", Title: "Old" }];
    stub.updateResults = PPagePhotos.methods.updateResults.bind(stub);
    searchSpy = vi.spyOn(Photo, "search").mockResolvedValue({ models: [{ UID: "uid-1", Title: "New", Quality: 2 }] });

    refetchResults.call(stub, ["uid-1", "uid-not-loaded"]);
    await Promise.resolve();
    await Promise.resolve();

    // Only the loaded UID is requested, in one query.
    expect(searchSpy).toHaveBeenCalledTimes(1);
    expect(searchSpy).toHaveBeenCalledWith({ uid: "uid-1", merged: true, count: 1 });
    expect(stub.results[0].Title).toBe("New");
    expect(stub.dirty).toBe(false);
  });

  it("does nothing when no affected photo is loaded", () => {
    const stub = newStub();
    searchSpy = vi.spyOn(Photo, "search").mockResolvedValue({ models: [] });

    refetchResults.call(stub, ["uid-1"]);

    expect(searchSpy).not.toHaveBeenCalled();
    expect(stub.dirty).toBe(false);
  });

  it("falls back to the dirty flag for oversized batches", () => {
    const stub = newStub();
    const uids = Array.from({ length: 51 }, (_, i) => `uid-${i}`);
    stub.results = uids.map((uid) => ({ UID: uid }));
    searchSpy = vi.spyOn(Photo, "search").mockResolvedValue({ models: [] });

    refetchResults.call(stub, uids);

    expect(searchSpy).not.toHaveBeenCalled();
    expect(stub.dirty).toBe(true);
    expect(stub.complete).toBe(false);
  });

  it("removes photos the scoped search no longer returns", async () => {
    const stub = newStub();
    stub.results = [{ UID: "uid-gone" }];
    searchSpy = vi.spyOn(Photo, "search").mockResolvedValue({ models: [] });

    refetchResults.call(stub, ["uid-gone"]);
    await Promise.resolve();
    await Promise.resolve();

    expect(stub.removeResult).toHaveBeenCalledWith(stub.results, "uid-gone");
    expect(stub.$clipboard.removeId).toHaveBeenCalledWith("uid-gone");
  });

  it("removes approved photos from the review context", async () => {
    const stub = newStub();
    stub.context = "review";
    stub.results = [{ UID: "uid-1", Quality: 1 }];
    searchSpy = vi.spyOn(Photo, "search").mockResolvedValue({ models: [{ UID: "uid-1", Quality: 3 }] });

    refetchResults.call(stub, ["uid-1"]);
    await Promise.resolve();
    await Promise.resolve();

    expect(stub.removeResult).toHaveBeenCalledWith(stub.results, "uid-1");
    expect(stub.$clipboard.removeId).toHaveBeenCalledWith("uid-1");
  });

  it("marks the results dirty when the refetch fails", async () => {
    const stub = newStub();
    stub.results = [{ UID: "uid-1" }];
    searchSpy = vi.spyOn(Photo, "search").mockRejectedValue(new Error("offline"));

    refetchResults.call(stub, ["uid-1"]);
    await Promise.resolve();
    await Promise.resolve();

    expect(stub.dirty).toBe(true);
    expect(stub.complete).toBe(false);
  });
});

describe("page/album/photos.vue onUpdate", () => {
  const onUpdate = PAlbumPhotos.methods.onUpdate;

  it("delegates photos.updated to the uid-scoped refetch", () => {
    const stub = newStub();

    onUpdate.call(stub, "photos.updated", { entities: ["uid-1"] });

    expect(stub.refetchResults).toHaveBeenCalledWith(["uid-1"]);
    expect(stub.refresh).not.toHaveBeenCalled();
  });

  it("ignores photos.updated when listen=false", () => {
    const stub = newStub();
    stub.listen = false;

    onUpdate.call(stub, "photos.updated", { entities: ["uid-1"] });

    expect(stub.refetchResults).not.toHaveBeenCalled();
  });
});

describe("page/album/photos.vue refetchResults", () => {
  const refetchResults = PAlbumPhotos.methods.refetchResults;

  let searchSpy;
  afterEach(() => {
    searchSpy?.mockRestore();
  });

  it("patches affected loaded photos through one uid-scoped search", async () => {
    const stub = newStub();
    stub.results = [{ UID: "uid-1", Title: "Old" }];
    stub.updateResults = PAlbumPhotos.methods.updateResults.bind(stub);
    searchSpy = vi.spyOn(Photo, "search").mockResolvedValue({ models: [{ UID: "uid-1", Title: "New" }] });

    refetchResults.call(stub, ["uid-1"]);
    await Promise.resolve();
    await Promise.resolve();

    expect(searchSpy).toHaveBeenCalledWith({ uid: "uid-1", merged: true, count: 1 });
    expect(stub.results[0].Title).toBe("New");
  });
});

describe("page/album/photos.vue onAlbumsUpdated", () => {
  const onAlbumsUpdated = PAlbumPhotos.methods.onAlbumsUpdated;

  // Captures the surface of `this` that onAlbumsUpdated touches; the
  // album model reload resolves synchronously via a stubbed load().
  function newAlbumStub() {
    return {
      listen: true,
      dirty: false,
      complete: true,
      scrollDisabled: true,
      lastParams: { order: "oldest" },
      collectionRoute: "albums",
      model: { UID: "album-1", Title: "Album", Order: "oldest", load: vi.fn(() => Promise.resolve()) },
      $config: { get: () => "PhotoPrism" },
      $router: { push: vi.fn() },
      updateQuery: vi.fn(),
      loadMore: vi.fn(),
    };
  }

  it("reloads the open album when reported as updated", async () => {
    const stub = newAlbumStub();

    onAlbumsUpdated.call(stub, "albums.updated", { entities: ["album-1"] });
    await Promise.resolve();

    expect(stub.model.load).toHaveBeenCalledTimes(1);
    expect(stub.dirty).toBe(true);
    expect(stub.complete).toBe(false);
    expect(stub.updateQuery).not.toHaveBeenCalled();
    expect(stub.loadMore).toHaveBeenCalledWith(true);
  });

  it("updates the query when the album order changed", async () => {
    const stub = newAlbumStub();
    stub.model.load = vi.fn(() => {
      stub.model.Order = "newest";
      return Promise.resolve();
    });

    onAlbumsUpdated.call(stub, "albums.updated", { entities: ["album-1"] });
    await Promise.resolve();

    expect(stub.updateQuery).toHaveBeenCalledTimes(1);
    expect(stub.loadMore).toHaveBeenCalledWith(true);
  });

  it("ignores events for other albums", () => {
    const stub = newAlbumStub();

    onAlbumsUpdated.call(stub, "albums.updated", { entities: ["album-2"] });

    expect(stub.model.load).not.toHaveBeenCalled();
    expect(stub.dirty).toBe(false);
  });

  it("leaves the view when the scoped reload returns 404", async () => {
    const stub = newAlbumStub();
    stub.model.load = vi.fn(() => Promise.reject({ response: { status: 404 } }));

    onAlbumsUpdated.call(stub, "albums.updated", { entities: ["album-1"] });
    await Promise.resolve();
    await Promise.resolve();

    expect(stub.$router.push).toHaveBeenCalledWith({ name: "albums" });
    expect(stub.loadMore).not.toHaveBeenCalled();
  });

  it("marks the view dirty when the reload fails transiently", async () => {
    const stub = newAlbumStub();
    stub.model.load = vi.fn(() => Promise.reject(new Error("offline")));

    onAlbumsUpdated.call(stub, "albums.updated", { entities: ["album-1"] });
    await Promise.resolve();
    await Promise.resolve();

    expect(stub.$router.push).not.toHaveBeenCalled();
    expect(stub.dirty).toBe(true);
    expect(stub.complete).toBe(false);
    expect(stub.loadMore).not.toHaveBeenCalled();
  });
});
