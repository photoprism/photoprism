import { describe, it, expect, beforeEach, vi } from "vitest";
import Album from "model/album";
import Label from "model/label";
import Subject from "model/subject";
import $event from "common/event";
import { typeaheadCache, CAP } from "common/typeahead-cache";

beforeEach(() => {
  typeaheadCache.clear();
  vi.restoreAllMocks();
});

describe("typeaheadCache.getLabels", () => {
  it("fetches labels on first call and caches the result", async () => {
    const models = [{ Name: "Cat", UID: "lbl-cat" }];
    const spy = vi.spyOn(Label, "search").mockResolvedValueOnce({ models });

    const first = await typeaheadCache.getLabels();
    const second = await typeaheadCache.getLabels();

    expect(first).toEqual(models);
    expect(second).toEqual(models);
    expect(spy).toHaveBeenCalledTimes(1);
    expect(spy).toHaveBeenCalledWith({ count: CAP, order: "name", all: true });
  });

  it("dedupes concurrent in-flight calls into one network request", async () => {
    let resolveFn;
    const spy = vi
      .spyOn(Label, "search")
      .mockImplementationOnce(() => new Promise((resolve) => (resolveFn = resolve)));

    const p1 = typeaheadCache.getLabels();
    const p2 = typeaheadCache.getLabels();
    expect(spy).toHaveBeenCalledTimes(1);

    const models = [{ Name: "Dog", UID: "lbl-dog" }];
    resolveFn({ models });
    await expect(p1).resolves.toEqual(models);
    await expect(p2).resolves.toEqual(models);
    expect(spy).toHaveBeenCalledTimes(1);
  });

  it("re-fetches after labels.updated WS event evicts the cache", async () => {
    const first = [{ Name: "First", UID: "lbl-1" }];
    const second = [{ Name: "Second", UID: "lbl-2" }];
    const spy = vi
      .spyOn(Label, "search")
      .mockResolvedValueOnce({ models: first })
      .mockResolvedValueOnce({ models: second });

    expect(await typeaheadCache.getLabels()).toEqual(first);
    $event.publishSync("labels.updated", { entities: [] });
    expect(await typeaheadCache.getLabels()).toEqual(second);
    expect(spy).toHaveBeenCalledTimes(2);
  });

  // Pins the create-channel subscription. Without this, freshly added
  // labels (e.g. typing 你好 in the sidebar combobox when no such label
  // exists yet) would stay invisible to subsequent typeahead consumers
  // until something else evicted the cache.
  it("re-fetches after labels.created WS event evicts the cache", async () => {
    const first = [{ Name: "Existing", UID: "lbl-1" }];
    const second = [
      { Name: "Existing", UID: "lbl-1" },
      { Name: "你好", UID: "lbl-2" },
    ];
    const spy = vi
      .spyOn(Label, "search")
      .mockResolvedValueOnce({ models: first })
      .mockResolvedValueOnce({ models: second });

    expect(await typeaheadCache.getLabels()).toEqual(first);
    $event.publishSync("labels.created", { entities: [{ UID: "lbl-2", Name: "你好" }] });
    expect(await typeaheadCache.getLabels()).toEqual(second);
    expect(spy).toHaveBeenCalledTimes(2);
  });

  it("re-fetches after labels.deleted WS event", async () => {
    const first = [{ Name: "First", UID: "lbl-1" }];
    const second = [];
    const spy = vi
      .spyOn(Label, "search")
      .mockResolvedValueOnce({ models: first })
      .mockResolvedValueOnce({ models: second });

    await typeaheadCache.getLabels();
    $event.publishSync("labels.deleted", { entities: ["lbl-1"] });
    expect(await typeaheadCache.getLabels()).toEqual(second);
    expect(spy).toHaveBeenCalledTimes(2);
  });

  it("warns when the response equals the cap", async () => {
    const models = Array.from({ length: CAP }, (_, i) => ({ Name: `lbl-${i}`, UID: `lbl-uid-${i}` }));
    vi.spyOn(Label, "search").mockResolvedValueOnce({ models });
    const warn = vi.spyOn(console, "warn").mockImplementation(() => {});
    await typeaheadCache.getLabels();
    expect(warn).toHaveBeenCalledWith(expect.stringContaining(`Label.search returned ${CAP} results`));
  });

  it("does not warn when the response is below the cap", async () => {
    vi.spyOn(Label, "search").mockResolvedValueOnce({ models: [{ Name: "x", UID: "y" }] });
    const warn = vi.spyOn(console, "warn").mockImplementation(() => {});
    await typeaheadCache.getLabels();
    expect(warn).not.toHaveBeenCalled();
  });

  it("returns an empty list when the response has no models", async () => {
    vi.spyOn(Label, "search").mockResolvedValueOnce({});
    expect(await typeaheadCache.getLabels()).toEqual([]);
  });

  it("propagates rejection and clears the in-flight slot", async () => {
    const err = new Error("boom");
    const spy = vi
      .spyOn(Label, "search")
      .mockRejectedValueOnce(err)
      .mockResolvedValueOnce({ models: [{ Name: "After", UID: "lbl-after" }] });

    await expect(typeaheadCache.getLabels()).rejects.toThrow("boom");
    expect(await typeaheadCache.getLabels()).toEqual([{ Name: "After", UID: "lbl-after" }]);
    expect(spy).toHaveBeenCalledTimes(2);
  });
});

