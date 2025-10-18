<template>
  <div
    ref="page"
    tabindex="1"
    class="p-page p-page-places fill-height"
    :class="$config.aclClasses('places')"
    @keydown="onKeyDown"
  >
    <div class="places" :class="'places--' + projection">
      <div v-if="mapError">
        <v-toolbar
          flat
          :density="$vuetify.display.smAndDown ? 'compact' : 'default'"
          class="page-toolbar"
          color="secondary"
        >
          <v-toolbar-title>
            {{ $gettext("Places") }}
          </v-toolbar-title>
        </v-toolbar>
        <div class="pa-3">
          <v-alert color="primary" icon="mdi-alert-circle-outline" class="v-alert--default" variant="outlined">
            <div class="font-weight-bold">
              {{ mapError }}
            </div>
          </v-alert>
        </div>
      </div>
      <div v-else-if="canSearch" class="map-control search-control">
        <div ref="search" class="maplibregl-ctrl maplibregl-ctrl-group map-control-search">
          <v-text-field
            v-model.lazy.trim="filter.q"
            :placeholder="$gettext('Search')"
            tabindex="1"
            density="compact"
            flat
            single-line
            overflow
            clearable
            hide-details
            theme="light"
            validate-on="invalid-input"
            prepend-inner-icon="mdi-magnify"
            autocomplete="off"
            autocorrect="off"
            autocapitalize="none"
            class="input-search pa-0"
            @click:clear="clearQuery"
            @keyup.enter="formChange"
          ></v-text-field>
        </div>
      </div>
      <div ref="background" class="map-background"></div>
      <div ref="map" class="map-container" :class="{ 'map-loaded': loaded }"></div>
      <div v-if="showCluster" class="cluster-control">
        <v-card class="cluster-control-container">
          <p-page-photos ref="cluster" :static-filter="cluster" :on-close="closeCluster" :embedded="true" />
        </v-card>
      </div>
    </div>
  </div>
</template>

<script>
import $api from "common/api";
import $fullscreen from "common/fullscreen";
import * as sky from "common/sky";
import * as map from "common/map";
import * as options from "options/options";
import Thumb from "model/thumb";
import PPagePhotos from "page/photos.vue";
import MapStyleControl from "component/places/style-control";

const ProjectionGlobe = "globe";
const ProjectionMercator = "mercator";
const ProjectionVertical = "vertical-perspective";

// Pixels the map pans when the up or down arrow is clicked:
const deltaDistance = 100;

// Degrees the map rotates when the left or right arrow is clicked:
const deltaDegrees = 25;

// Easing callback function.
const easing = (t) => {
  return t * (2 - t);
};

// MapLibre GL.
let maplibregl;

