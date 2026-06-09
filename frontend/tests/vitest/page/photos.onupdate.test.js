// Targets just the onUpdate switch in page/photos.vue and
// page/album/photos.vue. Booting the full SFC via @vue/test-utils
// would require Vuetify, the router, and the page's data() initial
// state — overkill for verifying dispatch on a single event type.
// Calling the Options API method directly with a stub `this`
// exercises the branch logic in isolation and pins the contract that
// photos.edited marks the cards list dirty without throwing on the
// unknown-event guard or warning to the console.
import { describe, it, expect, beforeEach, vi } from "vitest";

import PPagePhotos from "page/photos.vue";
import PAlbumPhotos from "page/album/photos.vue";

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

  it("marks the cards list dirty on photos.edited and keeps scroll position", () => {
    const stub = newStub();

    onUpdate.call(stub, "photos.edited", { entities: ["uid-1", "uid-2"] });

    expect(stub.dirty).toBe(true);
    expect(stub.complete).toBe(false);
    // photos.edited is a pure cache-stale signal — the cards list
    // should be flagged for refetch but no per-UID removeResult /
    // updateResults work should run.
    expect(stub.removeResult).not.toHaveBeenCalled();
    expect(stub.updateResults).not.toHaveBeenCalled();
    expect(warnSpy).not.toHaveBeenCalled();
  });

  it("ignores photos.edited when listen=false", () => {
    const stub = newStub();
    stub.listen = false;

    onUpdate.call(stub, "photos.edited", { entities: ["uid-1"] });

    expect(stub.dirty).toBe(false);
    expect(stub.complete).toBe(true);
  });

  it("short-circuits on malformed payloads without warning", () => {
    const stub = newStub();
    onUpdate.call(stub, "photos.edited", null);
    onUpdate.call(stub, "photos.edited", {});
    onUpdate.call(stub, "photos.edited", { entities: "not-an-array" });

    expect(stub.dirty).toBe(false);
    expect(warnSpy).not.toHaveBeenCalled();
  });

  it("still warns on unknown event types", () => {
    const stub = newStub();
    onUpdate.call(stub, "photos.unknown", { entities: ["uid-1"] });

    expect(warnSpy).toHaveBeenCalledTimes(1);
  });
});

describe("page/album/photos.vue onUpdate", () => {
  const onUpdate = PAlbumPhotos.methods.onUpdate;

  it("marks the album cards list dirty on photos.edited", () => {
    const stub = newStub();

    onUpdate.call(stub, "photos.edited", { entities: ["uid-1", "uid-2"] });

    expect(stub.dirty).toBe(true);
    expect(stub.complete).toBe(false);
    expect(stub.removeResult).not.toHaveBeenCalled();
    expect(stub.updateResults).not.toHaveBeenCalled();
  });

  it("ignores photos.edited when listen=false", () => {
    const stub = newStub();
    stub.listen = false;

    onUpdate.call(stub, "photos.edited", { entities: ["uid-1"] });

    expect(stub.dirty).toBe(false);
  });
});
