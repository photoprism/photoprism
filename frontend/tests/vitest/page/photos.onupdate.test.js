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
import $util from "common/util";

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

    // Only the loaded UID is requested, in one query. The count is a FILE
    // budget (affected photos x per-photo file headroom of 8), not a photo
    // count, so multi-file photos are not truncated out of the merged result.
    expect(searchSpy).toHaveBeenCalledTimes(1);
    expect(searchSpy).toHaveBeenCalledWith({ uid: "uid-1", merged: true, count: 8 });
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

  it("keeps a truncated multi-file photo selected when the scoped search hit the file limit", async () => {
    const stub = newStub();
    stub.results = [{ UID: "uid-1" }, { UID: "uid-2" }];
    // count >= limit marks the merged result as truncated: a still-visible multi-file
    // photo can be dropped from the response, so pruning must be skipped (#5738).
    searchSpy = vi.spyOn(Photo, "search").mockResolvedValue({ models: [{ UID: "uid-1", Title: "New" }], count: 8, limit: 8 });

    refetchResults.call(stub, ["uid-1", "uid-2"]);
    await Promise.resolve();
    await Promise.resolve();

    expect(stub.removeResult).not.toHaveBeenCalled();
    expect(stub.$clipboard.removeId).not.toHaveBeenCalled();
  });

  it("still prunes a removed photo when the limit was present but not reached", async () => {
    const stub = newStub();
    stub.results = [{ UID: "uid-gone" }];
    searchSpy = vi.spyOn(Photo, "search").mockResolvedValue({ models: [], count: 0, limit: 8 });

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

    // Count is a file budget (1 affected photo x per-photo headroom of 8), not a photo count.
    expect(searchSpy).toHaveBeenCalledWith({ uid: "uid-1", merged: true, count: 8 });
    expect(stub.results[0].Title).toBe("New");
  });

  it("keeps a truncated multi-file photo selected when the scoped search hit the file limit", async () => {
    const stub = newStub();
    stub.results = [{ UID: "uid-1" }, { UID: "uid-2" }];
    // count >= limit marks the merged result as truncated: a still-visible multi-file
    // photo can be dropped from the response, so pruning must be skipped (#5738).
    searchSpy = vi.spyOn(Photo, "search").mockResolvedValue({ models: [{ UID: "uid-1", Title: "New" }], count: 8, limit: 8 });

    refetchResults.call(stub, ["uid-1", "uid-2"]);
    await Promise.resolve();
    await Promise.resolve();

    expect(stub.removeResult).not.toHaveBeenCalled();
    expect(stub.$clipboard.removeId).not.toHaveBeenCalled();
  });

  it("still prunes a removed photo when the limit was present but not reached", async () => {
    const stub = newStub();
    stub.results = [{ UID: "uid-gone" }];
    searchSpy = vi.spyOn(Photo, "search").mockResolvedValue({ models: [], count: 0, limit: 8 });

    refetchResults.call(stub, ["uid-gone"]);
    await Promise.resolve();
    await Promise.resolve();

    expect(stub.removeResult).toHaveBeenCalledWith(stub.results, "uid-gone");
    expect(stub.$clipboard.removeId).toHaveBeenCalledWith("uid-gone");
  });
});

