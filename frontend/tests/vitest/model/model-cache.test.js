import { describe, it, expect, vi } from "vitest";
import "../fixtures";
import ModelCache, { ModelCacheStaleFetchError } from "model/model-cache";

// Minimal model-shaped object the tests can use without pulling in any
// real subclass. The snapshot is intentionally idempotent — it accepts
// either a fakeModel ({ values: {...} }) or already-snapshotted plain
// values — so the fixture mirrors Photo's "instanceof check, otherwise
// pass through" callback. Set() now routes through snapshot, so a non-
// idempotent callback would break every test that pre-snapshots raw
// values inline.
function buildCache(overrides = {}) {
  return new ModelCache({
    max: 3,
    ttl: 0,
    snapshot: (model) => (model && typeof model === "object" && Object.prototype.hasOwnProperty.call(model, "values") ? { ...model.values } : { ...model }),
    hydrate: (values) => ({ hydrated: true, values }),
    ...overrides,
  });
}

function fakeModel(uid, extra = {}) {
  return { values: { UID: uid, ...extra } };
}

// Lets the suite drive tick-by-tick. flushMicrotasks waits for all
// queued .then() callbacks; flushAndTick advances the manual clock too
// for TTL tests.
const flushMicrotasks = () => Promise.resolve();

describe("model/model-cache", () => {
  describe("constructor", () => {
    it("requires a snapshot callback", () => {
      expect(() => new ModelCache({ hydrate: (v) => v })).toThrow(/snapshot/);
    });

    it("requires a hydrate callback", () => {
      expect(() => new ModelCache({ snapshot: (m) => m.values })).toThrow(/hydrate/);
    });

    it("accepts numeric max and clamps non-positive values to 0", () => {
      const a = new ModelCache({ max: 5, snapshot: (m) => m, hydrate: (v) => v });
      const b = new ModelCache({ max: -1, snapshot: (m) => m, hydrate: (v) => v });
      expect(a.max).toBe(5);
      expect(b.max).toBe(0);
    });

    it("treats ttl <= 0 as disabled", () => {
      const a = new ModelCache({ ttl: 0, snapshot: (m) => m, hydrate: (v) => v });
      const b = new ModelCache({ ttl: -100, snapshot: (m) => m, hydrate: (v) => v });
      expect(a.ttl).toBe(0);
      expect(b.ttl).toBe(0);
    });
  });

  describe("snapshot routing", () => {
    // Pin the contract that set() always runs its argument through
    // the configured snapshot() callback — so direct callers passing
    // a model instance (Photo, anything with getValues()) end up with
    // the same normalized cache shape that fetch() would have stored.
    // Without this, a future call site like
    //   Photo._cache.set(uid, photoInstance)
    // would store the live instance and leak refs through hydrate().
    it("routes set() through the snapshot callback", () => {
      const snapshotCalls = [];
      const cache = new ModelCache({
        max: 3,
        snapshot: (model) => {
          snapshotCalls.push(model);
          return { ...model.values, _snapshotted: true };
        },
        hydrate: (values) => values,
      });
      cache.set("a", { values: { UID: "a", Title: "Loaded" } });
      expect(snapshotCalls).toHaveLength(1);
      const stored = cache.get("a");
      expect(stored.UID).toBe("a");
      expect(stored.Title).toBe("Loaded");
      expect(stored._snapshotted).toBe(true);
    });

    it("routes refreshIfPresent() through the snapshot callback", () => {
      const cache = new ModelCache({
        max: 3,
        snapshot: (model) => ({ ...model.values, _snapshotted: true }),
        hydrate: (values) => values,
      });
      cache.set("a", { values: { UID: "a", Title: "Old" } });
      const ok = cache.refreshIfPresent("a", { values: { UID: "a", Title: "New" } });
      expect(ok).toBe(true);
      const stored = cache.get("a");
      expect(stored.Title).toBe("New");
      expect(stored._snapshotted).toBe(true);
    });
  });

  describe("set + get", () => {
    it("stores a deep clone so caller-side mutations don't bleed into the cache", () => {
      const cache = buildCache();
      const model = fakeModel("a", { Title: "Original" });
      cache.set("a", cache.snapshot(model));
      // Mutate the source after set: the cache must hold the snapshot.
      model.values.Title = "Tampered";
      const hit = cache.get("a");
      expect(hit.values.Title).toBe("Original");
    });

    it("hydrates from a clone so caller-side mutations don't bleed back into the cache", () => {
      const cache = buildCache();
      cache.set("a", { UID: "a", Title: "Original" });
      const first = cache.get("a");
      first.values.Title = "Mutated";
      const second = cache.get("a");
      expect(second.values.Title).toBe("Original");
      // Different hydrated instances, no aliasing.
      expect(first).not.toBe(second);
    });

    it("returns null on miss without throwing", () => {
      const cache = buildCache();
      expect(cache.get("missing")).toBeNull();
    });

    it("get() promotes the entry to the most-recent LRU slot", () => {
      const cache = buildCache({ max: 3 });
      cache.set("a", { UID: "a" });
      cache.set("b", { UID: "b" });
      cache.set("c", { UID: "c" });
      // Touching "a" should move it to the tail.
      cache.get("a");
      const keys = [...cache.items.keys()];
      expect(keys[keys.length - 1]).toBe("a");
    });

    it("set() refresh moves the entry to the most-recent LRU slot", () => {
      const cache = buildCache({ max: 3 });
      cache.set("a", { UID: "a" });
      cache.set("b", { UID: "b" });
      cache.set("c", { UID: "c" });
      cache.set("a", { UID: "a", Title: "Refreshed" });
      const keys = [...cache.items.keys()];
      expect(keys[keys.length - 1]).toBe("a");
    });

    it("evicts the oldest entry when max is exceeded", () => {
      const cache = buildCache({ max: 3 });
      cache.set("a", { UID: "a" });
      cache.set("b", { UID: "b" });
      cache.set("c", { UID: "c" });
      cache.set("d", { UID: "d" });
      expect(cache.size()).toBe(3);
      expect(cache.has("a")).toBe(false);
      expect(cache.has("d")).toBe(true);
    });

    it("set() is a no-op when max <= 0", () => {
      const cache = buildCache({ max: 0 });
      cache.set("a", { UID: "a" });
      expect(cache.size()).toBe(0);
      expect(cache.get("a")).toBeNull();
    });
  });

  describe("has", () => {
    it("returns true for live entries and false after expiration", () => {
      let now = 1000;
      const cache = buildCache({ ttl: 500, now: () => now });
      cache.set("a", { UID: "a" });
      expect(cache.has("a")).toBe(true);
      now += 600;
      expect(cache.has("a")).toBe(false);
    });

    it("does not promote the entry to the most-recent slot", () => {
      const cache = buildCache({ max: 3 });
      cache.set("a", { UID: "a" });
      cache.set("b", { UID: "b" });
      cache.has("a");
      const keys = [...cache.items.keys()];
      // "b" should still be the most-recent because has() is read-only.
      expect(keys[keys.length - 1]).toBe("b");
    });
  });

  describe("ttl", () => {
    it("expires entries after ttl ms and returns null on get", () => {
      let now = 1000;
      const cache = buildCache({ ttl: 500, now: () => now });
      cache.set("a", { UID: "a", Title: "Fresh" });
      now += 100;
      expect(cache.get("a").values.Title).toBe("Fresh");
      now += 500;
      expect(cache.get("a")).toBeNull();
    });

    it("with ttl disabled (0) entries never expire on their own", () => {
      let now = 1000;
      const cache = buildCache({ ttl: 0, now: () => now });
      cache.set("a", { UID: "a", Title: "Forever" });
      now += 1_000_000_000;
      expect(cache.get("a").values.Title).toBe("Forever");
    });

    // Lazy-prune contract: has() removes the expired entry on probe so
    // size() / refreshIfPresent() / oldest-eviction don't keep
    // accounting for entries no reader can retrieve.
    it("has() lazy-prunes an expired entry and reports it as absent", () => {
      let now = 1000;
      const cache = buildCache({ ttl: 500, now: () => now });
      cache.set("a", { UID: "a" });
      cache.set("b", { UID: "b" });
      now += 600;
      expect(cache.has("a")).toBe(false);
      expect(cache.items.has("a")).toBe(false);
      // The other entry is also expired and gets pruned on its own probe.
      expect(cache.has("b")).toBe(false);
      expect(cache.size()).toBe(0);
    });

    // refreshIfPresent must not silently revive an expired entry under
    // a fresh expiry window — that would let a background event handler
    // resurrect data the user can no longer retrieve.
    it("refreshIfPresent() returns false for an expired entry instead of reviving it", () => {
      let now = 1000;
      const cache = buildCache({ ttl: 500, now: () => now });
      cache.set("a", { UID: "a", Title: "Old" });
      now += 600;
      const ok = cache.refreshIfPresent("a", { UID: "a", Title: "New" });
      expect(ok).toBe(false);
      // The expired entry should be gone — the probe pruned it.
      expect(cache.has("a")).toBe(false);
      expect(cache.size()).toBe(0);
    });

    // Without overflow-time pruning, an expired ghost (here "a", whose
    // LRU position got pushed past fresher entries by get() promotion)
    // would hold a slot through eviction churn — and under non-uniform
    // LRU promotion a fresh entry could even be evicted ahead of stale
    // ones. set() must reclaim every expired slot before considering
    // an oldest-eviction.
    it("set() reclaims every expired slot before evicting on overflow", () => {
      let now = 1000;
      const cache = buildCache({ max: 3, ttl: 500, now: () => now });
      cache.set("a", { UID: "a" }); // expires at 1500
      cache.set("b", { UID: "b" }); // expires at 1500
      cache.set("c", { UID: "c" }); // expires at 1500
      // Promote "a" to most-recent so the natural oldest-eviction
      // would drop "b", not "a", on the next overflow.
      cache.get("a");
      now = 1700; // a, b, c are all expired
      cache.set("d", { UID: "d" }); // fresh insert at the cap
      // Without the prune pass we'd evict a single "oldest" ghost ("b")
      // and leave "c" + "a" sitting in the LRU as expired ghosts. With
      // it, only the newly-inserted "d" survives.
      expect(cache.size()).toBe(1);
      expect(cache.has("d")).toBe(true);
      expect(cache.has("a")).toBe(false);
      expect(cache.has("b")).toBe(false);
      expect(cache.has("c")).toBe(false);
    });
  });

  describe("fetch", () => {
    it("returns a hydrated model on miss and stores a snapshot", async () => {
      const cache = buildCache();
      const loader = vi.fn().mockResolvedValue(fakeModel("a", { Title: "Loaded" }));
      const result = await cache.fetch("a", loader);
      expect(result.hydrated).toBe(true);
      expect(result.values.Title).toBe("Loaded");
      expect(cache.has("a")).toBe(true);
      expect(loader).toHaveBeenCalledTimes(1);
    });

    it("returns a hydrated model from cache on hit without invoking the loader", async () => {
      const cache = buildCache();
      cache.set("a", { UID: "a", Title: "Cached" });
      const loader = vi.fn();
      const result = await cache.fetch("a", loader);
      expect(result.values.Title).toBe("Cached");
      expect(loader).not.toHaveBeenCalled();
    });

    it("coalesces concurrent fetches behind a single in-flight request", async () => {
      const cache = buildCache();
      const loader = vi.fn().mockResolvedValue(fakeModel("a", { Title: "Loaded" }));
      const [r1, r2, r3] = await Promise.all([cache.fetch("a", loader), cache.fetch("a", loader), cache.fetch("a", loader)]);
      expect(loader).toHaveBeenCalledTimes(1);
      // Each waiter receives an isolated hydrated instance.
      expect(r1).not.toBe(r2);
      expect(r2).not.toBe(r3);
      expect(r1.values.Title).toBe("Loaded");
      expect(r2.values.Title).toBe("Loaded");
      expect(r3.values.Title).toBe("Loaded");
    });

    it("clears the pending entry after the loader resolves", async () => {
      const cache = buildCache();
      const loader = vi.fn().mockResolvedValue(fakeModel("a"));
      await cache.fetch("a", loader);
      expect(cache.pending.has("a")).toBe(false);
    });

    it("clears the pending entry even when the loader rejects", async () => {
      const cache = buildCache();
      const loader = vi.fn().mockRejectedValue(new Error("boom"));
      await expect(cache.fetch("a", loader)).rejects.toThrow("boom");
      expect(cache.pending.has("a")).toBe(false);
      expect(cache.has("a")).toBe(false);
    });

    it("hands waiters isolated clones when piggy-backing on an in-flight request", async () => {
      const cache = buildCache();
      // Use a deferred loader so the second fetch lands while the first is in flight.
      let resolveLoader;
      const loader = vi.fn().mockImplementation(() => new Promise((res) => (resolveLoader = res)));
      const p1 = cache.fetch("a", loader);
      const p2 = cache.fetch("a", loader);
      // fetch() invokes the loader inside a microtask, so wait one tick
      // before resolving — otherwise resolveLoader hasn't been assigned yet.
      await flushMicrotasks();
      resolveLoader(fakeModel("a", { Title: "Loaded" }));
      const [r1, r2] = await Promise.all([p1, p2]);
      // Mutating one waiter must not affect the other.
      r1.values.Title = "Mutated by r1";
      expect(r2.values.Title).toBe("Loaded");
    });
  });

  describe("refreshIfPresent", () => {
    it("updates an existing entry and returns true", () => {
      const cache = buildCache();
      cache.set("a", { UID: "a", Title: "Old" });
      const ok = cache.refreshIfPresent("a", { UID: "a", Title: "New" });
      expect(ok).toBe(true);
      expect(cache.get("a").values.Title).toBe("New");
    });

    it("does not seed a new entry and returns false when key is absent", () => {
      const cache = buildCache();
      const ok = cache.refreshIfPresent("a", { UID: "a", Title: "New" });
      expect(ok).toBe(false);
      expect(cache.has("a")).toBe(false);
    });

    it("promotes the refreshed entry to the most-recent LRU slot", () => {
      const cache = buildCache({ max: 3 });
      cache.set("a", { UID: "a" });
      cache.set("b", { UID: "b" });
      cache.set("c", { UID: "c" });
      cache.refreshIfPresent("a", { UID: "a", Title: "Refreshed" });
      const keys = [...cache.items.keys()];
      expect(keys[keys.length - 1]).toBe("a");
    });
  });

  describe("evict", () => {
    it("removes the entry and any pending request for the given key", async () => {
      const cache = buildCache();
      cache.set("a", { UID: "a" });
      // Start a pending fetch on a different key so we can confirm evict
      // only touches the targeted key.
      let resolveLoader;
      const loader = () => new Promise((res) => (resolveLoader = res));
      const inFlight = cache.fetch("b", loader);
      await flushMicrotasks();
      cache.evict("a");
      expect(cache.has("a")).toBe(false);
      expect(cache.pending.has("b")).toBe(true);
      resolveLoader(fakeModel("b"));
      await inFlight;
    });
  });

  describe("clear", () => {
    it("empties items and the pending-request map", () => {
      const cache = buildCache();
      cache.set("a", { UID: "a" });
      cache.set("b", { UID: "b" });
      // Seed a pending entry without resolving.
      cache.fetch("c", () => new Promise(() => {}));
      cache.clear();
      expect(cache.size()).toBe(0);
      expect(cache.pending.size).toBe(0);
    });

    it("rejects in-flight fetches that resolve after clear() bumps the epoch", async () => {
      // Fetches whose epoch no longer matches must REJECT, not resolve.
      // Two guarantees flow from this: (1) the cache stays empty so
      // a future read can't serve role-A data; (2) the original
      // waiter's .then never fires, so a caller chaining
      // "Photo.findCached(uid).then(p => this.photo = p)" can't
      // route role-A data into UI mounted under role B during the
      // post-logout window before the route-change unmount.
      const cache = buildCache();
      let resolveLoader;
      const loader = () => new Promise((res) => (resolveLoader = res));
      const inFlight = cache.fetch("a", loader);
      await flushMicrotasks();
      cache.clear();
      expect(cache.size()).toBe(0);
      resolveLoader(fakeModel("a", { Title: "Late" }));
      await expect(inFlight).rejects.toBeInstanceOf(ModelCacheStaleFetchError);
      expect(cache.has("a")).toBe(false);
      expect(cache.size()).toBe(0);
    });

    it("a fresh fetch after clear() runs under the new epoch and re-seeds normally", async () => {
      const cache = buildCache();
      let resolveStale;
      const staleLoader = () => new Promise((res) => (resolveStale = res));
      const stale = cache.fetch("a", staleLoader);
      await flushMicrotasks();
      cache.clear();
      // Start a second fetch on the same key after clear(). It runs
      // under the new epoch, so its result MUST land in the cache.
      const freshLoader = vi.fn().mockResolvedValue(fakeModel("a", { Title: "Fresh" }));
      const fresh = cache.fetch("a", freshLoader);
      // Resolve the stale loader after the new fetch is in flight.
      // The stale promise rejects (epoch mismatch); the fresh one runs
      // under the new epoch and resolves cleanly.
      resolveStale(fakeModel("a", { Title: "Stale" }));
      await expect(stale).rejects.toBeInstanceOf(ModelCacheStaleFetchError);
      const result = await fresh;
      expect(result.values.Title).toBe("Fresh");
      expect(cache.has("a")).toBe(true);
      expect(cache.get("a").values.Title).toBe("Fresh");
    });

    it("rejected in-flight fetches still drop their pending entry across clear()", async () => {
      const cache = buildCache();
      let rejectLoader;
      const loader = () => new Promise((_res, rej) => (rejectLoader = rej));
      const inFlight = cache.fetch("a", loader);
      await flushMicrotasks();
      cache.clear();
      rejectLoader(new Error("offline"));
      await expect(inFlight).rejects.toThrow("offline");
      // The stale fetch's pending slot must not linger after clear().
      expect(cache.pending.has("a")).toBe(false);
    });
  });

  describe("epoch", () => {
    it("starts at 0 and increments on each clear()", () => {
      const cache = buildCache();
      expect(cache._epoch).toBe(0);
      cache.clear();
      expect(cache._epoch).toBe(1);
      cache.clear();
      expect(cache._epoch).toBe(2);
    });
  });

  describe("size", () => {
    it("reports the current item count", () => {
      const cache = buildCache();
      expect(cache.size()).toBe(0);
      cache.set("a", { UID: "a" });
      cache.set("b", { UID: "b" });
      expect(cache.size()).toBe(2);
      cache.evict("a");
      expect(cache.size()).toBe(1);
    });
  });
});
