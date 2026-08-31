// Module-scope cache for the labels, albums, and people typeahead lists shared
// by the sidebar info panel, batch-edit dialog, edit-dialog tabs, and the
// people/face name pickers. getLabels / getAlbums / getPeople return a
// Promise<Array> and coalesce concurrent callers onto a single in-flight
// request. WS mutation events evict the matching slot, except when a people
// event only echoes a save this tab already merged.
import Album from "model/album";
import Label from "model/label";
import Subject from "model/subject";
import $event, { subscribeEntityActions, ACTION_CREATED, ACTION_UPDATED } from "common/event";

// Pragmatic ceiling shared by every consumer; over-cap libraries log a warning
// and are truncated. Server-side typeahead is the long-term answer.
export const CAP = 5000;

const state = {
  labels: { data: null, fetch: null, epoch: 0 },
  albums: { data: null, fetch: null, epoch: 0 },
  people: { data: null, fetch: null, epoch: 0 },
};

// Subject UIDs this tab merged via upsertPeople, with the time of the merge.
// The server echoes every save back over the websocket, and that echo would
// otherwise evict the slot the merge just updated.
const appliedPeople = new Map();

// Window in which an incoming people event still counts as our own echo.
export const SelfEchoTtl = 10000;

// People events are held this long before they evict. The websocket echo of a
// save routinely arrives before its HTTP response, and it is the merge that
// response triggers which identifies the event as this tab's own.
export const EchoGraceMs = 300;

// upsertPeople merges a save, so only a create or an update can be its echo.
// Any other verb names a change this tab cannot have made.
const SelfEchoActions = new Set([ACTION_CREATED, ACTION_UPDATED]);

const pendingPeople = new Set();
let pendingTimer = null;

// pruneAppliedPeople drops merge records that can no longer match an echo.
function pruneAppliedPeople(now) {
  for (const [uid, at] of appliedPeople) {
    if (now - at >= SelfEchoTtl) {
      appliedPeople.delete(uid);
    }
  }
}

// cancelPendingPeople drops a scheduled eviction that has become moot.
function cancelPendingPeople() {
  if (pendingTimer) {
    clearTimeout(pendingTimer);
    pendingTimer = null;
  }

  pendingPeople.clear();
}

// entityUids returns the UID list carried by an event payload, or null when the
// payload is not a recognizable, non-empty list of entity references.
function entityUids(data) {
  const entities = data?.entities;

  if (!Array.isArray(entities) || entities.length === 0) {
    return null;
  }

  const uids = entities.map((entity) => (typeof entity === "string" ? entity : entity?.UID));

  return uids.every((uid) => !!uid) ? uids : null;
}

// flushPendingPeople evicts unless every UID collected during the grace period
// was merged locally in the meantime, so one foreign UID still evicts.
function flushPendingPeople() {
  pendingTimer = null;

  const uids = [...pendingPeople];
  pendingPeople.clear();
  pruneAppliedPeople(Date.now());

  if (!uids.every((uid) => appliedPeople.has(uid))) {
    evict("people");
    return;
  }

  // A record answers for one echo only; a later change to the same person still
  // evicts, so the window cannot mask a second, foreign edit.
  uids.forEach((uid) => appliedPeople.delete(uid));
}

// Bumping the epoch invalidates any fetch that started before this point, so a
// request still in flight when the slot is evicted (or seeded) cannot resolve
// later and silently overwrite the slot with stale data. See get().
function evict(field) {
  const slot = state[field];
  if (!slot) {
    return;
  }
  slot.data = null;
  slot.fetch = null;
  slot.epoch += 1;

  if (field === "people") {
    appliedPeople.clear();
    cancelPendingPeople();
  }
}