describe("typeaheadCache.getAlbums", () => {
  it("fetches albums on first call and caches the result", async () => {
    const models = [{ Title: "Trip", UID: "alb-trip" }];
    const spy = vi.spyOn(Album, "search").mockResolvedValueOnce({ models });

    const first = await typeaheadCache.getAlbums();
    const second = await typeaheadCache.getAlbums();

    expect(first).toEqual(models);
    expect(second).toEqual(models);
    expect(spy).toHaveBeenCalledTimes(1);
    expect(spy).toHaveBeenCalledWith({ count: CAP, order: "name", type: "album" });
  });

  it("re-fetches after albums.updated WS event", async () => {
    const first = [{ Title: "First", UID: "alb-1" }];
    const second = [{ Title: "Second", UID: "alb-2" }];
    const spy = vi
      .spyOn(Album, "search")
      .mockResolvedValueOnce({ models: first })
      .mockResolvedValueOnce({ models: second });

    await typeaheadCache.getAlbums();
    $event.publishSync("albums.updated", { entities: [] });
    expect(await typeaheadCache.getAlbums()).toEqual(second);
    expect(spy).toHaveBeenCalledTimes(2);
  });

  // Mirrors the labels.created test: the entity layer publishes albums.created
  // via PublishUserEntities("albums", EntityCreated, …) from Album.Save(),
  // and the websocket writer strips the user.<uid>. prefix before relaying.
  // Without this subscription, brand-new albums would stay invisible to
  // every other typeahead consumer in the same browser session.
  it("re-fetches after albums.created WS event", async () => {
    const first = [{ Title: "Existing", UID: "alb-1" }];
    const second = [
      { Title: "Existing", UID: "alb-1" },
      { Title: "Trip", UID: "alb-2" },
    ];
    const spy = vi
      .spyOn(Album, "search")
      .mockResolvedValueOnce({ models: first })
      .mockResolvedValueOnce({ models: second });

    await typeaheadCache.getAlbums();
    $event.publishSync("albums.created", { entities: [{ UID: "alb-2", Title: "Trip" }] });
    expect(await typeaheadCache.getAlbums()).toEqual(second);
    expect(spy).toHaveBeenCalledTimes(2);
  });

  it("re-fetches after config.updated (album-delete signal)", async () => {
    const first = [{ Title: "First", UID: "alb-1" }];
    const second = [];
    const spy = vi
      .spyOn(Album, "search")
      .mockResolvedValueOnce({ models: first })
      .mockResolvedValueOnce({ models: second });

    await typeaheadCache.getAlbums();
    $event.publishSync("config.updated", {});
    expect(await typeaheadCache.getAlbums()).toEqual(second);
    expect(spy).toHaveBeenCalledTimes(2);
  });

  it("warns when the response equals the cap", async () => {
    const models = Array.from({ length: CAP }, (_, i) => ({ Title: `alb-${i}`, UID: `alb-uid-${i}` }));
    vi.spyOn(Album, "search").mockResolvedValueOnce({ models });
    const warn = vi.spyOn(console, "warn").mockImplementation(() => {});
    await typeaheadCache.getAlbums();
    expect(warn).toHaveBeenCalledWith(expect.stringContaining(`Album.search returned ${CAP} results`));
  });
});