describe("page/album/photos.vue onAlbumsUpdated", () => {
  const onAlbumsUpdated = PAlbumPhotos.methods.onAlbumsUpdated;

  // Captures the surface of `this` that onAlbumsUpdated touches; the
  // album model reload resolves synchronously via a stubbed load().
  function newAlbumStub() {
    const stub = {
      albumGeneration: 0,
      lightbox: {},
      uid: "album-1",
      headerRevision: 0,
      $route: { query: {} },
      search: vi.fn(),
      listen: true,
      dirty: false,
      complete: true,
      scrollDisabled: true,
      lastParams: { order: "oldest" },
      collectionRoute: "albums",
      model: { UID: "album-1", Title: "Album", Order: "oldest" },
      $config: { get: () => "PhotoPrism" },
      $router: { push: vi.fn() },
      $util,
      updateQuery: vi.fn(),
      loadMore: vi.fn(),
    };
    stub.model.find = vi.fn(() => Promise.resolve(stub.model));
    return stub;
  }

  it("reloads the open album when reported as updated", async () => {
    const stub = newAlbumStub();

    onAlbumsUpdated.call(stub, "albums.updated", { entities: ["album-1"] });
    await Promise.resolve();

    expect(stub.model.find).toHaveBeenCalledTimes(1);
    expect(stub.headerRevision).toBe(1);
    expect(stub.dirty).toBe(true);
    expect(stub.complete).toBe(false);
    expect(window.document.title).toBe("PhotoPrism: \u2068Album\u2069");
    expect(stub.updateQuery).not.toHaveBeenCalled();
    expect(stub.loadMore).toHaveBeenCalledWith(true);
  });

  it("updates the query when the album order changed", async () => {
    const stub = newAlbumStub();
    stub.model.find = vi.fn(() => {
      stub.model.Order = "newest";
      return Promise.resolve(stub.model);
    });

    onAlbumsUpdated.call(stub, "albums.updated", { entities: ["album-1"] });
    await Promise.resolve();

    expect(stub.search).toHaveBeenCalledTimes(1);
    expect(stub.updateQuery).not.toHaveBeenCalled();
    expect(stub.loadMore).not.toHaveBeenCalled();
  });

  it("preserves a temporary order when the saved order changes", async () => {
    const stub = newAlbumStub();
    stub.$route.query = { order: "size", reverse: "true", q: "cat" };
    stub.$router.replace = vi.fn();
    stub.model.find = vi.fn(() => { stub.model.Order = "newest"; return Promise.resolve(stub.model); });
    await onAlbumsUpdated.call(stub, "albums.updated", { entities: ["album-1"] });
    expect(stub.search).not.toHaveBeenCalled();
    expect(stub.$router.replace).not.toHaveBeenCalled();
  });

  it("queues metadata changes while photos are loading", () => {
    const stub = newAlbumStub();
    stub.listen = false;
    onAlbumsUpdated.call(stub, "albums.updated", { entities: ["album-1"] });
    expect(stub.pendingAlbumUpdate).toBe(true);
    expect(stub.model.find).not.toHaveBeenCalled();
  });

  it("discards superseded metadata responses", async () => {
    const stub = newAlbumStub();
    const pending = [];
    stub.model.find = vi.fn(() => new Promise((resolve) => pending.push(resolve)));
    const oldRequest = onAlbumsUpdated.call(stub, "albums.updated", { entities: ["album-1"] });
    const newRequest = onAlbumsUpdated.call(stub, "albums.updated", { entities: ["album-1"] });
    pending[1]({ UID: "album-1", Title: "New", Order: "oldest" });
    await newRequest;
    pending[0]({ UID: "album-1", Title: "Old", Order: "oldest" });
    await oldRequest;
    expect(stub.model.Title).toBe("New");
    expect(stub.headerRevision).toBe(1);
  });

  it("ignores events for other albums", () => {
    const stub = newAlbumStub();

    onAlbumsUpdated.call(stub, "albums.updated", { entities: ["album-2"] });

    expect(stub.model.find).not.toHaveBeenCalled();
    expect(stub.dirty).toBe(false);
  });

  it("leaves the view when the scoped reload returns 404", async () => {
    const stub = newAlbumStub();
    stub.model.find = vi.fn(() => Promise.reject({ response: { status: 404 } }));

    onAlbumsUpdated.call(stub, "albums.updated", { entities: ["album-1"] });
    await Promise.resolve();
    await Promise.resolve();

    expect(stub.$router.push).toHaveBeenCalledWith({ name: "albums" });
    expect(stub.loadMore).not.toHaveBeenCalled();
  });

  it("marks the view dirty when the reload fails transiently", async () => {
    const stub = newAlbumStub();
    stub.model.find = vi.fn(() => Promise.reject(new Error("offline")));

    onAlbumsUpdated.call(stub, "albums.updated", { entities: ["album-1"] });
    await Promise.resolve();
    await Promise.resolve();

    expect(stub.$router.push).not.toHaveBeenCalled();
    expect(stub.dirty).toBe(true);
    expect(stub.complete).toBe(false);
    expect(stub.loadMore).not.toHaveBeenCalled();
  });
});

describe("page/album/photos.vue findAlbum", () => {
  const findAlbum = PAlbumPhotos.methods.findAlbum;

  // Captures the surface of `this` that findAlbum touches.
  function newFindStub(find) {
    return {
      albumGeneration: 0,
      uid: "album-1",
      collectionRoute: "albums",
      model: { find },
      $config: { get: () => "PhotoPrism" },
      $router: { push: vi.fn() },
      $util,
    };
  }

  it("resolves with the album and sets the title", async () => {
    const album = { UID: "album-1", Title: "Found" };
    const stub = newFindStub(vi.fn(() => Promise.resolve(album)));

    await expect(findAlbum.call(stub)).resolves.toBe(album);

    expect(stub.model).toBe(album);
    expect(window.document.title).toBe("PhotoPrism: \u2068Found\u2069");
    expect(stub.$router.push).not.toHaveBeenCalled();
  });
  it("leaves the view and resolves with null when the album is missing", async () => {
    const stub = newFindStub(vi.fn(() => Promise.reject(new Error("Request failed with status code 404"))));

    await expect(findAlbum.call(stub)).resolves.toBeNull();

    expect(stub.$router.push).toHaveBeenCalledWith({ name: "albums" });
  });
});

