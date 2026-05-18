// Module-scope cache for the labels and albums typeahead lists shared by the
// sidebar info panel, batch-edit dialog, and edit-dialog labels tab.
// getLabels / getAlbums return a Promise<Array> and coalesce concurrent callers
// onto a single in-flight request. WS mutation events evict the matching slot.
import Album from "model/album";
import Label from "model/label";
import $event, { subscribeEntityActions } from "common/event";

// Pragmatic ceiling shared by every consumer; over-cap libraries log a warning
// and are truncated. Server-side typeahead is the long-term answer.
export const CAP = 5000;

const state = {
  labels: { data: null, fetch: null },
  albums: { data: null, fetch: null },
};

function evict(field) {
  const slot = state[field];
  if (!slot) {
    return;
  }
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
  if (slot.data) {
    return Promise.resolve(slot.data);
  }
  if (slot.fetch) {
    return slot.fetch;
  }
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

// Public surface — consumers map the returned model arrays to whatever shape
// they need at the boundary.
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

// Evict on any standard mutation verb in the labels/albums namespace; the
// action only matters as a "something changed" signal, so payload is ignored.
subscribeEntityActions("labels", () => evict("labels"));
subscribeEntityActions("albums", () => evict("albums"));

// Belt-and-braces eviction for album mutations that surface only as a config
// reload (covers future config-touching mutations not on the entity channel).
$event.subscribe("config.updated", () => evict("albums"));

// Drop both lists on logout so user A's data cannot be served to user B in
// the same tab; mirrors Photo.clearCache()'s session.logout path.
$event.subscribe("session.logout", () => {
  evict("labels");
  evict("albums");
});

export default typeaheadCache;