describe("typeaheadCache.getPeople", () => {
  it("fetches people on first call and caches the result", async () => {
    const models = [{ Name: "Jane Doe", UID: "ps-1" }];
    const spy = vi.spyOn(Subject, "search").mockResolvedValueOnce({ models });

    const first = await typeaheadCache.getPeople();
    const second = await typeaheadCache.getPeople();

    expect(first).toEqual(models);
    expect(second).toEqual(models);
    expect(spy).toHaveBeenCalledTimes(1);
    expect(spy).toHaveBeenCalledWith({ count: CAP, order: "name", type: "person" });
  });

  it("dedupes concurrent in-flight calls into one network request", async () => {
    let resolveFn;
    const spy = vi
      .spyOn(Subject, "search")
      .mockImplementationOnce(() => new Promise((resolve) => (resolveFn = resolve)));

    const p1 = typeaheadCache.getPeople();
    const p2 = typeaheadCache.getPeople();
    expect(spy).toHaveBeenCalledTimes(1);

    const models = [{ Name: "John Roe", UID: "ps-2" }];
    resolveFn({ models });
    await expect(p1).resolves.toEqual(models);
    await expect(p2).resolves.toEqual(models);
    expect(spy).toHaveBeenCalledTimes(1);
  });

  it("re-fetches after people.updated WS event evicts the cache", async () => {
    const first = [{ Name: "First", UID: "ps-1" }];
    const second = [{ Name: "Second", UID: "ps-1" }];
    const spy = vi
      .spyOn(Subject, "search")
      .mockResolvedValueOnce({ models: first })
      .mockResolvedValueOnce({ models: second });

    expect(await typeaheadCache.getPeople()).toEqual(first);
    $event.publishSync("people.updated", { entities: ["ps-1"] });
    expect(await typeaheadCache.getPeople()).toEqual(second);
    expect(spy).toHaveBeenCalledTimes(2);
  });

  it("re-fetches after subjects.updated WS event evicts the cache", async () => {
    const first = [{ Name: "First", UID: "ps-1" }];
    const second = [{ Name: "Renamed", UID: "ps-1" }];
    const spy = vi
      .spyOn(Subject, "search")
      .mockResolvedValueOnce({ models: first })
      .mockResolvedValueOnce({ models: second });

    expect(await typeaheadCache.getPeople()).toEqual(first);
    $event.publishSync("subjects.updated", { entities: ["ps-1"] });
    expect(await typeaheadCache.getPeople()).toEqual(second);
    expect(spy).toHaveBeenCalledTimes(2);
  });

  it("tolerates an empty or missing models payload", async () => {
    vi.spyOn(Subject, "search").mockResolvedValueOnce({});
    expect(await typeaheadCache.getPeople()).toEqual([]);
  });
});