export default {
  name: "PPagePlaces",
  components: {
    PPagePhotos,
  },
  props: {
    staticFilter: {
      type: Object,
      default: () => {},
    },
  },
  data() {
    const filter = {
      q: this.query(),
      s: this.scope(),
    };

    const settings = this.$config.getSettings();
    const features = settings.features;

    if (features) {
      if (features.private) {
        filter.public = "true";
      }

      if (features.review && (!this.staticFilter || !("quality" in this.staticFilter))) {
        filter.quality = "3";
      }
    }

    return {
      isRtl: this.$config.isRtl(),
      canSearch: this.$config.allow("places", "search"),
      canUpload: this.$config.allow("files", "upload") && features.upload,
      featExperimental: this.$config.featExperimental(),
      initialized: false,
      map: null,
      mapError: false,
      markers: {},
      markersOnScreen: {},
      markerHandlers: {},
      clusterIds: [],
      loading: false,
      style: "",
      projection: "",
      mapStyles: [],
      terrain: {
        "topo-v2": "terrain_rgb",
        "outdoor-v2": "terrain-rgb",
        "0195eda5-6f09-7acd-8520-ab103fc75810": "terrain-rgb-v2",
        "414c531c-926d-4164-a057-455a215c0eee": "terrain_rgb_virtual",
      },
      attribution:
        '<a href="https://www.maptiler.com/copyright/" target="_blank">&copy; MapTiler</a> <a href="https://www.openstreetmap.org/copyright" target="_blank">&copy; OpenStreetMap contributors</a>',
      maxCount: 500000,
      options: {},
      mapFont: ["Open Sans Regular"],
      result: {},
      filter: filter,
      lastFilter: {},
      cluster: {},
      showCluster: false,
      config: this.$config.values,
      settings: settings.maps,
      animate: settings.maps.animate,
      loaded: false,
      skyRendered: false,
    };
  },
  watch: {
    $route() {
      if (!this.$view.isActive(this)) {
        return;
      }

      this.$view.focus(this.$refs?.page);

      this.filter.q = this.query();
      this.filter.s = this.scope();
      this.initialized = false;

      this.search();
    },
  },
  created() {
    this._mapDataHandler = null;
    if (this.$config.has("mapKey")) {
      this.mapStyles = options.MapsStyle(this.featExperimental);
    } else {
      this.mapStyles = options.MapsStyle(this.featExperimental).filter((s) => !s.Sponsor);
    }
  },
  mounted() {
    this.$view.enter(this);

    map.load().then((m) => {
      maplibregl = m;
      this.initMap()
        .then(() => {
          this.renderMap();
          this.openClusterFromUrl();
        })
        .catch((err) => {
          this.mapError = err;
        });
    });
  },
  beforeUnmount() {
    // Exit fullscreen mode if enabled, has no effect otherwise.
    $fullscreen.exit();
    this.teardownMap();
  },
  unmounted() {
    this.$view.leave(this);
  },
  methods: {
    // Removes map event listeners and marker handlers before the component is disposed.
    teardownMap() {
      if (!this.map) {
        return;
      }

      if (typeof this.map.off === "function") {
        if (this._mapDataHandler) {
          this.map.off("data", this._mapDataHandler);
        }
        this.map.off("move", this.updateMarkers);
        this.map.off("moveend", this.updateMarkers);
        this.map.off("resize", this.updateMarkers);
        this.map.off("idle", this.updateMarkers);
      }

      this._mapDataHandler = null;

      this.clearAllMarkerClicks();

      Object.keys(this.markersOnScreen).forEach((markerId) => {
        const marker = this.markersOnScreen[markerId];
        if (marker && typeof marker.remove === "function") {
          marker.remove();
        }
      });

      this.markers = {};
      this.markersOnScreen = {};
      this.markerHandlers = {};

      if (typeof this.map.remove === "function") {
        this.map.remove();
      }

      this.map = null;
    },
    ensureMarkerClick(id, marker, key, handlerFactory) {
      const markerId = String(id);
      if (!marker || typeof marker.getElement !== "function") {
        return;
      }

      const element = marker.getElement();

      if (!element || typeof element.addEventListener !== "function") {
        return;
      }

      const existing = this.markerHandlers[markerId];

      if (existing && existing.key === key && typeof existing.handler === "function") {
        return;
      }

      if (existing && typeof existing.handler === "function" && typeof element.removeEventListener === "function") {
        element.removeEventListener("click", existing.handler);
      }

      const handler = handlerFactory();

      if (typeof handler !== "function") {
        delete this.markerHandlers[markerId];
        return;
      }

      element.addEventListener("click", handler);
      this.markerHandlers[markerId] = { handler, key };
    },
    clearMarkerClick(id, marker) {
      const markerId = String(id);
      const existing = this.markerHandlers[markerId];

      if (!existing) {
        return;
      }

      const element = marker && typeof marker.getElement === "function" ? marker.getElement() : null;

      if (element && typeof element.removeEventListener === "function" && typeof existing.handler === "function") {
        element.removeEventListener("click", existing.handler);
      }

      delete this.markerHandlers[markerId];
    },
    clearAllMarkerClicks() {
      Object.keys(this.markerHandlers).forEach((markerId) => {
        const marker = this.markers[markerId] || this.markersOnScreen[markerId];
        this.clearMarkerClick(markerId, marker);
      });
    },
    renderSky() {
      if (!this.skyRendered && sky.render && this.$refs.background) {
        this.$nextTick(() => {
          sky.render(this.$refs.background, 320);
          this.skyRendered = true;
        });
      }
    },
    onKeyDown(ev) {
      if (
        !ev ||
        !(ev instanceof KeyboardEvent) ||
        !this.$view.isActive(this) ||
        document.activeElement instanceof HTMLInputElement
      ) {
        return;
      }

      if (ev.ctrlKey) {
        switch (ev.code) {
          case "KeyR":
            ev.preventDefault();
            this.reload();
            break;
          case "KeyG":
            ev.preventDefault();
            this.toggleProjection();
            break;
          case "KeyF":
            ev.preventDefault();
            this.$view.focus(this.$refs?.search, ".input-search input", false);
            break;
          case "KeyU":
            ev.preventDefault();
            if (this.canUpload) {
              this.$event.publish("dialog.upload");
            }
            break;
        }
      } else if (this.initialized) {
        // Use the arrow keys to move around the map with game-like controls.
        switch (ev.code) {
          case "ArrowUp":
            ev.preventDefault();
            this.map.panBy([0, -deltaDistance], {
              easing,
            });
            break;
          case "ArrowDown":
            ev.preventDefault();
            this.map.panBy([0, deltaDistance], {
              easing,
            });
            break;
          case "ArrowRight":
            ev.preventDefault();
            this.map.easeTo({
              bearing: this.map.getBearing() + deltaDegrees,
              easing,
            });
            break;
          case "ArrowLeft":
            ev.preventDefault();
            this.map.easeTo({
              bearing: this.map.getBearing() - deltaDegrees,
              easing,
            });
            break;
        }
      }
    },
    toggleProjection() {
      if (!this.initialized || this.loading) {
        return;
      }

      const currentProjection = this.getProjection();

      if (currentProjection === ProjectionMercator || !currentProjection) {
        this.setProjection(ProjectionGlobe);
      } else {
        this.setProjection(ProjectionMercator);
      }
    },
    getProjection(fromStorage) {
      if (fromStorage || !this.map || typeof this.map.getProjection !== "function") {
        const lastProjection = localStorage.getItem("places.projection");
        return lastProjection ? lastProjection : "";
      }

      return this.map.getProjection()?.type;
    },
    setProjection(newProjection) {
      const currentProjection = this.getProjection();

      if (currentProjection === newProjection) {
        return;
      }

      switch (newProjection) {
        case ProjectionGlobe:
          this.map.setZoom(3);
          break;
        case ProjectionMercator:
          break;
        case ProjectionVertical:
          break;
      }

      this.map.setProjection({ type: newProjection });
      this.projection = newProjection;

      if (!(this.$refs?.map instanceof HTMLElement)) {
        return;
      }

      const btn = this.$refs.map.querySelector(".maplibregl-ctrl-globe, .maplibregl-ctrl-globe-enabled");

      if (btn && btn instanceof HTMLElement) {
        switch (newProjection) {
          case ProjectionGlobe:
            btn.classList.add("maplibregl-ctrl-globe-enabled");
            btn.classList.remove("maplibregl-ctrl-globe");
            btn.classList.title = this.map._getUIString("GlobeControl.Disable");
            break;
          default:
            btn.classList.add("maplibregl-ctrl-globe");
            btn.classList.remove("maplibregl-ctrl-globe-enabled");
            btn.classList.title = this.map._getUIString("GlobeControl.Enable");
            break;
        }
      }
    },
    onProjectionChange(ev) {
      // Update current projection.
      this.projection = ev.newProjection;

      // Remember last used projection.
      localStorage.setItem("places.projection", ev.newProjection);

      // Render sky if new project is globe.
      if (ev.newProjection === ProjectionGlobe) {
        this.renderSky();
      }
    },
    noWebGlSupport() {
      // see https://maplibre.org/maplibre-gl-js/docs/examples/check-for-support/
      if (window.WebGLRenderingContext) {
        const canvas = document.createElement("canvas");
        try {
          // Note that { failIfMajorPerformanceCaveat: true } can be passed as a second argument
          // to canvas.getContext(), causing the check to fail if hardware rendering is not available. See
          // https://developer.mozilla.org/en-US/docs/Web/API/HTMLCanvasElement/getContext
          // for more details.
          const context = canvas.getContext("webgl2") || canvas.getContext("webgl");
          if (context && typeof context.getParameter == "function") {
            return false;
          }
        } catch (e) {
          // WebGL is supported, but disabled.
        }
        return this.$gettext("WebGL support is disabled in your browser");
      }

      // WebGL is not supported.
      return this.$gettext("Your browser does not support WebGL");
    },
    initMap() {
      return this.$config.load().finally(() => {
        const err = this.noWebGlSupport();
        if (err) {
          return Promise.reject(err);
        }
        this.configureMap(this.$config.values.settings.maps.style);
        return Promise.resolve();
      });
    },
    setStyle(style) {
      if (this.loading) {
        return false;
      }

      this.$notify.blockUI("busy");

      this.lastFilter = {};
      this.initialized = false;
      this.$refs.map.innerHTML = "";

      this.configureMap(style);
      this.renderMap();

      this.$notify.unblockUI();

      return true;
    },
    configureMap(style) {
      const filter = {
        q: this.query(),
        s: this.scope(),
      };

      let mapKey = "";

      if (this.$config.has("mapKey")) {
        // Remove non-alphanumeric characters from key.
        mapKey = this.$config.get("mapKey").replace(/[^a-z0-9]/gi, "");
      }

      const settings = this.$config.getSettings();
      const features = settings.features;

      if (settings) {
        if (features.private) {
          filter.public = "true";
        }

        if (features.review && (!this.staticFilter || !("quality" in this.staticFilter))) {
          filter.quality = "3";
        }
      }

      switch (style) {
        case "offline":
          this.style = this.featExperimental ? "low-resolution" : "default";
          break;
        case "streets":
          this.style = "streets-v2";
          break;
        case "hybrid":
          this.style = "414c531c-926d-4164-a057-455a215c0eee";
          break;
        case "satellite":
          this.style = "0195eda5-6f09-7acd-8520-ab103fc75810";
          break;
        case "outdoor":
          this.style = "outdoor-v2";
          break;
        case "topographique":
          this.style = "topo-v2";
          break;
        case "":
        case "basic":
        case "standard":
        case "buildings":
          this.style = "default";
          break;
        default:
          this.style = style;
      }

      if (!mapKey && this.style !== "low-resolution") {
        this.style = "default";
      }

      let mapOptions = {
        container: this.$refs.map,
        style: `https://api.maptiler.com/maps/${this.style}/style.json?key=${mapKey}`,
        glyphs: `https://api.maptiler.com/fonts/{fontstack}/{range}.pbf?key=${mapKey}`,
        attributionControl: { compact: true },
        zoom: 0,
      };

      if (this.style === "default") {
        mapOptions = {
          container: this.$refs.map,
          // Styles can be edited/created with https://maplibre.org/maputnik/.
          // To test new styles, put the style file in /assets/static/maps
          // and include it from there e.g. "/static/maps/default.json":
          // style: `/static/maps/${this.style}.json`,
          style: `https://cdn.photoprism.app/maps/${this.style}.json`,
          glyphs: `https://cdn.photoprism.app/maps/font/{fontstack}/{range}.pbf`,
          zoom: 0,
        };
      } else if (this.style === "low-resolution") {
        mapOptions = {
          container: this.$refs.map,
          style: {
            version: 8,
            sources: {
              world: {
                type: "geojson",
                data: `${this.$config.staticUri}/geo/world.json`,
                maxzoom: 6,
              },
            },
            glyphs: `${this.$config.staticUri}/font/{fontstack}/{range}.pbf`,
            layers: [
              {
                id: "background",
                type: "background",
                paint: {
                  "background-color": "#aadafe",
                },
              },
              {
                id: "land",
                type: "fill",
                source: "world",
                // "source-layer": "land",
                paint: {
                  "fill-color": "#cbe5ca",
                },
              },
              {
                id: "country-abbrev",
                type: "symbol",
                source: "world",
                maxzoom: 3,
                layout: {
                  "text-field": "{abbrev}",
                  "text-font": ["Open Sans Semibold"],
                  "text-transform": "uppercase",
                  "text-max-width": 20,
                  "text-size": {
                    stops: [
                      [3, 10],
                      [4, 11],
                      [5, 12],
                      [6, 16],
                    ],
                  },
                  "text-letter-spacing": {
                    stops: [
                      [4, 0],
                      [5, 1],
                      [6, 2],
                    ],
                  },
                  "text-line-height": {
                    stops: [
                      [5, 1.2],
                      [6, 2],
                    ],
                  },
                },
                paint: {
                  "text-halo-color": "#fff",
                  "text-halo-width": 1,
                },
              },
              {
                id: "country-border",
                type: "line",
                source: "world",
                paint: {
                  "line-color": "#226688",
                  "line-opacity": 0.25,
                  "line-dasharray": [6, 2, 2, 2],
                  "line-width": 1.2,
                },
              },
              {
                id: "country-name",
                type: "symbol",
                minzoom: 3,
                source: "world",
                layout: {
                  "text-field": "{name}",
                  "text-font": ["Open Sans Semibold"],
                  "text-max-width": 20,
                  "text-size": {
                    stops: [
                      [3, 10],
                      [4, 11],
                      [5, 12],
                      [6, 16],
                    ],
                  },
                },
                paint: {
                  "text-halo-color": "#fff",
                  "text-halo-width": 1,
                },
              },
            ],
          },
          attributionControl: false,
          zoom: 0,
        };
      }

      this.filter = filter;
      this.options = mapOptions;
    },
    getClusterFromUrl() {
      const hasLatLng = this.$route.query.latlng !== undefined && this.$route.query.latlng !== "";

      if (!hasLatLng) {
        return undefined;
      }

      return {
        q: this.filter.q,
        s: this.filter.s,
        latlng: this.$route.query.latlng,
      };
    },
    openCluster: function (cluster) {
      this.cluster = cluster;
      this.showCluster = true;
    },
    openClusterFromUrl: function () {
      const cluster = this.getClusterFromUrl();

      if (!cluster) {
        return;
      }

      this.openCluster(cluster);
    },
    selectClusterByCoords: function (latNorth, lngEast, latSouth, lngWest) {
      this.openCluster({
        q: this.filter.q,
        s: this.filter.s,
        latlng: [latNorth, lngEast, latSouth, lngWest].join(","),
      });
    },
    selectClusterById: function (clusterId) {
      if (this.showCluster) {
        this.showCluster = false;
      }

      this.getClusterFeatures(clusterId, -1, (clusterFeatures) => {
        let latNorth, lngEast, latSouth, lngWest;

        for (const feature of clusterFeatures) {
          const [lng, lat] = feature.geometry.coordinates;
          if (latNorth === undefined || lat > latNorth) {
            latNorth = lat;
          }
          if (lngEast === undefined || lng > lngEast) {
            lngEast = lng;
          }
          if (latSouth === undefined || lat < latSouth) {
            latSouth = lat;
          }
          if (lngWest === undefined || lng < lngWest) {
            lngWest = lng;
          }
        }

        this.selectClusterByCoords(latNorth, lngEast, latSouth, lngWest);
      });
    },
    closeCluster: function () {
      this.cluster = {};
      this.showCluster = false;
    },
    query: function () {
      return this.$route.query.q ? this.$route.query.q : "";
    },
    scope: function () {
      return this.$route.params.s ? this.$route.params.s : "";
    },
    openPhoto(uid) {
      // Abort if uid is empty or results aren't loaded.
      if (!uid || this.loading || !this.result || !this.result.features || this.result.features.length === 0) {
        return;
      }

      // Get request parameters.
      const options = {
        params: {
          near: uid,
          count: 1000,
        },
      };

      if (this.filter.s) {
        options.params.s = this.filter.s;
      }

      this.loading = true;

      // Perform get request to find nearby photos.
      return $api
        .get("geo/view", options)
        .then((r) => {
          if (r && r.data && r.data.length > 0) {
            // Show photos.
            this.$lightbox.openModels(Thumb.wrap(r.data), 0);
          } else {
            // Don't open lightbox if nothing was found.
            this.$notify.warn(this.$gettext("No pictures found"));
          }
        })
        .finally(() => {
          this.loading = false;
        });
    },
    formChange() {
      if (this.loading) {
        return;
      }

      this.$router.push({
        query: {
          q: this.filter.q,
        },
      });
    },
    clearQuery() {
      if (this.loading) {
        return;
      }

      this.$router.push({
        query: {},
      });
    },
    updateQuery() {
      if (this.loading) {
        return false;
      }

      if (this.query() !== this.filter.q) {
        if (this.filter.s) {
          this.$router.replace({
            name: "places_view",
            params: { s: this.filter.s },
            query: { q: this.filter.q },
          });
        } else if (this.filter.q) {
          this.$router.replace({ name: "places", query: { q: this.filter.q } });
        } else {
          this.$router.replace({ name: "places" });
        }

        return true;
      }

      return false;
    },
    searchParams() {
      const params = {
        count: this.maxCount,
        offset: 0,
      };

      Object.assign(params, this.filter);

      if (this.staticFilter) {
        Object.assign(params, this.staticFilter);
      }

      return params;
    },
    reload() {
      if (!this.initialized || this.loading) {
        return;
      }

      this.search(true);
    },
    reset() {
      Object.assign(this.result, { features: [] });
      const map = this.map;

      if (!map || typeof map.getSource !== "function") {
        return;
      }

      const source = map.getSource("photos");

      if (!source || typeof source.setData !== "function") {
        return;
      }

      source.setData(this.result);

      this.updateMarkers();
    },
    search(force) {
      if (this.loading) {
        return;
      }

      // Do not query the same data more than once unless search results need to be updated.
      if (!force && this.initialized && JSON.stringify(this.lastFilter) === JSON.stringify(this.filter)) {
        return;
      }

      this.loading = true;

      this.closeCluster();

      Object.assign(this.lastFilter, this.filter);

      this.updateQuery();

      // Compose query params.
      const options = {
        params: this.searchParams(),
      };

      // Fetch results from server.
      return $api
        .get("geo", options)
        .then((response) => {
          if (!response.data.features || response.data.features.length === 0) {
            this.reset();
            this.initialized = true;
            this.loading = false;

            this.$notify.warn(this.$gettext("No pictures found"));

            return;
          }

          this.result = response.data;

          const map = this.map;

          if (!map || typeof map.getSource !== "function") {
            this.initialized = true;
            this.loading = false;
            return;
          }

          const source = map.getSource("photos");

          if (!source || typeof source.setData !== "function") {
            this.initialized = true;
            this.loading = false;
            return;
          }

          source.setData(this.result);

          if (this.filter.q || !this.initialized) {
            if (typeof map.fitBounds === "function") {
              map.fitBounds(this.result.bbox, {
                maxZoom: 17,
                padding: 100,
                duration: this.animate,
                essential: false,
                animate: true,
              });
            }
          }

          this.initialized = true;
          this.loading = false;

          this.updateMarkers();
        })
        .catch(() => {
          this.reset();
          this.initialized = true;
          this.loading = false;
        });
    },
    renderMap() {
      const lastProjection = this.getProjection(true);

      this.map = new maplibregl.Map(this.options);
      this.map.setLanguage(this.$config.values.settings.ui.language.split("-")[0]);

      // Get informed about projection type changes.
      this.map.on("projectiontransition", (ev) => this.onProjectionChange(ev));

      // Restore last used projection type, if any.
      if (lastProjection) {
        this.projection = lastProjection;

        // Restore last used projection type.
        this.map.on("style.load", () => {
          this.map.setProjection({
            type: lastProjection,
          });
        });
      }

      const controlPos = "top-right";

      // Add map navigation control.
      this.map.addControl(
        new maplibregl.NavigationControl({
          visualizePitch: true,
          showZoom: true,
          showCompass: true,
        }),
        controlPos
      );

      // Add 3D terrain toggle control, if supported.
      if (this.terrain[this.style]) {
        this.map.addControl(
          new maplibregl.TerrainControl({
            source: this.terrain[this.style],
            exaggeration: 1,
          })
        );
      }

      // Add 3D globe toggle control.
      this.map.addControl(new maplibregl.GlobeControl());

      // Add fullscreen toggle control, except on mobile devices.
      if (!this.$isMobile) {
        this.map.addControl(
          new maplibregl.FullscreenControl({ container: document.querySelector("body") }),
          controlPos
        );
      }

      // Add locate position control.
      this.map.addControl(
        new maplibregl.GeolocateControl({
          positionOptions: {
            enableHighAccuracy: true,
          },
          trackUserLocation: true,
        }),
        controlPos
      );

      // Add style switcher control.
      if (this.mapStyles.length > 1) {
        this.map.addControl(new MapStyleControl(this.mapStyles, this.style, this.setStyle), controlPos);
      }

      // Add map scale control.
      this.map.addControl(new maplibregl.ScaleControl({ maxWidth: 120, unit: "metric" }), "bottom-left");

      this.map.on("load", () => this.onMapLoad());
    },
    getClusterFeatures(clusterId, limit, callback) {
      const map = this.map;

      if (!map || typeof map.getSource !== "function") {
        return;
      }

      const source = map.getSource("photos");

      if (!source || typeof source.getClusterLeaves !== "function") {
        return;
      }

      source.getClusterLeaves(clusterId, limit, undefined).then((clusterFeatures) => {
        callback(clusterFeatures);
      });
    },
    getClusterSizeFromItemCount(itemCount) {
      if (itemCount >= 10000) {
        return 74;
      } else if (itemCount >= 1000) {
        return 70;
      } else if (itemCount >= 750) {
        return 68;
      } else if (itemCount >= 200) {
        return 66;
      } else if (itemCount >= 100) {
        return 64;
      }

      return 60;
    },
    abbreviateCount(val) {
      const value = Number.parseInt(val);
      if (value >= 1000) {
        return (value / 1000).toFixed(0).toString() + "k";
      }
      return value;
    },
    updateMarkers() {
      // Busy loading data from the server?
      if (this.loading) {
        // Skip updating map markers.
        return;
      }

      const map = this.map;

      if (!map || typeof map.querySourceFeatures !== "function") {
        return;
      }

      // Maps may emit resize events while a style reloads; skip processing until the source is ready.
      const source = typeof map.getSource === "function" ? map.getSource("photos") : null;
      if (!source) {
        return;
      }

      const newMarkers = {};

      // Get map features from the "photos" layer.
      const features = map.querySourceFeatures("photos");

      // Get API token required to show thumbnails.
      let token = this.$config.previewToken;

      // Loop through photos and clusters.
      for (let i = 0; i < features.length; i++) {
        let coords = features[i].geometry.coordinates;
        let props = features[i].properties;

        // Is it a cluster?
        if (props.cluster) {
          // Update cluster marker.

          // Attention: Do not confuse with photo feature IDs.
          // Clusters have their own ID number range!
          let id = -1 * props.cluster_id;

          let marker = this.markers[id];

          if (!marker) {
            const size = this.getClusterSizeFromItemCount(props.point_count);
            let el = document.createElement("div");

            el.style.width = `${size}px`;
            el.style.height = `${size}px`;

            const imageContainer = document.createElement("div");
            imageContainer.className = "marker cluster-marker";

            source
              .getClusterLeaves(props.cluster_id, 4, 0)
              .then((clusterFeatures) => {
                const previewImageCount = clusterFeatures.length >= 4 ? 4 : clusterFeatures.length > 1 ? 2 : 1;
                const images = Array(previewImageCount)
                  .fill(null)
                  .map((a, i) => {
                    const feature = clusterFeatures[Math.floor((clusterFeatures.length * i) / previewImageCount)];
                    const image = document.createElement("div");
                    image.style.backgroundImage = `url(${this.$config.contentUri}/t/${feature.properties.Hash}/${token}/tile_${50})`;
                    return image;
                  });

                imageContainer.append(...images);
              })
              .catch(() => {});

            const counterBubble = document.createElement("div");

            counterBubble.className = "badge";
            counterBubble.innerText = this.abbreviateCount(props.point_count);

            el.append(imageContainer);
            el.append(counterBubble);
            marker = this.markers[id] = new maplibregl.Marker({
              element: el,
            }).setLngLat(coords);
          } else {
            marker.setLngLat(coords);
          }

          const clusterId = props.cluster_id;
          this.ensureMarkerClick(id, marker, clusterId, () => () => this.selectClusterById(clusterId));

          newMarkers[id] = marker;

          if (!this.markersOnScreen[id]) {
            marker.addTo(map);
          }
        } else {
          // Update photo marker.
          const id = features[i].id;

          let marker = this.markers[id];
          if (!marker) {
            const el = document.createElement("div");
            el.className = "marker";
            el.title = props.Title;
            el.style.backgroundImage = `url(${this.$config.contentUri}/t/${props.Hash}/${token}/tile_50)`;
            el.style.width = "50px";
            el.style.height = "50px";

            marker = this.markers[id] = new maplibregl.Marker({
              element: el,
            }).setLngLat(coords);
          } else {
            marker.setLngLat(coords);
          }

          const photoUid = props.UID;
          this.ensureMarkerClick(id, marker, photoUid, () => () => this.openPhoto(photoUid));

          newMarkers[id] = marker;

          if (!this.markersOnScreen[id]) {
            marker.addTo(map);
          }
        }
      }

      // Hide markers that are not currently visible.
      Object.keys(this.markersOnScreen).forEach((id) => {
        if (!newMarkers[id]) {
          const marker = this.markersOnScreen[id];
          this.clearMarkerClick(id, marker);
          if (marker && typeof marker.remove === "function") {
            marker.remove();
          }
        }
      });

      // Remember the markers displayed on the map.
      this.markersOnScreen = newMarkers;
    },
    minimizeAttribCtrl() {
      if (this.$refs.map instanceof HTMLElement) {
        const attrCtrl = this.$refs.map.querySelector(".maplibregl-ctrl-attrib");

        if (attrCtrl && attrCtrl instanceof HTMLElement) {
          attrCtrl.classList?.remove("maplibregl-compact-show");
          attrCtrl.removeAttribute("open");
        }
      }
    },
    onMapLoad() {
      this.minimizeAttribCtrl();

      // Get projection type from map.
      this.projection = this.getProjection();

      // Add 'photos' data source.
      this.map.addSource("photos", {
        type: "geojson",
        data: { type: "FeatureCollection", features: [] },
        cluster: true,
        clusterMaxZoom: 17, // Max zoom to cluster points on
        clusterRadius: 80, // Radius of each cluster when clustering points (defaults to 50)
      });

      // Add 'clusters' layer.
      this.map.addLayer({
        id: "clusters",
        type: "circle",
        source: "photos",
        filter: ["has", "point_count"],
        paint: {
          "circle-color": "#FFFFFF",
          "circle-opacity": 0,
          "circle-radius": 0,
        },
      });

      // Example of dynamic map cluster rendering:
      // https://maplibre.org/maplibre-gl-js/docs/examples/cluster-html/
      if (this._mapDataHandler && typeof this.map.off === "function") {
        this.map.off("data", this._mapDataHandler);
      }

      this._mapDataHandler = (e) => {
        if (e?.sourceId === "photos" && e.isSourceLoaded) {
          this.updateMarkers();
        }
      };

      this.map.on("data", this._mapDataHandler);

      // Add additional event handlers to update the marker previews.
      this.map.on("move", this.updateMarkers);
      this.map.on("moveend", this.updateMarkers);
      this.map.on("resize", this.updateMarkers);
      this.map.on("idle", this.updateMarkers);

      // Load pictures.
      this.search().finally(() => {
        this.loaded = true;
      });
    },
  },
};
</script>
