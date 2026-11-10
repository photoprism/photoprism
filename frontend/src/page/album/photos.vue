<template>
  <div ref="page" tabindex="-1" class="p-page p-page-album-photos" :class="$config.aclClasses('photos')">
    <p-album-toolbar
      ref="toolbar"
      :compact="headerCollapsed"
      :collapse-progress="headerProgress"
      :cover-style="headerStyle"
      :filter="filter"
      :album="model"
      :settings="settings"
      :refresh="refresh"
      :update-filter="updateFilter"
      :update-query="updateQuery"
      class="p-page__navigation"
    ></p-album-toolbar>

    <p-album-header :album="model" :revision="headerRevision" @appearance="headerStyle = $event" @collapsed="headerCollapsed = $event" @collapse-progress="headerProgress = $event"></p-album-header>

    <div v-if="loading" class="p-page__loading">
      <p-loading></p-loading>
    </div>
    <div v-else class="p-page__content">
      <div v-if="!results.length && !searchFailed" class="album-empty-state" role="status">
        <p>{{ $gettext("No pictures found") }}</p>
        <v-btn v-if="filter.q" class="action-album-clear-search" @click="updateQuery({ q: '' })">{{ $gettext("Clear Search") }}</v-btn>
      </div>
      <p-scroll :load-more="loadMore" :load-disabled="scrollDisabled" :load-distance="scrollDistance" :loading="loading"> </p-scroll>

      <p-photo-clipboard :refresh="refresh" :album="model" :context="contexts.Album"></p-photo-clipboard>

      <p-photo-view-mosaic
        v-if="results.length && settings.view === 'mosaic'"
        :context="contexts.Album"
        :photos="results"
        :select-mode="selectMode"
        :filter="filter"
        :album="model"
        :edit-photo="editPhoto"
        :open-photo="openPhoto"
        :is-shared-view="isShared"
      ></p-photo-view-mosaic>
      <p-photo-view-list
        v-else-if="results.length && settings.view === 'list'"
        :context="contexts.Album"
        :photos="results"
        :select-mode="selectMode"
        :filter="filter"
        :album="model"
        :open-photo="openPhoto"
        :edit-photo="editPhoto"
        :open-date="openDate"
        :open-location="openLocation"
        :is-shared-view="isShared"
      ></p-photo-view-list>
      <p-photo-view-cards
        v-else-if="results.length"
        :context="contexts.Album"
        :photos="results"
        :select-mode="selectMode"
        :filter="filter"
        :album="model"
        :open-photo="openPhoto"
        :edit-photo="editPhoto"
        :open-date="openDate"
        :open-location="openLocation"
        :is-shared-view="isShared"
      ></p-photo-view-cards>
    </div>
  </div>
</template>

<script>
import { Photo } from "model/photo";
import Album from "model/album";
import Thumb from "model/thumb";
import { ACTION_UPDATED, ACTION_DELETED, ACTION_ARCHIVED, ACTION_RESTORED } from "common/event";
import * as contexts from "options/contexts";
import PAlbumToolbar from "component/album/toolbar.vue";
import PAlbumHeader from "component/album/header.vue";
import { albumOrder, albumCoverTone } from "common/album-view";
import PPhotoClipboard from "component/photo/clipboard.vue";
import PPhotoViewCards from "component/photo/view/cards.vue";
import PPhotoViewMosaic from "component/photo/view/mosaic.vue";
import PPhotoViewList from "component/photo/view/list.vue";
import PScroll from "component/scroll.vue";
import PLoading from "component/loading.vue";
import { getAppStorage } from "common/storage";

const appStorage = getAppStorage();

// Maximum number of affected entities reloaded in place per event;
// larger batches set the dirty flag and are refetched lazily instead.
const maxLiveRefetch = 50;

// Per-photo file headroom for a scoped refetch. The photos search `count` is a
// FILE limit, not a photo limit, so the budget must exceed the photo count — a
// photo can own several files (e.g. a vector or PDF with a generated preview),
// and with a photo-sized budget those extra rows truncate still-visible photos
// out of the merged result, wrongly dropping them from the selection. Typical
// stacks hold ~3 files (photo + video + XMP), so 8 leaves ample margin while
// keeping the count under the API max even at the maxLiveRefetch batch cap;
// refetchResults' truncation guard backstops the rare deeper stack.
const maxRefetchFilesPerPhoto = 8;

