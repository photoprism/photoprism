/*

Copyright (c) 2018 - 2025 PhotoPrism UG. All rights reserved.

    This program is free software: you can redistribute it and/or modify
    it under Version 3 of the GNU Affero General Public License (the "AGPL"):
    <https://docs.photoprism.app/license/agpl>

    This program is distributed in the hope that it will be useful,
    but WITHOUT ANY WARRANTY; without even the implied warranty of
    MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
    GNU Affero General Public License for more details.

    The AGPL is supplemented by our Trademark and Brand Guidelines,
    which describe how our Brand Assets may be used:
    <https://www.photoprism.app/trademark>

Feel free to send an email to hello@photoprism.app if you have questions,
want to support our work, or just want to say hello.

Additional information can be found in our Developer Guide:
<https://docs.photoprism.app/developer-guide/>

*/

// Deep-clones a plain object via JSON. Used at both ends of the cache
// lifecycle (set + hydrate) so callers can never share refs with cached
// values — see "isolation contract" in
// specs/frontend/model-lru-cache.md.
function deepClone(value) {
  if (value === null || typeof value !== "object") {
    return value;
  }
  return JSON.parse(JSON.stringify(value));
}

// ModelCacheStaleFetchError is raised by fetch() when clear() bumps
// the session-epoch counter while the loader is still in flight. It
// signals "the cache state this fetch belonged to is gone — discard
// the result." Existing callers absorb it via their .catch handlers;
// the export lets future callers discriminate this from real loader
// failures (network, auth, etc.) if they need to.
export class ModelCacheStaleFetchError extends Error {
  constructor(key) {
    super(`ModelCache: discarded stale fetch for "${key}" after clear()`);
    this.name = "ModelCacheStaleFetchError";
    this.key = key;
  }
}

// ModelCache is a small in-memory LRU for full-entity model snapshots. It
// is model-layer infrastructure: a subclass (e.g. Photo) supplies snapshot
// and hydrate hooks so the cache can stay neutral about model shape.
//
// Contract (see specs/frontend/model-lru-cache.md):
//   - Stores plain value snapshots, never live model instances.
//   - Returns a fresh hydrated instance for every cache hit so callers
//     can mutate freely without aliasing the cached source of truth.
//   - Coalesces concurrent fetch() calls for the same key behind one
//     in-flight Promise, then hydrates each waiter independently.
//   - LRU ordering: every read or refresh moves the entry to the most-
//     recent slot; size is capped at `max` and the oldest entry is
//     evicted when the cap is exceeded.
//   - Optional TTL (`ttl` ms). 0 / null / negative means "disabled" —
//     the recommended default for full-entity caches that lean on the
//     websocket update channel for freshness.
//   - clear() empties items and the pending-request map AND bumps an
//     internal session-epoch counter. In-flight loader Promises are
//     not literally aborted (we don't thread an AbortController
//     through every loader), but every fetch records the epoch at
//     start and a fetch whose epoch no longer matches REJECTS with
//     ModelCacheStaleFetchError instead of resolving. Net effect: a
//     logout-then-relogin sequence cannot serve data that was fetched
//     under the previous role — neither into the cache, nor through
//     a waiter's .then handler into UI state. Existing callers absorb
//     the rejection via their .catch handlers. Synchronous mutators
//     (set / refreshIfPresent) are not gated — they have no async
//     window to race against.
export class ModelCache {
  // Constructs a ModelCache with the given options:
  //   - max:      hard size cap; oldest entry is evicted on overflow.
  //   - ttl:      optional millisecond expiration; 0/null/negative disables.
  //   - snapshot: (model) => plain values to store. Required.
  //   - hydrate:  (values) => instantiated model. Required.
  //   - now:      () => epoch ms; injectable for deterministic TTL tests.
  constructor({ max = 50, ttl = 0, snapshot, hydrate, now = () => Date.now() } = {}) {
    if (typeof snapshot !== "function") {
      throw new Error("ModelCache: `snapshot` callback is required");
    }
    if (typeof hydrate !== "function") {
      throw new Error("ModelCache: `hydrate` callback is required");
    }
    this.max = max > 0 ? max : 0;
    this.ttl = typeof ttl === "number" && ttl > 0 ? ttl : 0;
    this.snapshot = snapshot;
    this.hydrate = hydrate;
    this.now = now;
    this.items = new Map();
    this.pending = new Map();
    // Monotonic counter bumped on clear(). fetch() and the direct
    // mutators capture this value at their entry point and compare
    // it before mutating items, so a Promise that started under epoch
    // N can never write into the cache after clear() advanced it to
    // N+1. See class-level docs.
    this._epoch = 0;
  }

  // Returns true if the key has a live (non-expired) entry. Lazy-prunes
  // an expired entry before reporting absence so size() / refreshIfPresent()
  // / oldest-eviction-on-overflow stay consistent with what readers can
  // actually retrieve. Does NOT change LRU ordering for live entries —
  // has() is a probe, not a touch.
  has(key) {
    if (!this.items.has(key)) {
      return false;
    }
    const entry = this.items.get(key);
    if (entry.expiresAt > 0 && entry.expiresAt <= this.now()) {
      this.items.delete(key);
      return false;
    }
    return true;
  }

  // Returns a freshly hydrated model for the given key, or null on miss
  // or expiration. Touching the entry promotes it to the most-recent LRU
  // slot. Hydration always happens against a deep clone of the snapshot
  // so the returned instance can be safely mutated.
  get(key) {
    const entry = this.items.get(key);
    if (!entry) {
      return null;
    }
    if (entry.expiresAt > 0 && entry.expiresAt <= this.now()) {
      this.items.delete(key);
      return null;
    }
    // LRU promotion: re-insert at the tail.
    this.items.delete(key);
    this.items.set(key, entry);
    return this.hydrate(deepClone(entry.value));
  }