describe("typeaheadCache in-flight eviction (epoch guard)", () => {
  // A fetch that is still in flight when the slot is evicted must NOT populate
  // the cache when it later resolves; otherwise a websocket eviction landing
  // mid-request is silently undone and the next read serves stale data.
  it("discards a people fetch that resolves after an eviction", async () => {
    let resolveStale;
    const spy = vi
      .spyOn(Subject, "search")
      .mockImplementationOnce(() => new Promise((resolve) => (resolveStale = resolve)))
      .mockResolvedValueOnce({ models: [{ Name: "Fresh", UID: "ps-2" }] });

    const inFlight = typeaheadCache.getPeople();
    typeaheadCache.evictPeople();
    resolveStale({ models: [{ Name: "Stale", UID: "ps-1" }] });
    await inFlight;

    expect(await typeaheadCache.getPeople()).toEqual([{ Name: "Fresh", UID: "ps-2" }]);
    expect(spy).toHaveBeenCalledTimes(2);
  });

  it("discards a fetch evicted by a WS event before it resolves", async () => {
    let resolveStale;
    const spy = vi
      .spyOn(Subject, "search")
      .mockImplementationOnce(() => new Promise((resolve) => (resolveStale = resolve)))
      .mockResolvedValueOnce({ models: [{ Name: "Renamed", UID: "ps-1" }] });

    const inFlight = typeaheadCache.getPeople();
    $event.publishSync("subjects.updated", { entities: ["ps-1"] });
    resolveStale({ models: [{ Name: "Original", UID: "ps-1" }] });
    await inFlight;

    expect(await typeaheadCache.getPeople()).toEqual([{ Name: "Renamed", UID: "ps-1" }]);
    expect(spy).toHaveBeenCalledTimes(2);
  });
});

describe("typeaheadCache.upsertPerson", () => {
  it("adds a saved name to a loaded cache without refetching", async () => {
    const spy = vi.spyOn(Subject, "search").mockResolvedValueOnce({ models: [{ Name: "Alpha", UID: "ps-1" }] });

    await typeaheadCache.getPeople();
    typeaheadCache.upsertPerson({ UID: "ps-2", Name: "Beta" });

    expect(await typeaheadCache.getPeople()).toEqual([
      { Name: "Alpha", UID: "ps-1" },
      { UID: "ps-2", Name: "Beta" },
    ]);
    expect(spy).toHaveBeenCalledTimes(1);
  });

  it("updates an existing person in place by UID", async () => {
    vi.spyOn(Subject, "search").mockResolvedValueOnce({ models: [{ Name: "Old", UID: "ps-1" }] });

    await typeaheadCache.getPeople();
    typeaheadCache.upsertPerson({ UID: "ps-1", Name: "New" });

    expect(await typeaheadCache.getPeople()).toEqual([{ Name: "New", UID: "ps-1" }]);
  });

  it("de-duplicates by case-insensitive name when no UID is given", async () => {
    vi.spyOn(Subject, "search").mockResolvedValueOnce({ models: [{ Name: "Alpha", UID: "ps-1" }] });

    await typeaheadCache.getPeople();
    typeaheadCache.upsertPerson({ Name: "alpha" });

    expect(await typeaheadCache.getPeople()).toEqual([{ Name: "alpha", UID: "ps-1" }]);
  });

  it("leaves the slot untouched when re-saving an unchanged name (idempotent)", async () => {
    vi.spyOn(Subject, "search").mockResolvedValueOnce({ models: [{ Name: "Alpha", UID: "ps-1" }] });

    const before = await typeaheadCache.getPeople();
    typeaheadCache.upsertPerson({ UID: "ps-1", Name: "Alpha" });
    const after = await typeaheadCache.getPeople();

    // Same array reference proves the slot was not replaced and the epoch was
    // not bumped, so a redundant save cannot invalidate a pending fetch.
    expect(after).toBe(before);
  });

  it("replaces the slot when a save actually changes the list", async () => {
    vi.spyOn(Subject, "search").mockResolvedValueOnce({ models: [{ Name: "Alpha", UID: "ps-1" }] });

    const before = await typeaheadCache.getPeople();
    typeaheadCache.upsertPerson({ UID: "ps-2", Name: "Beta" });
    const after = await typeaheadCache.getPeople();

    expect(after).not.toBe(before);
    expect(after).toContainEqual({ UID: "ps-2", Name: "Beta" });
  });

  it("is a no-op when the people cache is cold", async () => {
    const spy = vi.spyOn(Subject, "search").mockResolvedValueOnce({ models: [{ Name: "Alpha", UID: "ps-1" }] });

    typeaheadCache.upsertPerson({ UID: "ps-2", Name: "Beta" });

    expect(await typeaheadCache.getPeople()).toEqual([{ Name: "Alpha", UID: "ps-1" }]);
    expect(spy).toHaveBeenCalledTimes(1);
  });

  it("ignores entries without a usable name", async () => {
    vi.spyOn(Subject, "search").mockResolvedValueOnce({ models: [{ Name: "Alpha", UID: "ps-1" }] });

    await typeaheadCache.getPeople();
    typeaheadCache.upsertPerson({ UID: "ps-2", Name: "   " });
    typeaheadCache.upsertPerson(null);

    expect(await typeaheadCache.getPeople()).toEqual([{ Name: "Alpha", UID: "ps-1" }]);
  });

  // The reported bug: two names saved in quick succession. The optimistic seed
  // keeps the second name visible before its WS event arrives, and the eventual
  // event still refreshes from server truth.
  it("keeps a quick second save visible before the WS eviction lands", async () => {
    const spy = vi
      .spyOn(Subject, "search")
      .mockResolvedValueOnce({ models: [{ Name: "Alpha", UID: "ps-1" }] })
      .mockResolvedValueOnce({
        models: [
          { Name: "Alpha", UID: "ps-1" },
          { Name: "Beta", UID: "ps-2" },
        ],
      });

    expect(await typeaheadCache.getPeople()).toEqual([{ Name: "Alpha", UID: "ps-1" }]);

    typeaheadCache.upsertPerson({ UID: "ps-2", Name: "Beta" });
    expect(await typeaheadCache.getPeople()).toEqual([
      { Name: "Alpha", UID: "ps-1" },
      { UID: "ps-2", Name: "Beta" },
    ]);
    expect(spy).toHaveBeenCalledTimes(1);

    $event.publishSync("subjects.created", { entities: [{ UID: "ps-2", Name: "Beta" }] });
    expect(await typeaheadCache.getPeople()).toEqual([
      { Name: "Alpha", UID: "ps-1" },
      { Name: "Beta", UID: "ps-2" },
    ]);
    expect(spy).toHaveBeenCalledTimes(2);
  });
});

