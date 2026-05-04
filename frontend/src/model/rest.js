import $api from "common/api";
import { Form } from "common/form";
import Model from "model.js";
import Link from "link.js";
import { $gettext } from "common/gettext";

export const BatchRestoreMultiplier = 10;

// Rest is the abstract base for models backed by RESTful API resources.
// It builds on Model's setValues/getValues/wasChanged/rollback contract
// and adds CRUD verbs (find/load/save/update/remove/restore), shared
// link helpers, and an opt-in cache surface (getCache/findCached/
// prefetch). Subclasses MUST override getCollectionResource() to point
// at the API path, and SHOULD override getModelName() and getDefaults().
export class Rest extends Model {
  // Returns the canonical entity identifier, preferring UID over ID
  // because UIDs are stable across instances while numeric IDs are
  // assigned by the database. Returns false when no id is set so callers
  // can branch on truthiness.
  getId() {
    if (this.UID) {
      return this.UID;
    }

    return this.ID ? this.ID : false;
  }

  // Convenience predicate around getId() for "has the model been
  // persisted yet" branching (e.g. save() chooses POST vs PUT).
  hasId() {
    return !!this.getId();
  }

  // Returns the URL slug for this entity, or an empty string when
  // unset. Used by createLink() and route helpers; never null so
  // template interpolation is safe.
  getSlug() {
    return this.Slug ? this.Slug : "";
  }

  // Returns a new instance of the same subclass populated with the
  // current field values. Useful for "edit a copy" flows that should
  // not share refs with the original model.
  clone() {
    return new this.constructor(this.getValues());
  }

  // Fetches the entity at `id` and returns a new instance of the same
  // subclass — note this DOES NOT mutate `this`. Used as the loader
  // for findCached(); callers that want to refresh `this` in place
  // should use load() instead. `params` flows through as the axios
  // request config.
  find(id, params) {
    return $api.get(this.getEntityResource(id), params).then((resp) => Promise.resolve(new this.constructor(resp.data)));
  }

  // Fetches this instance's own resource and refreshes its fields in
  // place via setValues(), which also resets __originalValues so
  // wasChanged() reads false again. No-op (returns undefined) when
  // the model has no id yet.
  load() {
    if (!this.hasId()) {
      return;
    }

    return $api.get(this.getEntityResource(this.getId())).then((resp) => Promise.resolve(this.setValues(resp.data)));
  }

  // Persists the model to the backend: PUT (delegates to update()) for
  // an existing entity, POST against the collection resource for a new
  // one. The response payload reseeds the local fields so server-
  // assigned values (id, timestamps) become visible to the caller.
  save() {
    if (this.hasId()) {
      return this.update();
    }

    return $api.post(this.constructor.getCollectionResource(), this.getValues()).then((resp) => Promise.resolve(this.setValues(resp.data)));
  }

  // Sends only the changed fields (computed via getValues(true)) as a
  // PUT to this entity's resource. Resolves to `this` without a request
  // when nothing has changed since the last load/save, so callers can
  // safely call update() unconditionally after edits. The response
  // payload reseeds the local fields and __originalValues, so a
  // subsequent wasChanged() reads false.
  update() {
    // Get updated values.
    const values = this.getValues(true);

    // Return if no values were changed.
    if (Object.keys(values).length === 0) {
      return Promise.resolve(this);
    }

    // Send PUT request.
    return $api.put(this.getEntityResource(), values).then((resp) => Promise.resolve(this.setValues(resp.data)));
  }

  // Issues a DELETE against this entity's resource. Resolves to `this`
  // unchanged — the model itself is not mutated, so callers can still
  // read its fields for audit/UI purposes after the request resolves.
  remove() {
    return $api.delete(this.getEntityResource()).then(() => Promise.resolve(this));
  }

  // Undeletes a soft-deleted entity by clearing DeletedAt server-side.
  // Like remove(), resolves to `this` without re-reading server state;
  // call load() afterwards if the latest fields are needed.
  restore() {
    return $api.put(this.getEntityResource(), { DeletedAt: null }).then(() => Promise.resolve(this));
  }

  // Returns the OPTIONS-derived edit form schema for this entity,
  // wrapped in common/form.Form. Used by inline-edit components to
  // discover field constraints (max length, allowed values, etc.).
  getEditForm() {
    return $api.options(this.getEntityResource()).then((resp) => Promise.resolve(new Form(resp.data)));
  }