  // Stores or refreshes the entry for `key`. The argument is routed
  // through the configured snapshot() callback so the cache always
  // holds normalized snapshot values, never live model instances —
  // even when callers pass models directly. The snapshot is also
  // deep-cloned before being stored so future caller-side mutations
  // on `value` cannot bleed into the cache. LRU ordering and size
  // cap are enforced. snapshot() must be idempotent for already-
  // snapshotted plain values; the Photo snapshot satisfies that
  // ("photo instanceof Photo ? photo.getValues() : photo").
  set(key, value) {
    if (this.max <= 0) {
      return;
    }
    const snapshot = this.snapshot(value);
    if (this.items.has(key)) {
      this.items.delete(key);
    } else if (this.items.size >= this.max) {
      // Reclaim every expired slot before evicting a live entry.
      // Without this pass an expired ghost can hold a slot until
      // overflow churn happens to evict it — and under non-uniform
      // LRU promotion a fresh entry can even get evicted ahead of
      // a stale one. The pass is O(n) but only runs at the cap and
      // is a no-op when ttl <= 0 (the Photo default).
      if (this.ttl > 0) {
        const cutoff = this.now();
        for (const [k, entry] of this.items) {
          if (entry.expiresAt > 0 && entry.expiresAt <= cutoff) {
            this.items.delete(k);
          }
        }
      }
      if (this.items.size >= this.max) {
        const oldest = this.items.keys().next().value;
        this.items.delete(oldest);
      }
    }
    this.items.set(key, {
      value: deepClone(snapshot),
      expiresAt: this.ttl > 0 ? this.now() + this.ttl : 0,
    });
  }

  // Returns a hydrated model for `key`, fetching via `loader` on miss.
  // Concurrent fetches for the same key share a single in-flight Promise;
  // each waiter receives an isolated hydrated instance. The loader is
  // expected to return a model whose values flow through `snapshot`.
  //
  // Epoch gate: the entry epoch is captured before the loader runs.
  // If clear() bumps the epoch while the loader is still in flight,
  // the returned Promise REJECTS with ModelCacheStaleFetchError so
  // waiters' .then handlers never see the stale value — caller-side
  // .catch() handlers absorb the rejection. The cache stays empty;
  // the next read goes back through the loader under the new epoch.
  // Rejecting (rather than resolving with the value) is what makes
  // the post-logout guarantee airtight: lightbox.vue / dialog.vue
  // both already chain a .catch(), so a stale role-A fetch can't
  // sneak data into UI mounted under role B.
  fetch(key, loader) {
    const cached = this.get(key);
    if (cached) {
      return Promise.resolve(cached);
    }
    if (this.pending.has(key)) {
      return this.pending.get(key).then((snapshot) => this.hydrate(deepClone(snapshot)));
    }
    const epoch = this._epoch;
    const request = Promise.resolve()
      .then(loader)
      .then((model) => {
        if (this._epoch !== epoch) {
          // clear() bumped the epoch while this loader was in
          // flight. Reject so the original waiter's .then doesn't
          // fire — otherwise a logout-then-relogin window could
          // route role-A data into role-B's UI before the route
          // change unmounts the component. set() will re-snapshot;
          // the snapshot callback is required to be idempotent for
          // plain snapshot values.
          throw new ModelCacheStaleFetchError(key);
        }
        const snapshot = this.snapshot(model);
        this.set(key, snapshot);
        return snapshot;
      })
      .finally(() => {
        // Only forget the pending entry if it still belongs to this
        // fetch. clear() already wiped pending; a new fetch on the
        // same key under the next epoch may have re-seeded it.
        if (this.pending.get(key) === request) {
          this.pending.delete(key);
        }
      });
    this.pending.set(key, request);
    return request.then((snapshot) => this.hydrate(deepClone(snapshot)));
  }

  // Refreshes an entry only if it is already cached AND still live.
  // Used by background event handlers (e.g. websocket "updated"
  // payloads) so the cache doesn't grow from traffic the user hasn't
  // actively browsed. The expired-entry probe goes through has() so
  // an entry past its TTL is pruned and reported as absent rather
  // than silently revived under a fresh expiry. The value is routed
  // through snapshot() via set() so callers can pass models or
  // pre-snapshotted plain values interchangeably. Returns true when
  // an entry was refreshed, false otherwise.
  refreshIfPresent(key, value) {
    if (!this.has(key)) {
      return false;
    }
    this.set(key, value);
    return true;
  }

  // Drops the entry (and any pending request) for the given key.
  evict(key) {
    this.items.delete(key);
    this.pending.delete(key);
  }

  // Empties the cache and the pending-request map, then bumps the
  // session-epoch counter. In-flight loader Promises are not literally
  // aborted, but a fetch that started under the previous epoch will
  // see the bump in its .then handler and skip the cache write — so
  // a logout-then-relogin sequence cannot leak data fetched under the
  // previous role even though its loader resolves after clear().
  clear() {
    this.items.clear();
    this.pending.clear();
    this._epoch++;
  }

  // Returns the current entry count. With TTL enabled this is a coarse
  // upper bound — expired entries are lazy-pruned on has() / get() /
  // refreshIfPresent(), not by a sweeper. For tests and debug counters
  // call has() (or query items.size after a representative read pass)
  // when an exact live count matters.
  size() {
    return this.items.size;
  }
}

export default ModelCache;
