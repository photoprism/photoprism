<template>
  <div
    ref="page"
    tabindex="1"
    class="p-page p-page-album-photos"
    :class="$config.aclClasses('photos')"
    @keydown.ctrl="onCtrl"
  >
    <p-album-toolbar
      ref="toolbar"
      :filter="filter"
      :album="model"
      :settings="settings"
      :refresh="refresh"
      :update-filter="updateFilter"
      :update-query="updateQuery"
      class="p-page__navigation"
    ></p-album-toolbar>

    <div v-if="loading" class="p-page__loading">
      <p-loading></p-loading>
    </div>
    <div v-else class="p-page__content">
      <p-scroll
        :load-more="loadMore"
        :load-disabled="scrollDisabled"
        :load-distance="scrollDistance"
        :loading="loading"
      >
      </p-scroll>

      <p-photo-clipboard :refresh="refresh" :album="model" context="album"></p-photo-clipboard>

      <p-photo-view-mosaic
        v-if="settings.view === 'mosaic'"
        context="album"
        :photos="results"
        :select-mode="selectMode"
        :filter="filter"
        :album="model"
        :edit-photo="editPhoto"
        :open-photo="openPhoto"
        :is-shared-view="isShared"
      ></p-photo-view-mosaic>
      <p-photo-view-list
        v-else-if="settings.view === 'list'"
        context="album"
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
        v-else
        context="album"
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
import PAlbumToolbar from "component/album/toolbar.vue";
import PPhotoClipboard from "component/photo/clipboard.vue";
import PPhotoViewCards from "component/photo/view/cards.vue";
import PPhotoViewMosaic from "component/photo/view/mosaic.vue";
import PPhotoViewList from "component/photo/view/list.vue";
import PScroll from "component/scroll.vue";
import PLoading from "component/loading.vue";

