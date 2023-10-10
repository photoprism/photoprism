<template>
  <div v-infinite-scroll="loadMore" class="p-page p-page-album-photos" :infinite-scroll-disabled="scrollDisabled"
       :infinite-scroll-distance="scrollDistance" :infinite-scroll-listen-for-event="'scrollRefresh'">

    <p-album-toolbar :filter="filter" :album="model" :settings="settings" :refresh="refresh"
                     :update-filter="updateFilter" :update-query="updateQuery"></p-album-toolbar>

    <v-container v-if="loading" fluid class="pa-4">
      <v-progress-linear color="secondary-dark" :indeterminate="true"></v-progress-linear>
    </v-container>
    <v-container v-else fluid class="pa-0">
      <p-scroll-top></p-scroll-top>

      <p-photo-clipboard :refresh="refresh"
                         :selection="selection"
                         :album="model" context="album"></p-photo-clipboard>

      <p-photo-mosaic v-if="settings.view === 'mosaic'"
                      context="album"
                      :photos="results"
                      :select-mode="selectMode"
                      :filter="filter"
                      :album="model"
                      :edit-photo="editPhoto"
                      :open-photo="openPhoto"
                      :is-shared-view="isShared"></p-photo-mosaic>
      <p-photo-list v-else-if="settings.view === 'list'"
                    context="album"
                    :photos="results"
                    :select-mode="selectMode"
                    :filter="filter"
                    :album="model"
                    :open-photo="openPhoto"
                    :edit-photo="editPhoto"
                    :open-location="openLocation"
                    :is-shared-view="isShared"></p-photo-list>
      <p-photo-cards v-else
                     context="album"
                     :photos="results"
                     :select-mode="selectMode"
                     :filter="filter"
                     :album="model"
                     :open-photo="openPhoto"
                     :edit-photo="editPhoto"
                     :open-location="openLocation"
                     :is-shared-view="isShared"></p-photo-cards>
    </v-container>
  </div>
</template>

<script>
import {Photo, MediaLive, MediaRaw, MediaVideo, MediaAnimated} from "model/photo";
import Album from "model/album";
import Thumb from "model/thumb";
import Event from "pubsub-js";
import Viewer from "common/viewer";