  // Builds the API path for this entity. Falls back to the instance's
  // own id when `id` is omitted, so call sites that operate on `this`
  // can write `this.getEntityResource()` without a redundant argument.
  getEntityResource(id) {
    if (!id) {
      id = this.getId();
    }

    return this.constructor.getCollectionResource() + "/" + id;
  }

  // Returns a human-readable label of the form "ModelName id" suitable
  // for audit logs and confirmation dialogs. Falls back to "ModelName
  // false" when the model has no id; subclasses with a Title/Name
  // field typically override getEntityName() to return that instead.
  getEntityName() {
    return this.constructor.getModelName() + " " + this.getId();
  }

  // Creates a share-link for this entity. `password` and `expires` are
  // optional; both default to "no protection / no expiry" so the
  // template forms can pass empty values directly. Resolves to a fresh
  // Link instance hydrated from the server response.
  createLink(password, expires) {
    return $api
      .post(this.getEntityResource() + "/links", {
        Password: password ? password : "",
        Expires: expires ? expires : 0,
        Slug: this.getSlug(),
        Comment: "",
        Perm: 0,
      })
      .then((resp) => Promise.resolve(new Link(resp.data)));
  }

  // Persists changes to an existing share-link. Token and Password are
  // pulled from the link and forwarded explicitly because getValues(false)
  // returns all fields (including unchanged ones) but the backend treats
  // an empty Token as "do not rotate" and a missing Password as "leave
  // unchanged" — so the caller's intent has to be reconstructed here.
  updateLink(link) {
    let values = link.getValues(false);

    if (link.Token) {
      values["Token"] = link.getToken();
    }

    if (link.Password) {
      values["Password"] = link.Password;
    }

    return $api.put(this.getEntityResource() + "/links/" + link.getId(), values).then((resp) => Promise.resolve(link.setValues(resp.data)));
  }

  // Deletes the share-link and reseeds its local fields from the
  // server response so callers can still read the deleted state.
  removeLink(link) {
    return $api.delete(this.getEntityResource() + "/links/" + link.getId()).then((resp) => Promise.resolve(link.setValues(resp.data)));
  }

  // Lists every share-link attached to this entity. The response is
  // augmented with `models` (an array of hydrated Link instances) and
  // `count` so list-view components don't need to hydrate themselves.
  links() {
    return $api.get(this.getEntityResource() + "/links").then((resp) => {
      resp.models = [];
      resp.count = resp.data.length;

      for (let i = 0; i < resp.data.length; i++) {
        resp.models.push(new Link(resp.data[i]));
      }

      return Promise.resolve(resp);
    });
  }

  // Instance-level alias for the static getModelName(). Lets template
  // bindings such as `{{ photo.modelName() }}` skip the constructor hop.
  modelName() {
    return this.constructor.getModelName();
  }

  // Returns the API collection path (e.g. "photos", "albums") used by
  // search(), save(), and getEntityResource(). MUST be overridden by
  // every concrete subclass — the empty default exists only to make
  // the abstract-base intent explicit.
  static getCollectionResource() {
    // Needs to be implemented!
    return "";
  }

  // Returns the API path used by getCreateForm(). Defaults to the
  // collection resource so subclasses with the standard "POST to
  // /collection" create flow don't need to override it. Override when
  // creation lives at a sibling endpoint (e.g. "/users/register").
  static getCreateResource() {
    return this.getCollectionResource();
  }

  // Returns the OPTIONS-derived create-form schema, wrapped in
  // common/form.Form. Mirrors getEditForm() but targets the create
  // endpoint so client-side validation can match server expectations
  // before any entity exists.
  static getCreateForm() {
    return $api.options(this.getCreateResource()).then((resp) => Promise.resolve(new Form(resp.data)));
  }

  // Returns the i18n display name for this model type (e.g. "Photo",
  // "Album"). Used by getEntityName() and surfaced in dialogs/audit
  // text. Subclasses override to provide their own translation key.
  static getModelName() {
    return $gettext("Item");
  }

