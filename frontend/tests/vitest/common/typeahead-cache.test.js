import { describe, it, expect, beforeEach, vi } from "vitest";
import Album from "model/album";
import Label from "model/label";
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

  it("clear empties both lists", async () => {
    const labelSpy = vi
      .spyOn(Label, "search")
      .mockResolvedValueOnce({ models: [{ Name: "L1", UID: "1" }] })
      .mockResolvedValueOnce({ models: [{ Name: "L2", UID: "2" }] });
    const albumSpy = vi
      .spyOn(Album, "search")
      .mockResolvedValueOnce({ models: [{ Title: "A1", UID: "a1" }] })
      .mockResolvedValueOnce({ models: [{ Title: "A2", UID: "a2" }] });

    await typeaheadCache.getLabels();
    await typeaheadCache.getAlbums();
    typeaheadCache.clear();

    expect(await typeaheadCache.getLabels()).toEqual([{ Name: "L2", UID: "2" }]);
    expect(await typeaheadCache.getAlbums()).toEqual([{ Title: "A2", UID: "a2" }]);
    expect(labelSpy).toHaveBeenCalledTimes(2);
    expect(albumSpy).toHaveBeenCalledTimes(2);
  });

  it("session.logout clears both lists", async () => {
    const labelSpy = vi
      .spyOn(Label, "search")
      .mockResolvedValueOnce({ models: [{ Name: "L1", UID: "1" }] })
      .mockResolvedValueOnce({ models: [{ Name: "L2", UID: "2" }] });
    const albumSpy = vi
      .spyOn(Album, "search")
      .mockResolvedValueOnce({ models: [{ Title: "A1", UID: "a1" }] })
      .mockResolvedValueOnce({ models: [{ Title: "A2", UID: "a2" }] });

    await typeaheadCache.getLabels();
    await typeaheadCache.getAlbums();
    $event.publishSync("session.logout", {});

    expect(await typeaheadCache.getLabels()).toEqual([{ Name: "L2", UID: "2" }]);
    expect(await typeaheadCache.getAlbums()).toEqual([{ Title: "A2", UID: "a2" }]);
    expect(labelSpy).toHaveBeenCalledTimes(2);
    expect(albumSpy).toHaveBeenCalledTimes(2);
  });
});