export default {
  name: 'PPageAlbumPhotos',
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
    const order = query['order'] ? query['order'] : 'oldest';
    const camera = query['camera'] ? parseInt(query['camera']) : 0;
    const q = query['q'] ? query['q'] : '';
    const country = query['country'] ? query['country'] : '';
    const view = this.viewType();
    const filter = {country: country, camera: camera, order: order, q: q};
    const settings = {view: view};
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
      scrollDistance: window.innerHeight * 6,
      batchSize: batchSize,
      offset: 0,
      page: 0,
      selection: this.$clipboard.selection,
      settings: settings,
      filter: filter,
      lastFilter: {},
      routeName: routeName,
      loading: true,
      viewer: {
        results: [],
        loading: false,
        complete: false,
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
    '$route'() {
      const query = this.$route.query;

      this.filter.q = query['q'] ? query['q'] : '';
      this.filter.camera = query['camera'] ? parseInt(query['camera']) : 0;
      this.filter.country = query['country'] ? query['country'] : '';
      this.settings.view = this.viewType();

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
        this.lastFilter = {};
      }

      this.routeName = this.$route.name;

      if (this.uid !== this.$route.params.album) {
        this.uid = this.$route.params.album;
        this.findAlbum().then(() => this.search());
      } else {
        this.search();
      }
    }
  },
  created() {
    this.findAlbum().then(() => this.search());

    this.subscriptions.push(Event.subscribe("albums.updated", (ev, data) => this.onAlbumsUpdated(ev, data)));
    this.subscriptions.push(Event.subscribe("photos", (ev, data) => this.onUpdate(ev, data)));

    this.subscriptions.push(Event.subscribe("touchmove.top", () => this.refresh()));
    this.subscriptions.push(Event.subscribe("touchmove.bottom", () => this.loadMore()));
  },
  destroyed() {
    for (let i = 0; i < this.subscriptions.length; i++) {
      Event.unsubscribe(this.subscriptions[i]);
    }
  },
  methods: {
    viewType() {
      let queryParam = this.$route.query['view'] ? this.$route.query['view'] : "";
      let defaultType = window.localStorage.getItem("photos_view");
      let storedType = window.localStorage.getItem("album_photos_view");

      if (queryParam) {
        window.localStorage.setItem("album_photos_view", queryParam);
        return queryParam;
      } else if (storedType) {
        return storedType;
      } else if (defaultType) {
        return defaultType;
      } else if (window.innerWidth < 960) {
        return 'mosaic';
      }

      return 'cards';
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
        this.$router.push({name: "places", query: {q: photo.CellID}});
      } else if (this.uid) {
        this.$router.push({name: "places_view", params: {s: this.uid}, query: {q: photo.CellID}});
      }
    },
    editPhoto(index) {
      if (!this.canEdit) {
        return this.openPhoto(index);
      }

      let selection = this.results.map((p) => {
        return p.getId();
      });

      // Open Edit Dialog
      Event.publish("dialog.edit", {selection: selection, album: this.album, index: index});
    },
    openPhoto(index, showMerged = false, preferVideo = false) {
      if (this.loading || !this.listen || this.viewer.loading || !this.results[index]) {
        return false;
      }

      const selected = this.results[index];

      // Don't open as stack when user is selecting pictures, or a RAW has only one JPEG.
      if (this.selection.length > 0 || selected.Type === MediaRaw && selected.jpegFiles().length < 2) {
        showMerged = false;
      }

      /**
       * If the file is a video or an animation (like gif), then we always play
       * it in the video-player.
       * If the file is a live-image (an image with an embedded video), then we only
       * play it in the video-player if specifically requested.
       * This is because:
       * 1. the lower-resolution video in these files is already
       *    played when hovering the element (which does not happen for regular
       *    video files)
       * 2. The video in live-images is an addon. The main focus is usually still
       *    the high resolution image inside
       *
       * preferVideo is true, when the user explicitly clicks the live-image-icon.
       */
      if (preferVideo && selected.Type === MediaLive || selected.Type === MediaVideo || selected.Type === MediaAnimated) {
        if (selected.isPlayable()) {
          this.$viewer.play({video: selected, album: this.album});
        } else {
          this.$viewer.show(Thumb.fromPhotos(this.results), index);
        }
      } else if (showMerged) {
        this.$viewer.show(Thumb.fromFiles([selected]), 0);
      } else {
        Viewer.show(this, index);
      }

      return true;
    },
    loadMore() {
      if (this.scrollDisabled || this.$scrollbar.disabled()) return;

      this.scrollDisabled = true;
      this.listen = false;

      if (this.dirty) {
        this.viewer.dirty = true;
      }

      const count = this.dirty ? (this.page + 2) * this.batchSize : this.batchSize;
      const offset = this.dirty ? 0 : this.offset;

      const params = {
        count: count,
        offset: offset,
        s: this.uid,
        merged: true,
      };

      Object.assign(params, this.lastFilter);

      if (this.staticFilter) {
        Object.assign(params, this.staticFilter);
      }

      Photo.search(params).then(response => {
        this.results = Photo.mergeResponse(this.results, response);
        this.complete = (response.count < count);
        this.scrollDisabled = this.complete;

        if (this.complete) {
          this.offset = offset;

          if (this.results.length > 1) {
            this.$notify.info(this.$gettextInterpolate(this.$gettext("%{n} pictures found"), {n: this.results.length}));
          }
        } else if (this.results.length >= Photo.limit()) {
          this.offset = offset;
          this.scrollDisabled = true;
          this.complete = true;
          this.$notify.warn(this.$gettext("Can't load more, limit reached"));
        } else {
          this.offset = offset + count;
          this.page++;

          this.$nextTick(() => {
            if (this.$root.$el.clientHeight <= window.document.documentElement.clientHeight + 300) {
              this.$emit("scrollRefresh");
            }
          });
        }
      }).catch(() => {
        this.scrollDisabled = false;
      }).finally(() => {
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

        window.localStorage.setItem("album_photos_"+key, this.settings[key]);
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

      if (this.model.Order !== this.filter.order) {
        this.model.Order = this.filter.order;
        this.updateAlbum();
      }

      if (this.loading) return;

      const query = {
        view: this.settings.view
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

      this.$router.replace({query: query});
    },
    searchParams() {
      const params = {
        count: this.batchSize,
        offset: this.offset,
        s: this.uid,
        merged: true,
      };

      Object.assign(params, this.filter);

      if (this.staticFilter) {
        Object.assign(params, this.staticFilter);
      }

      return params;
    },
    refresh(props) {
      this.updateSettings(props);

      if (this.loading) return;

      this.loading = true;
      this.page = 0;
      this.dirty = true;
      this.complete = false;
      this.scrollDisabled = false;

      this.loadMore();
    },
    search() {
      this.scrollDisabled = true;

      // Don't query the same data more than once
      if (JSON.stringify(this.lastFilter) === JSON.stringify(this.filter)) {
        this.$nextTick(() => this.$emit("scrollRefresh"));
        return;
      }

      Object.assign(this.lastFilter, this.filter);

      this.offset = 0;
      this.page = 0;
      this.loading = true;
      this.listen = false;
      this.complete = false;

      const params = this.searchParams();

      Photo.search(params).then(response => {
        this.offset = this.batchSize;
        this.results = response.models;
        this.viewer.results = [];
        this.viewer.complete = false;
        this.complete = (response.count < this.batchSize);
        this.scrollDisabled = this.complete;

        if (this.complete) {
          if (!this.results.length) {
            this.$notify.warn(this.$gettext("No pictures found"));
          } else if (this.results.length === 1) {
            this.$notify.info(this.$gettext("One picture found"));
          } else {
            this.$notify.info(this.$gettextInterpolate(this.$gettext("%{n} pictures found"), {n: this.results.length}));
          }
        } else {
          // this.$notify.info(this.$gettextInterpolate(this.$gettext("More than %{n} pictures found"), {n: 100}));
          this.$nextTick(() => {
            if (this.$root.$el.clientHeight <= window.document.documentElement.clientHeight + 300) {
              this.$emit("scrollRefresh");
            }
          });
        }
      }).finally(() => {
        this.dirty = false;
        this.loading = false;
        this.listen = true;
      });
    },
    findAlbum() {
      return this.model.find(this.uid).then(m => {
        this.model = m;

        this.filter.order = m.Order;
        window.document.title = `${this.$config.get("siteTitle")}: ${this.model.Title}`;

        return Promise.resolve(this.model);
      }).catch((e) => {
        this.$router.push({ name: "albums" });
        return Promise.reject(e);
      });
    },
    onAlbumsUpdated(ev, data) {
      if (!this.listen) return;

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

          if (this.filter.order !== this.model.Order) {
            this.filter.order = this.model.Order;
            this.updateQuery();
          } else {
            this.loadMore();
          }

          return;
        }
      }
    },
    updateResults(entity) {
      this.results.filter((m) => m.UID === entity.UID).forEach((m) => {
        for (let key in entity) {
          if (key !== "UID" && entity.hasOwnProperty(key) && entity[key] != null && typeof entity[key] !== "object") {
            m[key] = entity[key];
          }
        }
      });

      this.viewer.results.filter((m) => m.UID === entity.UID).forEach((m) => {
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
      if (!this.listen) return;

      if (!data || !data.entities) {
        return;
      }

      const type = ev.split('.')[1];

      switch (type) {
        case 'updated':
          for (let i = 0; i < data.entities.length; i++) {
            this.updateResults(data.entities[i]);
          }
          break;
        case 'restored':
          this.dirty = true;
          this.scrollDisabled = false;
          this.complete = false;

          this.loadMore();

          break;
        case 'deleted':
        case 'archived':
          this.dirty = true;
          this.complete = false;

          for (let i = 0; i < data.entities.length; i++) {
            const uid = data.entities[i];

            this.removeResult(this.results, uid);
            this.removeResult(this.viewer.results, uid);
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