  // Returns the OPTIONS-derived search-form schema for this collection,
  // wrapped in common/form.Form. Drives the search UI's filter chip
  // construction so available filters track the backend.
  static getSearchForm() {
    return $api.options(this.getCollectionResource()).then((resp) => Promise.resolve(new Form(resp.data)));
  }

  // Hard upper bound on how many entities a single API call may touch.
  // Subclasses override to a smaller value when the underlying endpoint
  // enforces a tighter cap. Used by restoreCap() to clamp batch sizes.
  static limit() {
    return 100000;
  }

  // Computes how many entities a "restore from trash" batch may touch:
  // batchSize × multiplier, clamped to the subclass's limit(). Falls
  // back to the subclass's batchSize() when batchSize is not finite or
  // not positive. Returns 0 when no usable batch size can be derived,
  // signalling the caller to disable the action.
  static restoreCap(batchSize, multiplier = BatchRestoreMultiplier) {
    let size = Number(batchSize);

    if (!Number.isFinite(size) || size <= 0) {
      size = Number(this.batchSize ? this.batchSize() : 0);
    }

    if (!Number.isFinite(size) || size <= 0) {
      return 0;
    }

    const factor = Number(multiplier);
    const effectiveMultiplier = Number.isFinite(factor) && factor > 0 ? factor : BatchRestoreMultiplier;
    const cap = size * effectiveMultiplier;
    const limit = this.limit ? Number(this.limit()) : cap;

    if (Number.isFinite(limit) && limit > 0) {
      return Math.min(cap, limit);
    }

    return cap;
  }

  // getCache returns the ModelCache backing this subclass, or null when
  // the subclass does not opt in to caching. Default: no cache. Subclasses
  // that want findCached/prefetch behavior override this to return a
  // lazily-initialized ModelCache instance — see Photo.getCache for an
  // example. Per-subclass scoping is intentional so unrelated models do
  // not silently share one global budget or invalidation surface.
  static getCache() {
    return null;
  }

  // cacheKey converts an entity identifier into the string key used by
  // the ModelCache. The strict null-check matters: `String(id || "")`
  // collapses numeric `0` to `""`, which would alias every "no id" call
  // to the same slot.
  static cacheKey(id) {
    return id == null ? "" : String(id);
  }

  // findCached returns a hydrated model for the given id, fetching via
  // find() on cache miss. Falls through to a plain find() for subclasses
  // without a cache so callers can use this helper unconditionally.
  static findCached(id, params) {
    const cache = this.getCache();
    if (!cache) {
      return new this().find(id, params);
    }
    return cache.fetch(this.cacheKey(id), () => new this().find(id, params));
  }

  // prefetch is a fire-and-forget warm-up. Resolves to undefined so
  // callers cannot accidentally rely on the loaded value — use
  // findCached() when the value is needed. Returns immediately when
  // no cache is configured.
  static prefetch(id, params) {
    const cache = this.getCache();
    if (!cache) {
      return Promise.resolve();
    }
    return cache.fetch(this.cacheKey(id), () => new this().find(id, params)).then(() => undefined);
  }

  // Lists entities from the collection resource. `params` flows through
  // as the axios query string (count/offset/q/etc). The response is
  // augmented with `models` (an array of hydrated subclass instances),
  // `count`, `limit`, and `offset`, sourced from the X-Count, X-Limit,
  // and X-Offset response headers when present so paginated UIs can
  // tell when to load more. The raw `data` array is preserved so
  // callers that prefer plain JSON can still reach it.
  static search(params) {
    const options = {
      params: params,
    };

    return $api.get(this.getCollectionResource(), options).then((resp) => {
      let count = resp.data ? resp.data.length : 0;
      let limit = 0;
      let offset = 0;

      if (resp.headers) {
        if (resp.headers["x-count"]) {
          count = parseInt(resp.headers["x-count"]);
        }

        if (resp.headers["x-limit"]) {
          limit = parseInt(resp.headers["x-limit"]);
        }

        if (resp.headers["x-offset"]) {
          offset = parseInt(resp.headers["x-offset"]);
        }
      }

      resp.models = [];
      resp.count = count;
      resp.limit = limit;
      resp.offset = offset;

      if (count > 0) {
        for (let i = 0; i < resp.data.length; i++) {
          resp.models.push(new this(resp.data[i]));
        }
      }

      return Promise.resolve(resp);
    });
  }
}

export default Rest;
