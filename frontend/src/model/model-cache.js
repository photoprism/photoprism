/*

Copyright (c) 2018 - 2026 PhotoPrism UG. All rights reserved.

    This program is free software: you can redistribute it and/or modify
    it under Version 3 of the GNU Affero General Public License (the "AGPL"):
    <https://docs.photoprism.app/license/agpl>

    This program is distributed in the hope that it will be useful,
    but WITHOUT ANY WARRANTY; without even the implied warranty of
    MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
    GNU Affero General Public License for more details.

    The AGPL is supplemented by our Trademark and Brand Guidelines,
    which describe how our Brand Assets may be used:
    <https://www.photoprism.app/trademark/>

Feel free to send an email to hello@photoprism.app if you have questions,
want to support our work, or just want to say hello.

Additional information can be found in our Developer Guide:
<https://docs.photoprism.app/developer-guide/>

*/

// deepClone returns a JSON deep copy so cached values never share refs with callers.
function deepClone(value) {
  if (value === null || typeof value !== "object") {
    return value;
  }
  return JSON.parse(JSON.stringify(value));
}

// ModelCacheStaleFetchError signals that a fetch() result was discarded because
// clear() bumped the session-epoch counter while the loader was still in flight.
export class ModelCacheStaleFetchError extends Error {
  constructor(key) {
    super(`ModelCache: discarded stale fetch for "${key}" after clear()`);
    this.name = "ModelCacheStaleFetchError";
    this.key = key;
  }
}

// ModelCache is a small in-memory LRU for full-entity model snapshots; a
// subclass supplies snapshot/hydrate hooks so the cache stays shape-neutral.
// Contract:
//   - Stores plain snapshot values, never live model instances.
//   - Returns a fresh hydrated instance per cache hit (callers may mutate freely).
//   - Coalesces concurrent fetch() calls for the same key onto one in-flight Promise.
//   - LRU promotion on read/refresh; capped at `max`, oldest evicted on overflow.
//   - Optional TTL (`ttl` ms; 0/null/negative disables) for full-entity caches.
//   - clear() empties the cache and bumps an epoch counter so in-flight fetches
//     that started under the previous epoch REJECT with ModelCacheStaleFetchError
//     instead of leaking data across a logout/relogin boundary.
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
    // Monotonic counter bumped on clear(); see class-level docs.
    this._epoch = 0;
  }

  // has returns true if the key has a live (non-expired) entry; expired entries
  // are pruned before reporting absence. Probe only — does not promote LRU order.
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

  // get returns a freshly hydrated model for the key (or null on miss/expiry),
  // promoting the entry to the most-recent LRU slot.
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

  // set stores or refreshes the entry for `key`, routing the value through
  // snapshot() and a deep clone so the cache holds normalized plain values.
  // snapshot() must be idempotent for already-snapshotted inputs.
  set(key, value) {
    if (this.max <= 0) {
      return;
    }
    const snapshot = this.snapshot(value);
    if (this.items.has(key)) {
      this.items.delete(key);
    } else if (this.items.size >= this.max) {
      // Reclaim expired slots before evicting a live entry; no-op when ttl <= 0.
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

  // fetch returns a hydrated model for `key`, calling `loader` on miss and
  // sharing one in-flight Promise across concurrent waiters. If clear() bumps
  // the epoch mid-flight the returned Promise rejects with ModelCacheStaleFetchError
  // so stale role-A data cannot leak into role-B UI after a logout/relogin.
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
          // clear() advanced the epoch mid-flight — drop the stale result.
          throw new ModelCacheStaleFetchError(key);
        }
        const snapshot = this.snapshot(model);
        this.set(key, snapshot);
        return snapshot;
      })
      .finally(() => {
        // Only forget the pending entry if it still belongs to this fetch —
        // a post-clear() refetch on the same key may have re-seeded it.
        if (this.pending.get(key) === request) {
          this.pending.delete(key);
        }
      });
    this.pending.set(key, request);
    return request.then((snapshot) => this.hydrate(deepClone(snapshot)));
  }

  // refreshIfPresent updates an entry only if it is already cached and live,
  // so background event handlers don't grow the cache with un-browsed keys.
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

  // clear empties the cache and the pending-request map and bumps the epoch
  // counter so in-flight fetches under the previous epoch reject instead of writing.
  clear() {
    this.items.clear();
    this.pending.clear();
    this._epoch++;
  }

  // size returns the entry count; with TTL enabled this is a coarse upper
  // bound — expired entries are lazy-pruned on read, not by a sweeper.
  size() {
    return this.items.size;
  }
}

export default ModelCache;