describe("page/album/photos.vue loadAlbum", () => {
  const loadAlbum = PAlbumPhotos.methods.loadAlbum;

  it("searches once the album was found", async () => {
    const stub = { uid: "album-1", findAlbum: vi.fn(() => Promise.resolve({ UID: "album-1" })), search: vi.fn() };

    await loadAlbum.call(stub);

    expect(stub.search).toHaveBeenCalledTimes(1);
  });
  it("skips the search when the album was not found", async () => {
    const stub = { findAlbum: vi.fn(() => Promise.resolve(null)), search: vi.fn() };

    await expect(loadAlbum.call(stub)).resolves.toBeUndefined();

    expect(stub.search).not.toHaveBeenCalled();
  });
});

describe("page/album/photos.vue cover refresh", () => {
  const { onCoverResolved, onCoverPhotoEvent } = PAlbumPhotos.methods;

  afterEach(() => vi.restoreAllMocks());

  it("tracks an automatic cover even when it is outside the current search", async () => {
    const scope = "as6sg6bxpogaaba8";
    const photo = "pqmxlr12pi8arvey";
    const hash = "a".repeat(40);
    const stub = { uid: scope, model: { UID: scope }, headerRevision: 0, results: [] };
    const request = vi.spyOn(Photo, "search").mockResolvedValue({ models: [{ UID: photo }] });
    await onCoverResolved.call(stub, hash);
    expect(request).toHaveBeenCalledWith({ s: scope, hash, count: 1, merged: true });
    onCoverPhotoEvent.call(stub, "photos.archived", { entities: [photo] });
    expect(stub.headerRevision).toBe(1);
  });

  it("retries the current cover lookup after a temporary failure", async () => {
    const scope = "as6sg6bxpogaaba8";
    const photo = "pqmxlr12pi8arvey";
    const hash = "a".repeat(40);
    const stub = { uid: scope, model: { UID: scope } };
    const request = vi.spyOn(Photo, "search").mockRejectedValueOnce(new Error("offline")).mockResolvedValueOnce({ models: [{ UID: photo }] });
    await onCoverResolved.call(stub, hash);
    await onCoverResolved.call(stub, hash);
    expect(request).toHaveBeenCalledTimes(2);
    expect(stub.headerPhotoUID).toBe(photo);
  });

  it("only refreshes the cover for newly matching album members", async () => {
    const scope = "as6sg6bxpogaaba8";
    const photo = "pqmxlr12pi8arvey";
    const stub = { uid: scope, model: { UID: scope }, albumGeneration: 1, headerRevision: 0, results: [] };
    const request = vi.spyOn(Photo, "search").mockResolvedValueOnce({ models: [] }).mockResolvedValueOnce({ models: [{ UID: photo }] });
    await onCoverPhotoEvent.call(stub, "photos.updated", { entities: [photo] });
    expect(stub.headerRevision).toBe(0);
    expect(request).toHaveBeenCalledWith({ s: scope, uid: photo, count: 1, merged: true });
    await onCoverPhotoEvent.call(stub, "photos.restored", { entities: [photo] });
    expect(stub.headerRevision).toBe(1);
  });

  it("discards a membership probe for an earlier album generation", async () => {
    const scope = "as6sg6bxpogaaba8";
    const photo = "pqmxlr12pi8arvey";
    const stub = { uid: scope, model: { UID: scope }, albumGeneration: 1, headerRevision: 0, results: [] };
    vi.spyOn(Photo, "search").mockResolvedValue({ models: [{ UID: photo }] });
    const pending = onCoverPhotoEvent.call(stub, "photos.updated", { entities: [photo] });
    stub.albumGeneration++;
    await pending;
    expect(stub.headerRevision).toBe(0);
  });
});