describe("typeaheadCache.evict / clear", () => {
  it("evictLabels forces a fresh fetch on the next read", async () => {
    const spy = vi
      .spyOn(Label, "search")
      .mockResolvedValueOnce({ models: [{ Name: "A", UID: "1" }] })
      .mockResolvedValueOnce({ models: [{ Name: "B", UID: "2" }] });

    await typeaheadCache.getLabels();
    typeaheadCache.evictLabels();
    expect(await typeaheadCache.getLabels()).toEqual([{ Name: "B", UID: "2" }]);
    expect(spy).toHaveBeenCalledTimes(2);
  });

  it("clear empties all lists", async () => {
    const labelSpy = vi
      .spyOn(Label, "search")
      .mockResolvedValueOnce({ models: [{ Name: "L1", UID: "1" }] })
      .mockResolvedValueOnce({ models: [{ Name: "L2", UID: "2" }] });
    const albumSpy = vi
      .spyOn(Album, "search")
      .mockResolvedValueOnce({ models: [{ Title: "A1", UID: "a1" }] })
      .mockResolvedValueOnce({ models: [{ Title: "A2", UID: "a2" }] });
    const personSpy = vi
      .spyOn(Subject, "search")
      .mockResolvedValueOnce({ models: [{ Name: "P1", UID: "p1" }] })
      .mockResolvedValueOnce({ models: [{ Name: "P2", UID: "p2" }] });

    await typeaheadCache.getLabels();
    await typeaheadCache.getAlbums();
    await typeaheadCache.getPeople();
    typeaheadCache.clear();

    expect(await typeaheadCache.getLabels()).toEqual([{ Name: "L2", UID: "2" }]);
    expect(await typeaheadCache.getAlbums()).toEqual([{ Title: "A2", UID: "a2" }]);
    expect(await typeaheadCache.getPeople()).toEqual([{ Name: "P2", UID: "p2" }]);
    expect(labelSpy).toHaveBeenCalledTimes(2);
    expect(albumSpy).toHaveBeenCalledTimes(2);
    expect(personSpy).toHaveBeenCalledTimes(2);
  });

  it("session.logout clears all lists", async () => {
    const labelSpy = vi
      .spyOn(Label, "search")
      .mockResolvedValueOnce({ models: [{ Name: "L1", UID: "1" }] })
      .mockResolvedValueOnce({ models: [{ Name: "L2", UID: "2" }] });
    const albumSpy = vi
      .spyOn(Album, "search")
      .mockResolvedValueOnce({ models: [{ Title: "A1", UID: "a1" }] })
      .mockResolvedValueOnce({ models: [{ Title: "A2", UID: "a2" }] });
    const personSpy = vi
      .spyOn(Subject, "search")
      .mockResolvedValueOnce({ models: [{ Name: "P1", UID: "p1" }] })
      .mockResolvedValueOnce({ models: [{ Name: "P2", UID: "p2" }] });

    await typeaheadCache.getLabels();
    await typeaheadCache.getAlbums();
    await typeaheadCache.getPeople();
    $event.publishSync("session.logout", {});

    expect(await typeaheadCache.getLabels()).toEqual([{ Name: "L2", UID: "2" }]);
    expect(await typeaheadCache.getAlbums()).toEqual([{ Title: "A2", UID: "a2" }]);
    expect(await typeaheadCache.getPeople()).toEqual([{ Name: "P2", UID: "p2" }]);
    expect(labelSpy).toHaveBeenCalledTimes(2);
    expect(albumSpy).toHaveBeenCalledTimes(2);
    expect(personSpy).toHaveBeenCalledTimes(2);
  });
});

