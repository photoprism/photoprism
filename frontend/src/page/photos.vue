<template>
  <div v-infinite-scroll="loadMore" :class="$config.aclClasses('photos')" class="p-page p-page-photos" style="user-select: none"
       :infinite-scroll-disabled="scrollDisabled" :infinite-scroll-distance="scrollDistance"
       :infinite-scroll-listen-for-event="'scrollRefresh'">

    <p-photo-toolbar
      :context="context"
      :filter="filter"
      :static-filter="staticFilter"
      :settings="settings"
      :refresh="refresh"
      :update-filter="updateFilter"
      :update-query="updateQuery"
      :on-close="onClose"
      :embedded="embedded"
    />

    <v-container v-if="loading" fluid class="pa-4">
      <v-progress-linear color="secondary-dark" :indeterminate="true"></v-progress-linear>
    </v-container>
    <v-container v-else fluid class="pa-0">
      <p-scroll-top></p-scroll-top>

      <p-photo-clipboard :context="context" :refresh="refresh" :selection="selection"></p-photo-clipboard>

      <p-photo-mosaic v-if="settings.view === 'mosaic'"
                      :context="context"
                      :photos="results"
                      :select-mode="selectMode"
                      :filter="filter"
                      :edit-photo="editPhoto"
                      :open-photo="openPhoto"
                      :is-shared-view="isShared"></p-photo-mosaic>
      <p-photo-list v-else-if="settings.view === 'list'"
                    :context="context"
                    :photos="results"
                    :select-mode="selectMode"
                    :filter="filter"
                    :open-photo="openPhoto"
                    :edit-photo="editPhoto"
                    :open-location="openLocation"
                    :is-shared-view="isShared"></p-photo-list>
      <p-photo-cards v-else
                     :context="context"
                     :photos="results"
                     :select-mode="selectMode"
                     :filter="filter"
                     :open-photo="openPhoto"
                     :edit-photo="editPhoto"
                     :open-location="openLocation"
                     :is-shared-view="isShared"></p-photo-cards>
    </v-container>
  </div>
</template>

<script>
import {MediaAnimated, MediaLive, MediaRaw, MediaVideo, Photo} from "model/photo";
import Thumb from "model/thumb";
import Viewer from "common/viewer";
import Event from "pubsub-js";