export default {
  name: "PPageAlbumPhotos",
  components: {
    PLoading,
    PAlbumToolbar,
    PAlbumHeader,
    PPhotoClipboard,
    PPhotoViewCards,
    PPhotoViewMosaic,
    PPhotoViewList,
    PScroll,
  },
  props: {
    staticFilter: {
      type: Object,
      default: () => {},
    },
  },
  expose: ["onShortCut"],
  // data initializes album-scoped photo loading and URL state.
  data() {
    const uid = this.$route.params.album;
    const query = this.$route.query;
    const routeName = this.$route.name;
    const camera = query["camera"] ? parseInt(query["camera"]) : 0;
    const q = query["q"] ? query["q"] : "";
    const country = query["country"] ? query["country"] : "";
    const view = this.getViewType();
    const filter = { country: country, camera: camera, q: q };
    const settings = { view: view };
    const batchSize = Photo.batchSize();

    return {
      headerCollapsed: false,
      headerProgress: 0,
      headerStyle: albumCoverTone([]),
      headerRevision: 0,
      pendingAlbumUpdate: false,
      searchGeneration: 0,
      albumGeneration: 0,
      searchFailed: false,
      isShared: this.$config.deny("photos", "manage"),
      canEdit: this.$config.allow("photos", "update") && this.$config.feature("edit"),
      hasPlaces: this.$config.allow("places", "view") && this.$config.feature("places"),
      canSearchPlaces: this.$config.allow("places", "search") && this.$config.feature("places"),
      canAccessLibrary: this.$config.allow("photos", "access_library"),
      contexts,
      subscriptions: [],
      listen: false,
      dirty: false,
      complete: false,
      model: new Album(),
      uid: uid,
      results: [],
      scrollDisabled: true,
      scrollDistance: window.innerHeight * 4,
      batchSize: batchSize,
      offset: 0,
      page: 0,
      selection: this.$clipboard.selection,
      settings: settings,
      filter: filter,
      lastFilter: {},
      lastParams: {},
      routeName: routeName,
      collectionRoute: this.$route.meta?.collectionRoute ? this.$route.meta.collectionRoute : "albums",
      loading: true,
      lightbox: {
        results: [],
        loading: false,
        complete: false,
        open: false,
        dirty: false,
        batchSize: 6000,
      },
    };
  },
  computed: {
    selectMode: function () {
      return this.selection.length > 0;
    },
  },
  watch: {
    // $route restores search and ordering before loading the current album.
    $route() {
      if (!this.$view.isActive(this)) {
        return;
      }

      this.$view.focus(this.$refs?.page);

      const query = this.$route.query;

      this.filter.q = query["q"] ? query["q"] : "";
      this.filter.camera = query["camera"] ? parseInt(query["camera"]) : 0;
      this.filter.country = query["country"] ? query["country"] : "";
      this.settings.view = this.getViewType();

      // Reused routes must invalidate a query even when their visible filters match.
      const routeChanged = this.routeName !== this.$route.name;

      if (routeChanged) {
        this.resetLastFilter();
      }

      this.routeName = this.$route.name;

      if (this.uid !== this.$route.params.album) {
        this.searchGeneration++;
        this.albumGeneration++;
        this.pendingAlbumUpdate = false;
        this.resetLastFilter();
        this.reset();
        this.loading = true;
        this.uid = this.$route.params.album;
        this.loadAlbum();
      } else {
        this.search();
      }
    },
  },
  // created loads the album and subscribes to metadata and photo changes.
  created() {
    this.loadAlbum();

    this.subscriptions.push(this.$event.subscribe("albums.updated", (ev, data) => this.onAlbumsUpdated(ev, data)));
    this.subscriptions.push(this.$event.subscribe("albums.deleted", (ev, data) => this.onAlbumsDeleted(ev, data)));
    this.subscriptions.push(this.$event.subscribe("photos", (ev, data) => {
      if (!this.model.Thumb) {
        this.headerRevision++;
      }
      this.onUpdate(ev, data);
    }));

    this.subscriptions.push(
      this.$event.subscribe("lightbox.opened", () => {
        this.lightbox.open = true;
      })
    );
    this.subscriptions.push(
      this.$event.subscribe("lightbox.closed", () => {
        this.lightbox.open = false;
      })
    );

    this.subscriptions.push(this.$event.subscribe("touchmove.top", () => this.refresh()));
    this.subscriptions.push(this.$event.subscribe("touchmove.bottom", () => this.loadMore()));
  },
  mounted() {
    this.$view.enter(this, this.$refs?.page);
  },
  // beforeUnmount invalidates requests and releases event subscriptions.
  beforeUnmount() {
    this.albumGeneration++;
    this.searchGeneration++;
    for (let i = 0; i < this.subscriptions.length; i++) {
      this.$event.unsubscribe(this.subscriptions[i]);
    }
  },
  unmounted() {
    this.$view.leave(this);
  },
  methods: {
    // resetLastFilter invalidates the previous search signature.
    resetLastFilter() {
      this.lastSearch = null;
      this.lastFilter = {};
    },
    onShortCut(ev) {
      switch (ev.code) {
        case "KeyR":
          this.refresh();
          return true;
        case "KeyU":
          if (this.$config.allow("files", "upload") && this.$config.feature("upload")) {
            this.$event.publish("dialog.upload");
          }
          return true;
      }
    },
    hideExpansionPanel() {
      return this.$refs?.toolbar?.hideExpansionPanel();
    },
    getViewType() {
      let queryParam = this.$route.query["view"] ? this.$route.query["view"] : "";
      let defaultType = appStorage.getItem("photos.view");
      let storedType = appStorage.getItem("album.photos.view");

      if (queryParam) {
        appStorage.setItem("album.photos.view", queryParam);
        return queryParam;
      } else if (storedType) {
        return storedType;
      } else if (defaultType) {
        return defaultType;
      } else if (window.innerWidth < 960) {
        return "mosaic";
      }

      return "cards";
    },
    // sortOrder resolves the temporary URL order or the album default.
    sortOrder() {
      return albumOrder(this.$route.query, this.model?.Order).order;
    },
    // sortReverse permits reverse only for supported two-direction orders.
    sortReverse() {
      return albumOrder(this.$route.query, this.model?.Order).reverse;
    },
    openDate(index) {
      if (!this.canEdit) {
        return this.openPhoto(index);
      }

      const photo = this.results[index];

      if (!photo) {
        return;
      } else if (!photo.TakenAt || photo.TakenAt.length < 10) {
        this.editPhoto(index);
        return;
      }

      this.$router.push({ query: { q: "taken:" + photo.TakenAt.substring(0, 10) } });
    },
    // openLocation opens the album map without the current text search.
    openLocation(index) {
      if (!this.hasPlaces) {
        return;
      }

      const photo = this.results[index];

      if (!photo) {
        return;
      }

      if (!this.uid && this.canAccessLibrary && photo.CellID && photo.CellID !== "zz") {
        this.$router.push({ name: "places", query: { q: photo.CellID } });
      } else if (this.uid) {
        this.$router.push({ name: "places_view", params: { s: this.uid }, query: { q: photo.CellID } });
      }
    },
    editPhoto(index, tab) {
      if (!this.canEdit) {
        return this.openPhoto(index);
      }

      let selection = this.results.map((p) => {
        return p.getId();
      });

      // Open Edit Dialog
      this.$event.publish("dialog.edit", { selection, album: this.album, index, tab });
    },
    openPhoto(index, showMerged = false) {
      if (this.loading || !this.listen || this.lightbox.loading || !this.results[index]) {
        return false;
      }

      const selected = this.results[index];

      // Do not open as stack if there is only one JPEG or if multiple pictures are selected.
      if (this.selection.length > 0 || selected.jpegFiles().length < 2) {
        showMerged = false;
      }

      if (showMerged) {
        this.$lightbox.openModels(Thumb.fromFiles([selected]), 0, this.model);
      } else if (this.sortOrder() === "random") {
        this.$lightbox.openModels(Thumb.fromPhotos(this.results), index, this.model);
      } else {
        this.$lightbox.openView(this, index);
      }

      return true;
    },
    // loadMore refreshes or extends the current scoped result window.
    loadMore(force) {
      if (!force && (this.scrollDisabled || this.$view.isHidden(this))) {
        return;
      }

      this.scrollDisabled = true;
      this.listen = false;

      if (this.dirty) {
        this.lightbox.dirty = true;
      }

      const count = this.dirty ? (this.page + 2) * this.batchSize : this.batchSize;
      const offset = this.dirty ? 0 : this.offset;

      const params = {
        count: count,
        offset: offset,
        s: this.uid,
        merged: true,
        order: this.sortOrder(),
        ...(this.sortReverse() ? { reverse: true } : {}),
      };

      Object.assign(params, this.lastFilter);

      if (this.staticFilter) {
        Object.assign(params, this.staticFilter);
      }

      this.lastParams = params;

      const generation = this.searchGeneration;
      Photo.search(params)
        .then((response) => {
          if (generation !== this.searchGeneration) {
            return;
          }
          this.results = Photo.mergeResponse(this.results, response);
          this.complete = response.count < count;
          this.scrollDisabled = this.complete;

          if (this.complete) {
            this.offset = offset;
            if (this.results.length > 1) {
              if (!this.lightbox.open) {
                this.$notify.info(this.$gettext("%{n} pictures found", { n: this.results.length }));
              }
            }
          } else if (this.results.length >= Photo.limit()) {
            this.offset = offset;
            this.scrollDisabled = true;
            this.complete = true;
            if (!this.lightbox.open) {
              this.$notify.warn(this.$gettext("Can't load more, limit reached"));
            }
          } else {
            this.offset = offset + count;
            this.page++;
            this.$nextTick(() => {
              if (this.$root.$el.clientHeight <= window.document.documentElement.clientHeight + 300) {
                this.loadMore();
              }
            });
          }
        })
        .catch(() => {
          if (generation !== this.searchGeneration) {
            return;
          }
          this.scrollDisabled = false;
        })
        .finally(() => {
          if (generation !== this.searchGeneration) {
            return;
          }
          this.dirty = false;
          this.loading = false;
          this.listen = true;
          if (this.pendingAlbumUpdate) {
            this.pendingAlbumUpdate = false;
            this.onAlbumsUpdated("albums.updated", { entities: [this.uid] });
          }
        });
    },
    updateSettings(props) {
      if (!props || typeof props !== "object" || props.target) {
        return;
      }

      for (const [key, value] of Object.entries(props)) {
        if (!this.settings.hasOwnProperty(key)) {
          continue;
        }
        switch (typeof value) {
          case "string":
            this.settings[key] = value.trim();
            break;
          default:
            this.settings[key] = value;
        }

        appStorage.setItem("album.photos." + key, this.settings[key]);
      }
    },
    // updateFilter applies supported user search fields.
    updateFilter(props) {
      if (!props || typeof props !== "object" || props.target) {
        return;
      }

      for (const [key, value] of Object.entries(props)) {
        if (!this.filter.hasOwnProperty(key)) {
          continue;
        }
        switch (typeof value) {
          case "string":
            this.filter[key] = value.trim();
            break;
          default:
            this.filter[key] = value;
        }
      }
    },
    // updateQuery writes search changes while retaining temporary ordering.
    updateQuery(props) {
      this.updateFilter(props);

      const query = {
        ...this.$route.query,
        view: this.settings.view,
      };

      Object.assign(query, this.filter);

      for (let key in query) {
        if (query[key] === undefined || !query[key]) {
          delete query[key];
        }
      }

      if (JSON.stringify(this.$route.query) === JSON.stringify(query)) {
        return false;
      }

      this.$router.replace({ query: query });

      return true;
    },
    // searchParams builds the active album request without unsupported direction flags.
    searchParams() {
      const params = {
        count: this.batchSize,
        offset: this.offset,
        s: this.uid,
        merged: true,
        order: this.sortOrder(),
        ...(this.sortReverse() ? { reverse: true } : {}),
      };

      Object.assign(params, this.filter);

      if (this.staticFilter) {
        Object.assign(params, this.staticFilter);
      }

      return params;
    },
    refresh(props) {
      this.updateSettings(props);

      // Do not refresh results if the view is already loading
      // or should not be listening for events.
      if (this.loading || !this.listen) {
        return;
      }

      // Make sure enough results are loaded to maintain the scroll position.
      if (this.page > 2) {
        this.page = this.page - 1;
      } else {
        this.page = 1;
      }

      // Flag results as dirty and incomplete to force a refresh.
      this.dirty = true;
      this.complete = false;

      // Enable infinite scrolling if it was disabled.
      this.scrollDisabled = false;

      this.loadMore(true);
    },
    reset() {
      this.results = [];
      this.lightbox.results = [];
    },
    // search reloads changed album queries and ignores superseded responses.
    search() {
      if (this.lightbox.open) {
        return;
      }

      const key = { ...this.filter, s: this.uid, order: this.sortOrder(), reverse: this.sortReverse() };
      if (JSON.stringify(this.lastSearch) === JSON.stringify(key)) {
        // this.$nextTick(() => this.$emit("scrollRefresh"));
        return;
      }

      this.scrollDisabled = true;
      this.lastSearch = key;
      this.lastFilter = { ...this.filter };
      const generation = ++this.searchGeneration;
      this.searchFailed = false;

      this.offset = 0;
      this.page = 0;
      this.loading = true;
      this.listen = false;
      this.complete = false;

      const params = this.searchParams();

      this.lastParams = params;

      Photo.search(params)
        .then((response) => {
          if (generation !== this.searchGeneration) {
            return;
          }
          // Hide search toolbar expansion panel when matching pictures were found.
          if (this.offset === 0 && response.count > 0) {
            this.hideExpansionPanel();
          }

          this.offset = this.batchSize;
          this.results = response.models;
          this.lightbox.results = [];
          this.lightbox.complete = false;
          this.complete = response.count < this.batchSize;
          this.scrollDisabled = this.complete;

          if (this.complete) {
            if (!this.results.length) {
              this.$notify.warn(this.$gettext("No pictures found"));
            } else if (this.results.length === 1) {
              this.$notify.info(this.$gettext("One picture found"));
            } else {
              this.$notify.info(this.$gettext("%{n} pictures found", { n: this.results.length }));
            }
          } else {
            this.$nextTick(() => {
              if (this.$root.$el.clientHeight <= window.document.documentElement.clientHeight + 300) {
                this.loadMore();
              }
            });
          }
        })
        .catch(() => {
          if (generation !== this.searchGeneration) {
            return;
          }
          this.searchFailed = true;
          this.lastSearch = null;
          this.reset();
        })
        .finally(() => {
          if (generation !== this.searchGeneration) {
            return;
          }
          this.dirty = false;
          this.loading = false;
          this.listen = true;
          if (this.pendingAlbumUpdate) {
            this.pendingAlbumUpdate = false;
            this.onAlbumsUpdated("albums.updated", { entities: [this.uid] });
          }
        });
    },
    // loadAlbum finds the album and searches its photos once it was found.
    loadAlbum() {
      return this.findAlbum().then((m) => {
        if (m) {
          this.search();
        }
      });
    },
    // findAlbum fetches the album by its UID, or leaves the view and resolves with null if that fails.
    findAlbum() {
      const uid = this.uid;
      const generation = ++this.albumGeneration;
      return this.model
        .find(uid)
        .then((m) => {
          if (uid !== this.uid || generation !== this.albumGeneration) {
            return null;
          }
          this.model = m;

          window.document.title = `${this.$config.get("siteTitle")}: ${this.$util.isolate(this.model.Title)}`;

          return this.model;
        })
        .catch(() => {
          if (uid === this.uid && generation === this.albumGeneration) {
            this.$router.push({ name: this.collectionRoute });
          }
          return null;
        });
    },
    // onAlbumsUpdated reloads scoped album metadata without replacing the active photo view.
    onAlbumsUpdated(ev, data) {
      if (!Array.isArray(data?.entities) || !data.entities.includes(this.uid)) {
        return;
      }
      if (!this.listen) {
        this.pendingAlbumUpdate = true;
        return;
      }
      const uid = this.uid;
      const oldOrder = this.model.Order;
      const generation = ++this.albumGeneration;
      return this.model.find(uid).then((model) => {
        if (this.uid !== uid || generation !== this.albumGeneration) {
          return;
        }
        this.model = model;
        this.headerRevision++;
        this.lightbox.dirty = true;
        window.document.title = `${this.$config.get("siteTitle")}: ${this.$util.isolate(this.model.Title)}`;
        if (!this.$route.query.order && oldOrder !== this.model.Order) {
          this.search();
        } else {
          this.dirty = true;
          this.complete = false;
          this.scrollDisabled = false;
          this.loadMore(true);
        }
      }).catch((err) => {
        if (generation !== this.albumGeneration) {
          return;
        }
        if (this.uid === uid && err?.response?.status === 404) {
          this.$router.push({ name: this.collectionRoute });
        } else if (this.uid === uid) {
          this.dirty = true;
          this.complete = false;
        }
      });
    },
    onAlbumsDeleted(ev, data) {
      if (!this.listen) {
        return;
      }

      if (!data || !data.entities || !Array.isArray(data.entities)) {
        return;
      }

      const type = ev.split(".")[1];
      switch (type) {
        case ACTION_DELETED:
          if (data.entities.includes(this.uid)) {
            this.$notify.success(this.$gettext("Album deleted"));
            this.$router.push({ name: this.collectionRoute });
          }
          return;
      }
    },
    // updateResults patches the scalar fields of all loaded copies of
    // the given photo, so an edit reflects in the open view without
    // re-running the full result query.
    updateResults(entity) {
      this.results
        .filter((m) => m.UID === entity.UID)
        .forEach((m) => {
          for (let key in entity) {
            if (key !== "UID" && entity.hasOwnProperty(key) && entity[key] != null && typeof entity[key] !== "object") {
              m[key] = entity[key];
            }
          }
        });

      this.lightbox.results
        .filter((m) => m.UID === entity.UID)
        .forEach((m) => {
          for (let key in entity) {
            if (key !== "UID" && entity.hasOwnProperty(key) && entity[key] != null && typeof entity[key] !== "object") {
              m[key] = entity[key];
            }
          }
        });
    },
    // refetchResults reloads the affected photos through the scoped
    // search API and patches the loaded results in place, so a single
    // edit costs one uid-filtered query instead of re-running the full
    // result query. Larger batches fall back to the dirty flag and are
    // refetched lazily on the next return-to-view.
    refetchResults(uids) {
      const affected = uids.filter((uid) => this.results.some((m) => m.UID === uid) || this.lightbox.results.some((m) => m.UID === uid));

      if (affected.length === 0) {
        return;
      }

      if (affected.length > maxLiveRefetch) {
        this.dirty = true;
        this.complete = false;
        return;
      }

      Photo.search({ uid: affected.join("|"), merged: true, count: affected.length * maxRefetchFilesPerPhoto })
        .then((resp) => {
          const found = new Set();

          resp.models.forEach((values) => {
            found.add(values.UID);
            this.updateResults(values);
          });

          // Rows the scoped search no longer returns are not visible to this
          // session anymore — drop them like a full refresh would. Skip this
          // when the response hit the file limit (resp.count >= resp.limit):
          // a truncated result can omit a still-visible multi-file photo, and
          // pruning it would wrongly deselect it.
          const truncated = resp.limit > 0 && resp.count >= resp.limit;
          if (!truncated) {
            affected
              .filter((uid) => !found.has(uid))
              .forEach((uid) => {
                this.removeResult(this.results, uid);
                this.removeResult(this.lightbox.results, uid);
                this.$clipboard.removeId(uid);
              });
          }
        })
        .catch(() => {
          this.dirty = true;
          this.complete = false;
        });
    },
    removeResult(results, uid) {
      const index = results.findIndex((m) => m.UID === uid);

      if (index >= 0) {
        results.splice(index, 1);
      }
    },
    onUpdate(ev, data) {
      if (!this.listen) {
        return;
      }

      if (!data || !data.entities) {
        return;
      }

      const type = ev.split(".")[1];

      switch (type) {
        case ACTION_RESTORED:
          this.dirty = true;
          this.scrollDisabled = false;
          this.complete = false;

          this.loadMore();

          break;
        case ACTION_DELETED:
        case ACTION_ARCHIVED:
          this.dirty = true;
          this.complete = false;

          for (let i = 0; i < data.entities.length; i++) {
            const uid = data.entities[i];

            this.removeResult(this.results, uid);
            this.removeResult(this.lightbox.results, uid);
            this.$clipboard.removeId(uid);
          }

          break;
        case ACTION_UPDATED:
          // photos.updated is a lightweight UID-only signal that carries
          // no entity fields; affected photos are reloaded through the
          // scoped search API and patched in place. The model-layer
          // subscriber in model/photo.js handles the lightbox-cache
          // eviction independently.
          this.refetchResults(data.entities);

          break;
      }

      // TODO: Needed?
      this.$forceUpdate();
    },
  },
};
</script>