// Forward-compat coverage for the subscribeEntityActions refactor:
// every entity-mutation verb in ENTITY_MUTATIONS — including ones the
// backend does not currently emit on these channels — flows through
// the namespace-level subscriber and evicts without any per-channel
// wiring in this module. Non-mutation actions stay no-ops so an
// unrelated future event under the same namespace can't pull the
// cache out from under live consumers.
describe("subscribeEntityActions integration", () => {
  it("re-fetches after labels.restored (mutation verb without a current emitter)", async () => {
    const first = [{ Name: "A", UID: "1" }];
    const second = [{ Name: "B", UID: "2" }];
    const spy = vi
      .spyOn(Label, "search")
      .mockResolvedValueOnce({ models: first })
      .mockResolvedValueOnce({ models: second });

    await typeaheadCache.getLabels();
    $event.publishSync("labels.restored", { entities: ["1"] });
    expect(await typeaheadCache.getLabels()).toEqual(second);
    expect(spy).toHaveBeenCalledTimes(2);
  });

  it("re-fetches after albums.archived", async () => {
    const first = [{ Title: "First", UID: "alb-1" }];
    const second = [{ Title: "Second", UID: "alb-2" }];
    const spy = vi
      .spyOn(Album, "search")
      .mockResolvedValueOnce({ models: first })
      .mockResolvedValueOnce({ models: second });

    await typeaheadCache.getAlbums();
    $event.publishSync("albums.archived", { entities: ["alb-1"] });
    expect(await typeaheadCache.getAlbums()).toEqual(second);
    expect(spy).toHaveBeenCalledTimes(2);
  });

  it("ignores labels.* events whose action is not in ENTITY_MUTATIONS", async () => {
    const cached = [{ Name: "A", UID: "1" }];
    const spy = vi.spyOn(Label, "search").mockResolvedValueOnce({ models: cached });

    await typeaheadCache.getLabels();
    $event.publishSync("labels.merged", { entities: ["1"] });
    $event.publishSync("labels.viewed", {});
    expect(await typeaheadCache.getLabels()).toEqual(cached);
    expect(spy).toHaveBeenCalledTimes(1);
  });

  it("ignores albums.* events whose action is not in ENTITY_MUTATIONS", async () => {
    const cached = [{ Title: "First", UID: "alb-1" }];
    const spy = vi.spyOn(Album, "search").mockResolvedValueOnce({ models: cached });

    await typeaheadCache.getAlbums();
    $event.publishSync("albums.merged", { entities: ["alb-1"] });
    $event.publishSync("albums.viewed", {});
    expect(await typeaheadCache.getAlbums()).toEqual(cached);
    expect(spy).toHaveBeenCalledTimes(1);
  });
});
