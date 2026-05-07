// Module-scope cache for the labels and albums typeahead lists used by
// the sidebar info panel, the batch-edit dialog, and the edit-dialog
// labels tab. Each consumer would otherwise re-fetch the same dataset
// on mount, costing repeated GET /api/v1/{labels,albums}?count=<cap>
// round-trips for the same browser session.
//
// API: getLabels() / getAlbums() each return a Promise<Array> of raw
// model instances. Concurrent callers share the same in-flight promise
// so exactly one request fires for any number of consumers. WS-driven
// invalidation evicts on labels.updated / labels.deleted / albums.updated
// / albums.deleted; the next read after invalidation triggers a fresh
// fetch. Album deletion currently only emits config.updated (no
// dedicated channel), so we also evict albums on config.updated to
// catch that case — the cost is one extra fetch per unrelated config
// change, which is bounded.
//
// The cache stays a client-side preload. Server-side debounced
// typeahead (search-as-you-type) is the right answer once libraries
// genuinely exceed the cap and lives in its own future proposal; this
// module would become its orchestrator at that point.
import Album from "model/album";
import Label from "model/label";
import $event from "common/event";

// Pragmatic ceiling shared by every consumer. Power users with more
// than CAP labels or albums see a console.warn and a truncated list;
// the long-term answer for those libraries is server-side debounced
// typeahead, not raising the cap further.
export const CAP = 5000;

const state = {
  labels: { data: null, fetch: null },
  albums: { data: null, fetch: null },
};

function evict(field) {
  const slot = state[field];
  if (!slot) return;
  slot.data = null;
  slot.fetch = null;
}

function fetchLabels() {
  return Label.search({ count: CAP, order: "name", all: true }).then((resp) => {
    const models = Array.isArray(resp?.models) ? resp.models : [];
    if (models.length === CAP) {
      console.warn(`Label.search returned ${CAP} results — list may be truncated.`);
    }
    return models;
  });
}

function fetchAlbums() {
  return Album.search({ count: CAP, order: "name", type: "album" }).then((resp) => {
    const models = Array.isArray(resp?.models) ? resp.models : [];
    if (models.length === CAP) {
      console.warn(`Album.search returned ${CAP} results — list may be truncated.`);
    }
    return models;
  });
}

function get(field, fetcher) {
  const slot = state[field];
  if (slot.data) return Promise.resolve(slot.data);
  if (slot.fetch) return slot.fetch;
  slot.fetch = fetcher()
    .then((data) => {
      slot.data = data;
      slot.fetch = null;
      return data;
    })
    .catch((err) => {
      slot.fetch = null;
      throw err;
    });
  return slot.fetch;
}

// Public surface — call-site agnostic. Consumers map the returned
// model arrays to whatever shape they need at the boundary.
export const typeaheadCache = {
  getLabels: () => get("labels", fetchLabels),
  getAlbums: () => get("albums", fetchAlbums),
  evictLabels: () => evict("labels"),
  evictAlbums: () => evict("albums"),
  clear: () => {
    evict("labels");
    evict("albums");
  },
};

// Backend publishes labels.updated / albums.updated through
// PublishLabelEvent / PublishAlbumEvent for create + update. Batch
// label deletion publishes labels.deleted via EntitiesDeleted. Album
// deletion does not publish a dedicated channel today — it only calls
// UpdateClientConfig() which fires config.updated, so we subscribe to
// that as the eviction signal for albums. Subscribing here at module
// scope (mirrors the photos.* pattern in model/photo.js) means every
// consumer benefits without per-component wiring.
$event.subscribe("labels.updated", () => evict("labels"));
$event.subscribe("labels.deleted", () => evict("labels"));
$event.subscribe("albums.updated", () => evict("albums"));
$event.subscribe("albums.deleted", () => evict("albums"));
$event.subscribe("config.updated", () => evict("albums"));

// Drop both lists on logout so user A's labels/albums cannot be
// served to user B inside the same tab. Mirrors Photo.clearCache()'s
// session.logout path in common/session.js (via deleteData → reset).
$event.subscribe("session.logout", () => {
  evict("labels");
  evict("albums");
});

export default typeaheadCache;