export default {
  name: 'PPagePhotos',
  props: {
    staticFilter: {
      type: Object,
      default: () => {
      },
    },
    onClose: {
      type: Function,
      default: undefined,
    },
    embedded: {
      type: Boolean,
      default: false
    },
  },
  data() {
    const query = this.$route.query;
    const routeName = this.$route.name;
    const order = this.sortOrder();
    const camera = query['camera'] ? parseInt(query['camera']) : 0;
    const q = query['q'] ? query['q'] : '';
    const country = query['country'] ? query['country'] : '';
    const lens = query['lens'] ? parseInt(query['lens']) : 0;
    const year = query['year'] ? parseInt(query['year']) : 0;
    const month = query['month'] ? parseInt(query['month']) : 0;
    const color = query['color'] ? query['color'] : '';
    const label = query['label'] ? query['label'] : '';
    const latlng = query['latlng'] ? query['latlng'] : '';
    const view = this.viewType();
    const filter = {
      country: country,
      camera: camera,
      lens: lens,
      label: label,
      latlng: latlng,
      year: year,
      month: month,
      color: color,
      order: order,
      q: q,
    };

    const settings = this.$config.settings();
    const features = settings.features;

    if (settings) {
      if (features.private) {
        filter.public = "true";
      }

      if (features.review && (!this.staticFilter || !("quality" in this.staticFilter))) {
        filter.quality = "3";
      }
    }

    const batchSize = Photo.batchSize();

    return {
      isShared: this.$config.deny("photos", "manage"),
      canEdit: this.$config.allow("photos", "update") && features.edit,
      hasPlaces: this.$config.allow("places", "view") && features.places,
      canSearchPlaces: this.$config.allow("places", "search") && features.places,
      subscriptions: [],
      listen: false,
      dirty: false,
      complete: false,
      results: [],
      scrollDisabled: true,
      scrollDistance: window.innerHeight * 6,
      batchSize: batchSize,
      offset: 0,
      page: 0,
      selection: this.$clipboard.selection,
      settings: {
        view,
      },
      filter: filter,
      lastFilter: {},
      routeName: routeName,
      loading: true,
      viewer: {
        results: [],
        loading: false,
        complete: false,
        open: false,
        dirty: false,
        batchSize: batchSize * 4,
      },
    };
  },
  computed: {
    selectMode: function () {
      return this.selection.length > 0;
    },
    context: function () {
      if (!this.staticFilter) {
        return "photos";
      }

      if (this.staticFilter.review) {
        return "review";
      } else if (this.staticFilter.archived) {
        return "archive";
      } else if (this.staticFilter.favorite) {
        return "favorites";
      }

      return "";
    }
  },
  watch: {
    '$route'() {
      const query = this.$route.query;

      const settings = this.$config.settings();

      if (settings.features) {
        if (settings.features.private) {
          this.filter.public = "true";
        }

        if (settings.features.review && (!this.staticFilter || !("quality" in this.staticFilter))) {
          this.filter.quality = "3";
        }
      }

      this.filter.q = query['q'] ? query['q'] : '';
      this.filter.camera = query['camera'] ? parseInt(query['camera']) : 0;
      this.filter.country = query['country'] ? query['country'] : '';
      this.filter.lens = query['lens'] ? parseInt(query['lens']) : 0;
      this.filter.year = query['year'] ? parseInt(query['year']) : 0;
      this.filter.month = query['month'] ? parseInt(query['month']) : 0;
      this.filter.color = query['color'] ? query['color'] : '';
      this.filter.label = query['label'] ? query['label'] : '';
      this.filter.latlng = query['latlng'] ? query['latlng'] : '';
      this.filter.order = this.sortOrder();

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
      this.search();
    }
  },
  created() {
    this.search();

    this.subscriptions.push(Event.subscribe("import.completed", (ev, data) => this.onImportCompleted(ev, data)));
    this.subscriptions.push(Event.subscribe("photos", (ev, data) => this.onUpdate(ev, data)));

    this.subscriptions.push(Event.subscribe("viewer.show", (ev, data) => {
      this.viewer.open = true;
    }));
    this.subscriptions.push(Event.subscribe("viewer.hide", (ev, data) => {
      this.viewer.open = false;
    }));


    this.subscriptions.push(Event.subscribe("touchmove.top", () => this.refresh()));
    this.subscriptions.push(Event.subscribe("touchmove.bottom", () => this.loadMore()));
  },
  destroyed() {
    for (let i = 0; i < this.subscriptions.length; i++) {
      Event.unsubscribe(this.subscriptions[i]);
    }
  },
  methods: {
    searchCount() {
      const offset = parseInt(window.localStorage.getItem("photos_offset"));
      if(this.offset > 0 || !offset) {
        return this.batchSize;
      }
      return offset + this.batchSize;
    },
    setOffset(offset) {
      this.offset = offset;
      window.localStorage.setItem("photos_offset", offset);
    },
    viewType() {
      if (this.embedded) {
        return 'mosaic';
      }

      let queryParam = this.$route.query['view'] ? this.$route.query['view'] : '';
      let storedType = window.localStorage.getItem('photos_view');

      if (queryParam) {
        window.localStorage.setItem("photos_view", queryParam);
        return queryParam;
      } else if (storedType) {
        return storedType;
      } else if (window.innerWidth < 960) {
        return 'mosaic';
      }

      return 'cards';
    },
    sortOrder() {
      if (this.embedded) {
        return 'newest';
      }

      let queryParam = this.$route.query['order'];
      let storedType = window.localStorage.getItem('photos_order');

      if (queryParam) {
        window.localStorage.setItem('photos_order', queryParam);
        return queryParam;
      } else if (storedType) {
        return storedType;
      }

      return 'newest';
    },
    openLocation(index) {
      if (!this.hasPlaces || !this.canSearchPlaces) {
        return;
      }

      const photo = this.results[index];

      if (!photo) {
        return;
      }

      if (photo.CellID && photo.CellID !== "zz") {
        this.$router.push({name: "places", query: {q: photo.CellID}});
      } else if (photo.Country && photo.Country !== "zz") {
        this.$router.push({name: "places", query: {q: "country:" + photo.Country}});
      } else {
        this.$notify.warn("unknown location");
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
      Event.publish("dialog.edit", {selection: selection, album: null, index: index});
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
          this.$viewer.play({video: selected});
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
        merged: true,
      };

      Object.assign(params, this.lastFilter);

      if (this.staticFilter) {
        Object.assign(params, this.staticFilter);
      }

      Photo.search(params).then(response => {
        this.results = this.dirty ? response.models : Photo.mergeResponse(this.results, response);
        this.complete = (response.count < response.limit);
        this.scrollDisabled = this.complete;

        if (this.complete) {
          this.setOffset(response.offset);

          if (!this.embedded && this.results.length > 1) {
            this.$notify.info(this.$gettextInterpolate(this.$gettext("%{n} pictures found"), {n: this.results.length}));
          }
        } else if (this.results.length >= Photo.limit()) {
          this.setOffset(response.offset);
          this.complete = true;
          this.scrollDisabled = true;
          this.$notify.warn(this.$gettext("Can't load more, limit reached"));
        } else {
          this.setOffset(response.offset + response.limit);
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

        window.localStorage.setItem("photos_"+key, this.settings[key]);
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

      this.$router.replace({query});
    },
    searchParams() {
      const params = {
        count: this.searchCount(),
        offset: this.offset,
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
      /**
       * search is called on mount or route change. If the route changed to an
       * open viewer, no search is required. There is no reason to do an
       * initial results load, if the results aren't currently visible
       */
      if (this.viewer.open) {
        return;
      }

      /**
       * re-creating the last scroll-position should only ever happen when using
       * back-navigation. We therefore reset the remembered scroll-position
       * in any other scenario
       */
      if (!window.backwardsNavigationDetected) {
        this.setOffset(0);
      }
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
        this.offset = response.limit;
        this.results = response.models;
        this.viewer.results = [];
        this.viewer.complete = false;
        this.complete = (response.count < response.limit);
        this.scrollDisabled = this.complete;

        if (this.complete) {
          if (!this.results.length) {
            this.$notify.warn(this.$gettext("No pictures found"));
          } else if (!this.embedded && this.results.length === 1) {
            this.$notify.info(this.$gettext("One picture found"));
          } else if (!this.embedded) {
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
    onImportCompleted() {
      if (!this.listen) return;

      this.loadMore();
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

      if (!data || !data.entities || !Array.isArray(data.entities)) {
        return;
      }

      const type = ev.split('.')[1];

      switch (type) {
        case 'updated':
          for (let i = 0; i < data.entities.length; i++) {
            const values = data.entities[i];

            if (this.context === "review" && values.Quality >= 3) {
              this.removeResult(this.results, values.UID);
              this.removeResult(this.viewer.results, values.UID);
              this.$clipboard.removeId(values.UID);
            } else {
              this.updateResults(values);
            }
          }
          break;
        case 'restored':
          this.dirty = true;
          this.complete = false;

          if (this.context !== "archive") break;

          for (let i = 0; i < data.entities.length; i++) {
            const uid = data.entities[i];

            this.removeResult(this.results, uid);
            this.removeResult(this.viewer.results, uid);
          }

          break;
        case 'archived':
          this.dirty = true;
          this.complete = false;

          if (this.context !== "archive") {
            for (let i = 0; i < data.entities.length; i++) {
              const uid = data.entities[i];

              this.removeResult(this.results, uid);
              this.removeResult(this.viewer.results, uid);
              this.$clipboard.removeId(uid);
            }
          } else if (!this.results.length) {
            this.refresh();
          }

          break;
        case 'deleted':
          this.dirty = true;
          this.complete = false;

          for (let i = 0; i < data.entities.length; i++) {
            const uid = data.entities[i];

            this.removeResult(this.results, uid);
            this.removeResult(this.viewer.results, uid);
            this.$clipboard.removeId(uid);
          }

          break;
        case 'created':
          this.dirty = true;
          this.scrollDisabled = false;
          this.complete = false;

          break;
        default:
          console.warn("unexpected event type", ev);
      }

      // TODO: Needed?
      this.$forceUpdate();
    },
  },
};
</script>