// upsertPeople merges saved subjects into a populated people slot so a freshly
// created or renamed name is suggestible immediately, without waiting for the
// subjects/people WS event (which can lag a quick re-entry). It only mutates an
// already-loaded slot; a cold slot is left untouched so the next getPeople()
// fetches the full, server-ordered list. Entries are matched by UID, falling
// back to a case-insensitive name match so an existing person is updated rather
// than duplicated. The merge is idempotent: re-saving an unchanged name leaves
// the slot and its epoch untouched, so a redundant save cannot invalidate a
// pending fetch or replace the list with an equivalent copy.
//
// Contract: Marker.setName() and Face.setName() seed automatically. Callers that
// create or rename a person through any other path (e.g. Subject.update() from
// the recognized-people page) must call upsertPerson() themselves.
function upsertPeople(items) {
  const slot = state.people;
  if (!Array.isArray(slot.data) || !Array.isArray(items)) {
    return;
  }

  const next = slot.data.slice();
  const now = Date.now();
  let changed = false;

  pruneAppliedPeople(now);

  for (const raw of items) {
    const name = (raw?.Name || "").trim();
    if (!name) {
      continue;
    }

    const uid = raw?.UID || "";

    // Remember the save so its websocket echo does not evict what we merge here.
    if (uid) {
      appliedPeople.set(uid, now);
    }

    let idx = uid ? next.findIndex((p) => p && p.UID === uid) : -1;
    if (idx === -1) {
      idx = next.findIndex((p) => p && p.Name && p.Name.localeCompare(name, undefined, { sensitivity: "base" }) === 0);
    }

    if (idx >= 0) {
      const existing = next[idx];
      const mergedUid = uid || existing.UID;
      // Nothing to do when the slot already holds this exact UID and name.
      if (existing.UID === mergedUid && existing.Name === name) {
        continue;
      }
      next[idx] = { ...existing, UID: mergedUid, Name: name };
    } else {
      next.push({ UID: uid, Name: name });
    }
    changed = true;
  }

  if (changed) {
    slot.data = next;
    slot.epoch += 1;
  }
}

// evictPeopleUnlessSelfEcho handles a people mutation event. A verb that cannot
// be a local echo, or an unrecognized payload, evicts at once; a create or update
// carrying UIDs waits out the grace period so a merge can claim it.
function evictPeopleUnlessSelfEcho(data, action) {
  const uids = SelfEchoActions.has(action) ? entityUids(data) : null;

  if (!uids) {
    evict("people");
    return;
  }

  // Invalidate any fetch in flight right away, so a response captured before
  // this change cannot be cached while the grace period runs.
  state.people.fetch = null;
  state.people.epoch += 1;

  uids.forEach((uid) => pendingPeople.add(uid));

  if (!pendingTimer) {
    pendingTimer = setTimeout(flushPendingPeople, EchoGraceMs);
  }
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

function fetchPeople() {
  return Subject.search({ count: CAP, order: "name", type: "person" }).then((resp) => {
    const models = Array.isArray(resp?.models) ? resp.models : [];
    if (models.length === CAP) {
      console.warn(`Subject.search returned ${CAP} results — list may be truncated.`);
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
  // Snapshot the epoch so a fetch that finishes after an evict/upsert (which
  // bump the epoch) is discarded instead of caching now-stale results.
  const epoch = slot.epoch;
  slot.fetch = fetcher()
    .then((data) => {
      if (slot.epoch === epoch) {
        slot.data = data;
        slot.fetch = null;
      }
      return data;
    })
    .catch((err) => {
      if (slot.epoch === epoch) {
        slot.fetch = null;
      }
      throw err;
    });
  return slot.fetch;
}

// Public surface — consumers map the returned model arrays to whatever shape
// they need at the boundary.
export const typeaheadCache = {
  getLabels: () => get("labels", fetchLabels),
  getAlbums: () => get("albums", fetchAlbums),
  getPeople: () => get("people", fetchPeople),
  upsertPerson: (person) => upsertPeople([person]),
  evictLabels: () => evict("labels"),
  evictAlbums: () => evict("albums"),
  evictPeople: () => evict("people"),
  clear: () => {
    evict("labels");
    evict("albums");
    evict("people");
  },
};

// Evict on any standard mutation verb in the labels/albums namespace; the
// action only matters as a "something changed" signal, so payload is ignored.
subscribeEntityActions("labels", () => evict("labels"));
subscribeEntityActions("albums", () => evict("albums"));

// People mutations surface on both the subjects and people channels (a rename
// publishes subjects.updated and people.updated); either evicts the people slot
// unless the payload is the echo of a save this tab already merged.
subscribeEntityActions("subjects", (ev, data, action) => evictPeopleUnlessSelfEcho(data, action));
subscribeEntityActions("people", (ev, data, action) => evictPeopleUnlessSelfEcho(data, action));

// Belt-and-braces eviction for album mutations that surface only as a config
// reload (covers future config-touching mutations not on the entity channel).
$event.subscribe("config.updated", () => evict("albums"));

// Drop all lists on logout so user A's data cannot be served to user B in
// the same tab; mirrors Photo.clearCache()'s session.logout path.
$event.subscribe("session.logout", () => {
  evict("labels");
  evict("albums");
  evict("people");
});

export default typeaheadCache;