export default {
  name: "PPageAlbumPhotos",
  components: {
    PLoading,
    PAlbumToolbar,
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
      isShared: this.$config.deny("photos", "manage"),
      canEdit: this.$config.allow("photos", "update") && this.$config.feature("edit"),
      hasPlaces: this.$config.allow("places", "view") && this.$config.feature("places"),
      canSearchPlaces: this.$config.allow("places", "search") && this.$config.feature("places"),
      canAccessLibrary: this.$config.allow("photos", "access_library"),
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

      /**
       * Even if the filter is unchanged, if the route is changed (for example
       * from `/review` to `/browse`), then the lastFilter must be reset, so that
       * a new search is actually triggered. That is because both routes use
       * this component, so it is reused by vue. See
       * https://github.com/photoprism/photoprism/pull/2782#issuecomment-1279821448.
       *
       * However, if the route is unchanged, the not resetting lastFilter prevents
       * unnecessary search-api-calls! These search-calls would otherwise reset
       * the view, even if we for example just returned from a fullscreen-download
       * in the ios-pwa. See
       * https://github.com/photoprism/photoprism/pull/2782#issue-1409954466
       */
      const routeChanged = this.routeName !== this.$route.name;

      if (routeChanged) {
        this.resetLastFilter();
      }

      this.routeName = this.$route.name;

      if (this.uid !== this.$route.params.album) {
        this.uid = this.$route.params.album;
        this.findAlbum().then(() => this.search());
      } else {
        this.search();
      }
    },
  },
  created() {
    this.findAlbum().then(() => this.search());

    this.subscriptions.push(this.$event.subscribe("albums.updated", (ev, data) => this.onAlbumsUpdated(ev, data)));
    this.subscriptions.push(this.$event.subscribe("photos", (ev, data) => this.onUpdate(ev, data)));

    this.subscriptions.push(
      this.$event.subscribe("lightbox.opened", (ev, data) => {
        this.lightbox.open = true;
      })
    );
    this.subscriptions.push(
      this.$event.subscribe("lightbox.closed", (ev, data) => {
        this.lightbox.open = false;
      })
    );

    this.subscriptions.push(this.$event.subscribe("touchmove.top", () => this.refresh()));
    this.subscriptions.push(this.$event.subscribe("touchmove.bottom", () => this.loadMore()));
  },
  mounted() {
    this.$view.enter(this, this.$refs?.page);
  },
  beforeUnmount() {
    for (let i = 0; i < this.subscriptions.length; i++) {
      this.$event.unsubscribe(this.subscriptions[i]);
    }
  },
  unmounted() {
    this.$view.leave(this);
  },
  methods: {
    resetLastFilter() {
      this.lastFilter = {};
    },
    onCtrl(ev) {
      if (!ev || !(ev instanceof KeyboardEvent) || !ev.ctrlKey || !this.$view.isActive(this)) {
        return;
      }

      switch (ev.code) {
        case "KeyR":
          ev.preventDefault();
          this.refresh();
          break;
        case "KeyU":
          ev.preventDefault();
          if (this.$config.allow("files", "upload") && this.$config.feature("upload")) {
            this.$event.publish("dialog.upload");
          }
          break;
      }
    },
    hideExpansionPanel() {
      return this.$refs?.toolbar?.hideExpansionPanel();
    },
    getViewType() {
      let queryParam = this.$route.query["view"] ? this.$route.query["view"] : "";
      let defaultType = window.localStorage.getItem("photos.view");
      let storedType = window.localStorage.getItem("album.photos.view");

      if (queryParam) {
        window.localStorage.setItem("album.photos.view", queryParam);
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
    getSortOrder() {
      const query = this.$route.query;
      return query["order"] ? query["order"] : this.model?.Order;
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
    openLocation(index) {
      if (!this.hasPlaces) {
        return;
      }

      const photo = this.results[index];

      if (!photo) {
        return;
      }

      if (this.canAccessLibrary && photo.CellID && photo.CellID !== "zz") {
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
      } else if (this.getSortOrder() === "random") {
        this.$lightbox.openModels(Thumb.fromPhotos(this.results), index, this.model);
      } else {
        this.$lightbox.openView(this, index);
      }

      return true;
    },
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
        order: this.getSortOrder(),
      };

      Object.assign(params, this.lastFilter);

      if (this.staticFilter) {
        Object.assign(params, this.staticFilter);
      }

      this.lastParams = params;

      Photo.search(params)
        .then((response) => {
          this.results = Photo.mergeResponse(this.results, response);
          this.complete = response.count < count;
          this.scrollDisabled = this.complete;

          if (this.complete) {
            this.offset = offset;
            if (this.results.length > 1) {
              if (!this.lightbox.open) {
                this.$notify.info(
                  this.$gettextInterpolate(this.$gettext("%{n} pictures found"), { n: this.results.length })
                );
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
          this.scrollDisabled = false;
        })
        .finally(() => {
          this.dirty = false;
          this.loading = false;
          this.listen = true;
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

        window.localStorage.setItem("album.photos." + key, this.settings[key]);
      }
    },
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
    updateQuery(props) {
      this.updateFilter(props);

      if (this.loading) {
        return;
      }

      const query = {
        view: this.settings.view,
      };

      Object.assign(query, this.filter);

      for (let key in query) {
        if (query[key] === undefined || !query[key]) {
          delete query[key];
        }
      }

      if (JSON.stringify(this.$route.query) === JSON.stringify(query)) {
        return;
      }

      this.$router.replace({ query: query });
    },
    searchParams() {
      const params = {
        count: this.batchSize,
        offset: this.offset,
        s: this.uid,
        merged: true,
        order: this.getSortOrder(),
      };

      Object.assign(params, this.filter);

      if (this.staticFilter) {
        Object.assign(params, this.staticFilter);
      }

      return params;
    },
    refresh(props) {
      this.updateSettings(props);

      if (this.loading) {
        return;
      }

      this.loading = true;
      this.page = 0;
      this.dirty = true;
      this.complete = false;
      this.scrollDisabled = false;

      this.loadMore(true);
    },
    search() {
      /**
       * search is called on mount or route change. If the route changed to an
       * open lightbox, no search is required. There is no reason to do an
       * initial results load, if the results aren't currently visible
       */
      if (this.lightbox.open) {
        return;
      }

      this.scrollDisabled = true;

      // Don't query the same data more than once
      if (JSON.stringify(this.lastFilter) === JSON.stringify(this.filter)) {
        // this.$nextTick(() => this.$emit("scrollRefresh"));
        return;
      }

      Object.assign(this.lastFilter, this.filter);

      this.offset = 0;
      this.page = 0;
      this.loading = true;
      this.listen = false;
      this.complete = false;

      const params = this.searchParams();

      this.lastParams = params;

      Photo.search(params)
        .then((response) => {
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
              this.$notify.info(
                this.$gettextInterpolate(this.$gettext("%{n} pictures found"), { n: this.results.length })
              );
            }
          } else {
            // this.$notify.info(this.$gettextInterpolate(this.$gettext("More than %{n} pictures found"), {n: 100}));
            this.$nextTick(() => {
              if (this.$root.$el.clientHeight <= window.document.documentElement.clientHeight + 300) {
                this.loadMore();
              }
            });
          }
        })
        .finally(() => {
          this.dirty = false;
          this.loading = false;
          this.listen = true;
        });
    },
    findAlbum() {
      return this.model
        .find(this.uid)
        .then((m) => {
          this.model = m;

          window.document.title = `${this.$config.get("siteTitle")}: ${this.model.Title}`;

          return Promise.resolve(this.model);
        })
        .catch((e) => {
          this.$router.push({ name: this.collectionRoute });
          return Promise.reject(e);
        });
    },
    onAlbumsUpdated(ev, data) {
      if (!this.listen) {
        return;
      }

      if (!data || !data.entities || !Array.isArray(data.entities)) {
        return;
      }

      for (let i = 0; i < data.entities.length; i++) {
        if (this.model.UID === data.entities[i].UID) {
          let values = data.entities[i];

          for (let key in values) {
            if (values.hasOwnProperty(key)) {
              this.model[key] = values[key];
            }
          }

          window.document.title = `${this.$config.get("siteTitle")}: ${this.model.Title}`;

          this.dirty = true;
          this.complete = false;
          this.scrollDisabled = false;

          if (this.lastParams?.order !== this.model?.Order) {
            this.updateQuery();
            this.loadMore(true);
          } else {
            this.loadMore(true);
          }

          return;
        }
      }
    },
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
        case "updated":
          for (let i = 0; i < data.entities.length; i++) {
            this.updateResults(data.entities[i]);
          }
          break;
        case "restored":
          this.dirty = true;
          this.scrollDisabled = false;
          this.complete = false;

          this.loadMore();

          break;
        case "deleted":
        case "archived":
          this.dirty = true;
          this.complete = false;

          for (let i = 0; i < data.entities.length; i++) {
            const uid = data.entities[i];

            this.removeResult(this.results, uid);
            this.removeResult(this.lightbox.results, uid);
            this.$clipboard.removeId(uid);
          }

          break;
      }

      // TODO: Needed?
      this.$forceUpdate();
    },
  },
};
</script>
