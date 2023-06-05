<template>
  <v-container fluid fill-height :class="$config.aclClasses('places')" class="pa-0 p-page p-page-places">
    <div id="map" style="width: 100%; height: 100%;">
      <div v-if="canSearch" class="map-control">
        <div class="maplibregl-ctrl maplibregl-ctrl-group">
          <v-text-field v-model.lazy.trim="filter.q"
                        solo hide-details clearable flat single-line validate-on-blur
                        class="input-search pa-0 ma-0"
                        :label="$gettext('Search')"
                        prepend-inner-icon="search"
                        browser-autocomplete="off"
                        autocorrect="off"
                        autocapitalize="none"
                        color="secondary-dark"
                        @click:clear="clearQuery"
                        @keyup.enter.native="formChange"
          ></v-text-field>
        </div>
      </div>
    </div>
  </v-container>
</template>

<script>
import maplibregl from "maplibre-gl";
import Api from "common/api";
import Thumb from "model/thumb";

export default {
  name: 'PPagePlaces',
  props: {
    staticFilter: {
      type: Object,
      default: () => {
      },
    },
  },
  data() {
    return {
      canSearch: this.$config.allow("places", "search"),
      initialized: false,
      map: null,
      markers: {},
      markersOnScreen: {},
      loading: false,
      url: "",
      attribution: '<a href="https://www.maptiler.com/copyright/" target="_blank">&copy; MapTiler</a> <a href="https://www.openstreetmap.org/copyright" target="_blank">&copy; OpenStreetMap contributors</a>',
      maxCount: 500000,
      options: {},
      mapFont: ["Open Sans Regular"],
      result: {},
      filter: {q: this.query(), s: this.scope()},
      lastFilter: {},
      config: this.$config.values,
      settings: this.$config.values.settings.maps,
    };
  },
  watch: {
    '$route'() {
      this.filter.q = this.query();
      this.filter.s = this.scope();
      this.lastFilter = {};

      this.search();
    }
  },
  mounted() {
    this.$scrollbar.hide();
    this.configureMap().then(() => this.renderMap());
  },
  destroyed() {
    this.$scrollbar.show();
  },
  methods: {
    configureMap() {
      return this.$config.load().finally(() => {
        const s = this.$config.values.settings.maps;
        const filter = {
          q: this.query(),
          s: this.scope(),
        };

        let mapKey = "";

        if (this.$config.has("mapKey")) {
          // Remove non-alphanumeric characters from key.
          mapKey = this.$config.get("mapKey").replace(/[^a-z0-9]/gi, '');
        }

        const settings = this.$config.settings();

        if (settings && settings.features.private) {
          filter.public = "true";
        }

        if (settings && settings.features.review && (!this.staticFilter || !("quality" in this.staticFilter))) {
          filter.quality = "3";
        }

        let mapsStyle = s.style;

        if (!mapKey && mapsStyle !== "low-resolution") {
          mapsStyle = "";
        }

        let mapOptions = {
          container: "map",
          style: "https://api.maptiler.com/maps/" + mapsStyle + "/style.json?key=" + mapKey,
          glyphs: "https://api.maptiler.com/fonts/{fontstack}/{range}.pbf?key=" + mapKey,
          attributionControl: true,
          customAttribution: this.attribution,
          zoom: 0,
        };

        if (mapsStyle === "") {
          mapOptions = {
            container: "map",
            style:  "https://cdn.photoprism.app/maps/default.json",
            glyphs: `https://cdn.photoprism.app/maps/font/{fontstack}/{range}.pbf`,
            attributionControl: true,
            zoom: 0,
          };
          this.url = '';
        } else if (mapsStyle === "low-resolution") {
          mapOptions = {
            container: "map",
            style: {
              "version": 8,
              "sources": {
                "world": {
                  "type": "geojson",
                  "data": `${this.$config.staticUri}/geo/world.json`,
                  "maxzoom": 6
                }
              },
              "glyphs": `${this.$config.staticUri}/font/{fontstack}/{range}.pbf`,
              "layers": [
                {
                  "id": "background",
                  "type": "background",
                  "paint": {
                    "background-color": "#aadafe"
                  }
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
                  "id": "country-abbrev",
                  "type": "symbol",
                  "source": "world",
                  "maxzoom": 3,
                  "layout": {
                    "text-field": "{abbrev}",
                    "text-font": ["Open Sans Semibold"],
                    "text-transform": "uppercase",
                    "text-max-width": 20,
                    "text-size": {
                      "stops": [[3, 10], [4, 11], [5, 12], [6, 16]]
                    },
                    "text-letter-spacing": {
                      "stops": [[4, 0], [5, 1], [6, 2]]
                    },
                    "text-line-height": {
                      "stops": [[5, 1.2], [6, 2]]
                    }
                  },
                  "paint": {
                    "text-halo-color": "#fff",
                    "text-halo-width": 1
                  },
                },
                {
                  "id": "country-border",
                  "type": "line",
                  "source": "world",
                  "paint": {
                    "line-color": "#226688",
                    "line-opacity": 0.25,
                    "line-dasharray": [6, 2, 2, 2],
                    "line-width": 1.2
                  }
                },
                {
                  "id": "country-name",
                  "type": "symbol",
                  "minzoom": 3,
                  "source": "world",
                  "layout": {
                    "text-field": "{name}",
                    "text-font": ["Open Sans Semibold"],
                    "text-max-width": 20,
                    "text-size": {
                      "stops": [[3, 10], [4, 11], [5, 12], [6, 16]]
                    }
                  },
                  "paint": {
                    "text-halo-color": "#fff",
                    "text-halo-width": 1
                  },
                },
              ],
            },
            attributionControl: false,
            customAttribution:  '',
            zoom: 0,
          };
          this.url = '';
        } else {
          this.url = 'https://api.maptiler.com/maps/' + mapsStyle + '/{z}/{x}/{y}.png?key=' + mapKey;
        }

        this.filter = filter;
        this.options = mapOptions;
      });
    },
    query: function () {
      return this.$route.params.q ? this.$route.params.q : '';
    },
    scope: function () {
      return this.$route.params.s ? this.$route.params.s : '';
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
      return Api.get("geo/view", options).then((r) => {
        if (r && r.data && r.data.length > 0) {
          // Show photos.
          this.$viewer.show(Thumb.wrap(r.data), 0);
        } else {
          // Don't open viewer if nothing was found.
          this.$notify.warn(this.$gettext("No pictures found"));
        }
      }).finally(() => {
        this.loading = false;
      });
    },
    formChange() {
      if (this.loading) return;
      this.search();
    },
    clearQuery() {
      this.filter.q = '';
      this.search();
    },
    updateQuery() {
      if (this.loading) return;

      if (this.query() !== this.filter.q) {
        if (this.filter.s) {
          this.$router.replace({name: "places_scope", params: {s: this.filter.s, q: this.filter.q}});
        } else if (this.filter.q) {
          this.$router.replace({name: "places_query", params: {q: this.filter.q}});
        } else {
          this.$router.replace({name: "places"});
        }
      }
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
    search() {
      if (this.loading) return;

      // Don't query the same data more than once
      if (JSON.stringify(this.lastFilter) === JSON.stringify(this.filter)) return;
      this.loading = true;

      Object.assign(this.lastFilter, this.filter);

      this.updateQuery();

      // Compose query params.
      const options = {
        params: this.searchParams(),
      };

      // Fetch results from server.
      return Api.get("geo", options).then((response) => {
        if (!response.data.features || response.data.features.length === 0) {
          this.loading = false;

          this.$notify.warn(this.$gettext("No pictures found"));

          return;
        }

        this.result = response.data;

        this.map.getSource("photos").setData(this.result);

        if (this.filter.q || !this.initialized) {
          this.map.fitBounds(this.result.bbox, {
            maxZoom: 17,
            padding: 100,
            duration: this.settings.animate,
            essential: false,
            animate: this.settings.animate > 0
          });
        }

        this.initialized = true;

        this.updateMarkers();
      }).finally(() => {
        this.loading = false;
      });
    },
    renderMap() {
      this.map = new maplibregl.Map(this.options);
      this.map.setLanguage(this.$config.values.settings.ui.language.split("-")[0]);

      const controlPos = this.$rtl ? 'top-left' : 'top-right';

      this.map.addControl(new maplibregl.NavigationControl({showCompass: true}), controlPos);
      this.map.addControl(new maplibregl.FullscreenControl({container: document.querySelector('body')}), controlPos);
      this.map.addControl(new maplibregl.GeolocateControl({
        positionOptions: {
          enableHighAccuracy: true
        },
        trackUserLocation: true
      }), controlPos);

      this.map.on("load", () => this.onMapLoad());
    },
    updateMarkers() {
      if (this.loading) return;
      let newMarkers = {};
      let features = this.map.querySourceFeatures("photos");

      for (let i = 0; i < features.length; i++) {
        let coords = features[i].geometry.coordinates;
        let props = features[i].properties;
        if (props.cluster) continue;
        let id = features[i].id;

        let marker = this.markers[id];
        let token = this.$config.previewToken;
        if (!marker) {
          let el = document.createElement('div');
          el.className = 'marker';
          el.title = props.Title;
          el.style.backgroundImage = `url(${this.$config.contentUri}/t/${props.Hash}/${token}/tile_50)`;
          el.style.width = '50px';
          el.style.height = '50px';

          el.addEventListener('click', () => this.openPhoto(props.UID));
          marker = this.markers[id] = new maplibregl.Marker({
            element: el
          }).setLngLat(coords);
        } else {
          marker.setLngLat(coords);
        }

        newMarkers[id] = marker;

        if (!this.markersOnScreen[id]) {
          marker.addTo(this.map);
        }
      }
      for (let id in this.markersOnScreen) {
        if (!newMarkers[id]) {
          this.markersOnScreen[id].remove();
        }
      }
      this.markersOnScreen = newMarkers;
    },
    onMapLoad() {
      this.map.addSource('photos', {
        type: 'geojson',
        data: null,
        cluster: true,
        clusterMaxZoom: 14, // Max zoom to cluster points on
        clusterRadius: 50 // Radius of each cluster when clustering points (defaults to 50)
      });

      this.map.addLayer({
        id: 'clusters',
        type: 'circle',
        source: 'photos',
        filter: ['has', 'point_count'],
        paint: {
          'circle-color': [
            'step',
            ['get', 'point_count'],
            '#2DC4B2',
            100,
            '#3BB3C3',
            750,
            '#669EC4'
          ],
          'circle-radius': [
            'step',
            ['get', 'point_count'],
            20,
            100,
            30,
            750,
            40
          ]
        }
      });

      this.map.addLayer({
        id: 'cluster-count',
        type: 'symbol',
        source: 'photos',
        filter: ['has', 'point_count'],
        layout: {
          'text-field': '{point_count_abbreviated}',
          'text-font': this.mapFont,
          'text-size': 13
        }
      });

      this.map.on('render', this.updateMarkers);

      this.map.on('click', 'clusters', (e) => {
        const features = this.map.queryRenderedFeatures(e.point, {
          layers: ['clusters']
        });
        const clusterId = features[0].properties.cluster_id;
        this.map.getSource('photos').getClusterExpansionZoom(
          clusterId,
          (err, zoom) => {
            if (err) return;

            this.map.easeTo({
              center: features[0].geometry.coordinates,
              zoom: zoom
            });
          }
        );
      });

      this.map.on('mouseenter', 'clusters', () => {
        this.map.getCanvas().style.cursor = 'pointer';
      });
      this.map.on('mouseleave', 'clusters', () => {
        this.map.getCanvas().style.cursor = '';
      });

      this.search();
    },
  },
};
</script>

