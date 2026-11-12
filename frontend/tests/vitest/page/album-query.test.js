import { describe, it, expect, vi, afterEach } from "vitest";
import PAlbumPhotos from "page/album/photos.vue";
import { Photo } from "model/photo";

// queryStub supplies route and query state without mounting the photo views.
function queryStub() {
  const stub = {
    uid: "as6sg6bxpogaaba8", filter: { q: "cat", country: "", camera: 0 }, settings: { view: "cards" },
    $route: { query: { q: "cat", order: "size", reverse: "true" } }, $router: { replace: vi.fn(), push: vi.fn() },
    updateFilter: PAlbumPhotos.methods.updateFilter,
  };
  return stub;
}

afterEach(() => vi.restoreAllMocks());

describe("album URL queries", () => {
  it("preserves temporary sorting when searching or clearing", () => {
    const stub = queryStub();
    PAlbumPhotos.methods.updateQuery.call(stub, { q: "zebra" });
    expect(stub.$router.replace).toHaveBeenLastCalledWith({ query: { q: "zebra", order: "size", reverse: "true", view: "cards" } });
    PAlbumPhotos.methods.updateQuery.call(stub, { q: "" });
    expect(stub.$router.replace).toHaveBeenLastCalledWith({ query: { order: "size", reverse: "true", view: "cards" } });
  });
  it("opens the map with album scope and the clicked location", () => {
    const stub = { ...queryStub(), hasPlaces: true, canAccessLibrary: true, results: [{ CellID: "47a851" }] };
    PAlbumPhotos.methods.openLocation.call(stub, 0);
    expect(stub.$router.push).toHaveBeenCalledWith({ name: "places_view", params: { s: stub.uid }, query: { q: "47a851" } });
  });
  it("reloads a sort-only change and ignores a superseded response", async () => {
    const deferred = [];
    vi.spyOn(Photo, "search").mockImplementation(() => new Promise((resolve) => deferred.push(resolve)));
    const stub = {
      ...queryStub(), searchGeneration: 0, batchSize: 60, lightbox: { open: false }, results: [],
      sortOrder: PAlbumPhotos.methods.sortOrder, sortReverse: PAlbumPhotos.methods.sortReverse,
      model: { Order: "newest" }, searchParams: PAlbumPhotos.methods.searchParams,
      $notify: { info: vi.fn(), warn: vi.fn() }, $gettext: (text) => text, hideExpansionPanel: vi.fn(),
      $nextTick: vi.fn(),
    };
    PAlbumPhotos.methods.search.call(stub);
    stub.$route.query.order = "title";
    PAlbumPhotos.methods.search.call(stub);
    expect(Photo.search).toHaveBeenCalledTimes(2);
    deferred[1]({ count: 1, models: [{ UID: "new-result" }] });
    await vi.waitFor(() => expect(stub.loading).toBe(false));
    deferred[0]({ count: 1, models: [{ UID: "old-result" }] });
    await Promise.resolve();
    expect(stub.results).toEqual([{ UID: "new-result" }]);
  });
});
