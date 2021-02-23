<template>
  <div class="p-page p-page-instant">
    <p-photo-toolbar
      :settings="settings"
      :filter="filter"
      :filter-change="updateQuery"
      :dirty="dirty"
      :refresh="refresh"
    ></p-photo-toolbar>

    <v-container v-if="loading" fluid class="pa-4">
      <v-progress-linear
        color="secondary-dark"
        :indeterminate="true"
      ></v-progress-linear>
    </v-container>
    <v-container v-else fluid class="pa-0">
      <div
        v-for="i in chunkCount"
        :id="'photo-chunk-' + (i - 1)"
        :key="i"
      ></div>
    </v-container>
  </div>
</template>

<script>
import Photo from "model/photo";
import Thumb from "model/thumb";

const PIG_THUMBNAIL_SIZE = 50;
const PIG_IMAGE_SIZE = 224;

export default {
  name: "PPageInstant",
  props: {
    staticFilter: Object,
  },
  data() {
    const query = this.$route.query;
    const routeName = this.$route.name;
    const order = this.sortOrder();
    const camera = query["camera"] ? parseInt(query["camera"]) : 0;
    const q = query["q"] ? query["q"] : "";
    const country = query["country"] ? query["country"] : "";
    const lens = query["lens"] ? parseInt(query["lens"]) : 0;
    const year = query["year"] ? parseInt(query["year"]) : 0;
    const month = query["month"] ? parseInt(query["month"]) : 0;
    const color = query["color"] ? query["color"] : "";
    const label = query["label"] ? query["label"] : "";
    const filter = {
      country: country,
      camera: camera,
      lens: lens,
      label: label,
      year: year,
      month: month,
      color: color,
      order: order,
      q: q,
    };

    const settings = this.$config.settings();

    if (settings && settings.features.private) {
      filter.public = true;
    }

    if (
      settings &&
      settings.features.review &&
      (!this.staticFilter || !("quality" in this.staticFilter))
    ) {
      filter.quality = 3;
    }

    return {
      listen: false,
      dirty: false,
      complete: false,
      results: [],
      scrollDisabled: true,
      batchSize: Photo.batchSize(),
      offset: 0,
      page: 0,
      selection: this.$clipboard.selection,
      settings: {},
      filter: filter,
      lastFilter: {},
      routeName: routeName,
      loading: true,
      viewer: {
        results: [],
        loading: false,
      },
    };
  },
  computed: {
    chunkCount: function () {
      return Math.ceil(this.results.length / this.batchSize);
    },
  },
  watch: {
    $route() {
      const query = this.$route.query;

      this.filter.q = query["q"] ? query["q"] : "";
      this.filter.camera = query["camera"] ? parseInt(query["camera"]) : 0;
      this.filter.country = query["country"] ? query["country"] : "";
      this.filter.lens = query["lens"] ? parseInt(query["lens"]) : 0;
      this.filter.year = query["year"] ? parseInt(query["year"]) : 0;
      this.filter.month = query["month"] ? parseInt(query["month"]) : 0;
      this.filter.color = query["color"] ? query["color"] : "";
      this.filter.label = query["label"] ? query["label"] : "";
      this.filter.order = this.sortOrder();
      this.lastFilter = {};
      this.routeName = this.$route.name;
      this.refresh();
    },
    results() {
      this.$forceUpdate();
    },
  },
  created() {
    this.search();
  },
  mounted() {
    let externalScript = document.createElement("script");
    externalScript.setAttribute("src", "/static/scripts/pig.min.js?1123");
    document.head.appendChild(externalScript);
  },
  methods: {
    sortOrder() {
      let queryParam = this.$route.query["order"];
      let storedType = window.localStorage.getItem("photo_order");

      if (queryParam) {
        window.localStorage.setItem("photo_order", queryParam);
        return queryParam;
      } else if (storedType) {
        return storedType;
      }

      return "newest";
    },
    openPhoto(photo) {
      this.viewerResults().then((results) => {
        const thumbsIndex = results.findIndex(
          (result) => result.UID === photo.UID
        );

        if (thumbsIndex >= 0) {
          this.$viewer.show(Thumb.fromPhotos(results), thumbsIndex);
        } else {
          this.$notify.warn(this.$gettext("Can't find the image"));
        }
      });
    },
    viewerResults() {
      if (this.complete || this.loading || this.viewer.loading) {
        return Promise.resolve(this.results);
      }

      if (this.viewer.results.length > this.results.length + this.batchSize) {
        return Promise.resolve(this.viewer.results);
      }

      this.viewer.loading = true;

      const count = this.batchSize * (this.page + 6);
      const offset = 0;

      const params = {
        count: count,
        offset: offset,
        merged: true,
      };

      Object.assign(params, this.lastFilter);

      if (this.staticFilter) {
        Object.assign(params, this.staticFilter);
      }

      return Photo.search(params).then(
        (resp) => {
          // Success.
          this.viewer.loading = false;
          this.viewer.results = resp.models;
          return Promise.resolve(this.viewer.results);
        },
        () => {
          // Error.
          this.viewer.loading = false;
          return Promise.resolve(this.results);
        }
      );
    },
    updateQuery() {
      this.filter.q = this.filter.q.trim();

      const query = {};

      Object.assign(query, this.filter);

      for (let key in query) {
        if (query[key] === undefined || !query[key]) {
          delete query[key];
        }
      }

      if (JSON.stringify(this.$route.query) === JSON.stringify(query)) {
        return;
      }

      this.$router.replace({ query });
    },
    searchParams() {
      const params = {
        count: this.batchSize,
        offset: this.offset,
        merged: true,
      };

      Object.assign(params, this.filter);

      if (this.staticFilter) {
        Object.assign(params, this.staticFilter);
      }

      return params;
    },
    refresh() {
      if (this.loading) {
        return;
      }

      this.loading = true;
      this.dirty = true;
      this.results = [];

      this.search();
    },
    search() {
      // Don't query the same data more than once
      if (
        JSON.stringify(this.lastFilter) === JSON.stringify(this.filter) &&
        !this.dirty
      ) {
        this.$nextTick(() => this.$emit("scrollRefresh"));
        return;
      }

      Object.assign(this.lastFilter, this.filter);

      this.offset = 0;
      this.page = 0;
      this.listen = false;
      this.complete = false;

      const params = this.searchParams();

      this.fetchRecursively(params).then(() => {
        this.dirty = false;
        this.loading = false;
        this.listen = true;

        this.viewerResults();
      });
    },
    fetchRecursively(params) {
      return Photo.search(params).then((response) => {
        this.complete = response.count < this.batchSize;
        this.results = Photo.mergeResponse(this.results, response);
        this.offset += response.count;

        const observer = new MutationObserver(() => {
          if (!document.getElementById(`photo-chunk-${this.page}`)) {
            return;
          }

          // Add a new chunk
          const imageData = response.models.map((result) => ({
            filename: result,
            aspectRatio: 1,
          }));

          const pig = new Pig(imageData, {
            containerId: `photo-chunk-${this.page}`,
            classPrefix: "pig-item",
            thumbnailSize: PIG_THUMBNAIL_SIZE,
            getImageSize: () => PIG_IMAGE_SIZE,
            urlForSize: (photo, size) => photo.thumbnailUrl(`tile_${size}`),
            onClickHandler: this.openPhoto,
          });
          pig.enable();

          this.page++;

          observer.disconnect();
        });

        observer.observe(document, {
          attributes: true,
          childList: true,
          subtree: true,
        });

        if (this.complete) {
          return;
        } else {
          params.offset += this.batchSize;
          return this.fetchRecursively(params);
        }
      });
    },
  },
};
</script>