describe("page/album/photos.vue search lifecycle", () => {
  const { search, searchParams, loadMore, sortOrder, sortReverse } = PAlbumPhotos.methods;

  afterEach(() => vi.restoreAllMocks());

  // searchState configures an album query without rendering the page.
  function searchState(order = "oldest") {
    return {
      uid: "as6sg6bxpogaaba8", batchSize: 10, offset: 0, page: 0, searchGeneration: 0,
      lightbox: { results: [] }, results: [], filter: { q: "cat" }, lastFilter: { q: "cat" },
      model: { UID: "as6sg6bxpogaaba8", Order: "newest" },
      $route: { query: { order, reverse: "true" } },
      $view: { isHidden: () => false }, $notify: { warn: vi.fn(), info: vi.fn() },
      $gettext: (message) => message,
      sortOrder, sortReverse, searchParams, reset: vi.fn(), hideExpansionPanel: vi.fn(),
    };
  }

  it.each(["newest", "oldest"])("omits reverse when requesting Date Taken (%s)", async (order) => {
    const stub = searchState(order);
    const request = vi.spyOn(Photo, "search").mockResolvedValue({ count: 0, models: [] });
    await search.call(stub);
    expect(request.mock.calls[0][0]).toMatchObject({ s: stub.uid, q: "cat", order });
    expect(request.mock.calls[0][0]).not.toHaveProperty("reverse");
    expect(stub.results).toEqual([]);
    expect(stub.searchFailed).toBe(false);
    expect(stub.$notify.warn).toHaveBeenCalledWith("No pictures found");
  });

  it("does not reset the current search when an old request fails", async () => {
    const stub = searchState();
    let reject;
    vi.spyOn(Photo, "search").mockReturnValue(new Promise((resolve, fail) => { reject = fail; }));
    const pending = search.call(stub);
    stub.searchGeneration++;
    stub.dirty = true;
    reject(new Error("old search"));
    await pending;
    expect(stub.reset).not.toHaveBeenCalled();
    expect(stub.searchFailed).toBe(false);
    expect(stub.loading).toBe(true);
    expect(stub.listen).toBe(false);
    expect(stub.dirty).toBe(true);
  });

  it("does not merge a stale loadMore response into the next album", async () => {
    const stub = searchState();
    let resolve;
    vi.spyOn(Photo, "search").mockReturnValue(new Promise((done) => { resolve = done; }));
    const merge = vi.spyOn(Photo, "mergeResponse");
    const pending = loadMore.call(stub, true);
    stub.searchGeneration++;
    resolve({ count: 1, models: [{ UID: "pqmxlr12pi8arvey" }] });
    await pending;
    expect(merge).not.toHaveBeenCalled();
    expect(stub.results).toEqual([]);
    expect(stub.scrollDisabled).toBe(true);
    expect(stub.listen).toBe(false);
  });

  it("keeps scrolling disabled when an older loadMore request fails", async () => {
    const stub = searchState();
    let reject;
    vi.spyOn(Photo, "search").mockReturnValue(new Promise((resolve, fail) => { reject = fail; }));
    const pending = loadMore.call(stub, true);
    stub.searchGeneration++;
    stub.dirty = true;
    reject(new Error("old scroll request"));
    await pending;
    expect(stub.scrollDisabled).toBe(true);
    expect(stub.listen).toBe(false);
    expect(stub.dirty).toBe(true);
  });

  it("does not persist the URL view when the route updates after sorting", () => {
    const stub = searchState();
    stub.$route = { params: { album: stub.uid }, name: "album", query: { order: "title", view: "cards" } };
    stub.routeName = "album";
    stub.$view = { isActive: () => true, focus: vi.fn() };
    stub.settings = {};
    stub.getViewType = PAlbumPhotos.methods.getViewType;
    stub.search = vi.fn();
    const write = vi.spyOn(Storage.prototype, "setItem");
    PAlbumPhotos.watch.$route.call(stub);
    expect(stub.settings.view).toBe("cards");
    expect(stub.search).toHaveBeenCalledOnce();
    expect(write).not.toHaveBeenCalled();
  });

  it("restores route filters and pauses events before loading another album", () => {
    const stub = searchState();
    stub.$route = { params: { album: "as6sg6bxpogaaba7" }, name: "album", query: { q: "landscape" } };
    stub.$view = { isActive: () => true, focus: vi.fn() };
    stub.getViewType = () => "cards";
    stub.settings = {};
    stub.resetLastFilter = vi.fn();
    stub.albumGeneration = 0;
    stub.listen = true;
    stub.scrollDisabled = false;
    stub.loadAlbum = vi.fn(() => {
      expect(stub.uid).toBe("as6sg6bxpogaaba7");
      expect(stub.filter.q).toBe("landscape");
      expect(stub.listen).toBe(false);
      expect(stub.scrollDisabled).toBe(true);
    });
    PAlbumPhotos.watch.$route.call(stub);
    expect(stub.loadAlbum).toHaveBeenCalledOnce();
  });
});
