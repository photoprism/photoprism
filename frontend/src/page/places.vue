<template>
  <v-container fluid fill-height :class="$config.aclClasses('places')" class="pa-0 p-page p-page-places">
    <div id="map" style="width: 100%; height: 100%;">
      <div v-if="canSearch" class="map-control">
        <div class="maplibregl-ctrl maplibregl-ctrl-group map-control-search">
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
    <v-dialog v-model="showClusterPictures" overflowed width="100%">
      <v-card min-height="80vh">
        <p-page-photos
          v-if="showClusterPictures"
          :static-filter="selectedClusterBounds"
          :on-close="unselectCluster"
          sticky-toolbar
        />
      </v-card>
    </v-dialog>
  </v-container>
</template>

<script>
import maplibregl from "maplibre-gl";
import Api from "common/api";
import Thumb from "model/thumb";
import PPagePhotos from 'page/photos.vue';

export default {
  name: 'PPagePlaces',
  components: {
    PPagePhotos,
  },
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
      style: "",
      terrain: {
        'topo-v2': 'terrain_rgb',
        'outdoor-v2': 'terrain-rgb',
        '414c531c-926d-4164-a057-455a215c0eee': 'terrain_rgb_virtual',
      },
      attribution: '<a href="https://www.maptiler.com/copyright/" target="_blank">&copy; MapTiler</a> <a href="https://www.openstreetmap.org/copyright" target="_blank">&copy; OpenStreetMap contributors</a>',
      maxCount: 500000,
      options: {},
      mapFont: ["Open Sans Regular"],
      result: {},
      filter: {q: this.query(), s: this.scope()},
      lastFilter: {},
      config: this.$config.values,
      settings: this.$config.values.settings.maps,
      selectedClusterBounds: undefined,
      showClusterPictures: false,
    };
  },
  watch: {
    '$route'() {

      const clusterWasOpenBeforeRouterChange = this.selectedClusterBounds !== undefined;
      const clusterIsOpenAfterRouteChange = this.getSelectedClusterFromUrl() !== undefined;
      const lastRouteChangeWasClusterOpenOrClose = clusterWasOpenBeforeRouterChange !== clusterIsOpenAfterRouteChange;

      if (lastRouteChangeWasClusterOpenOrClose) {
        this.updateSelectedClusterFromUrl();

        /**
         * dont touch any filters or searches if the only action taken was
         * opening or closing a cluster.
         * This currently assumes that when a cluster was opened or closed,
         * nothing else changed. I currently can't think of a scenario, where
         * a route-change is triggered by the user wanting to open/close a cluster
         * AND for example update the filter at the same time.
         *
         * Without this, opening or closing a cluster triggers a search, even
         * though no search parameter changed. Also without this, closing a
         * cluster resets the filter, because closing a cluster is done via
         * backwards navigation.
         * (closing is cluster is done via backwards navigation so that it can
         * be closed using the back-button. This is especially useful on android
         * smartphones)
         */
        return;
      }

      this.filter.q = this.query();
      this.filter.s = this.scope();
      this.lastFilter = {};

      this.search();
    },
    showClusterPictures:function(newValue, old){
      if(!newValue){
        this.unselectCluster();
      }
    }
  },
  mounted() {
    this.configureMap().then(() => this.renderMap());
    this.updateSelectedClusterFromUrl();
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

        switch (s.style) {
          case "basic":
          case "offline":
            this.style = "";
            break;
          case "hybrid":
            this.style = "414c531c-926d-4164-a057-455a215c0eee";
            break;
          case "outdoor":
            this.style = "outdoor-v2";
            break;
          case "topographique":
            this.style = "topo-v2";
            break;
          default:
            this.style = s.style;
        }

        if (!mapKey && this.style !== "low-resolution") {
          this.style = "";
        }

        let mapOptions = {
          container: "map",
          style: "https://api.maptiler.com/maps/" + this.style + "/style.json?key=" + mapKey,
          glyphs: "https://api.maptiler.com/fonts/{fontstack}/{range}.pbf?key=" + mapKey,
          attributionControl: true,
          customAttribution: this.attribution,
          zoom: 0,
        };

        if (this.style === "") {
          mapOptions = {
            container: "map",
            style: "https://cdn.photoprism.app/maps/default.json",
            glyphs: `https://cdn.photoprism.app/maps/font/{fontstack}/{range}.pbf`,
            attributionControl: true,
            zoom: 0,
          };
        } else if (this.style === "low-resolution") {
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
            customAttribution: '',
            zoom: 0,
          };
        }

        this.filter = filter;
        this.options = mapOptions;
      });
    },
    getSelectedClusterFromUrl() {
      const clusterIsSelected = this.$route.query.selectedCluster !== undefined
                            && this.$route.query.selectedCluster !== '';
      if (!clusterIsSelected) {
        return undefined;
      }

      const [latmin, latmax, lngmin, lngmax] = this.$route.query.selectedCluster.split(',');
      return {latmin, latmax, lngmin, lngmax};
    },
    updateSelectedClusterFromUrl: function() {
      this.selectedClusterBounds = this.getSelectedClusterFromUrl();
      this.showClusterPictures = this.selectedClusterBounds !== undefined;
    },
    selectClusterByCoords: function(latMin, latMax, lngMin, lngMax) {
      this.$router.push({
        query: {
          selectedCluster: [latMin, latMax, lngMin, lngMax].join(','),
        },
        params: this.filter,
      });
    },
    selectClusterById: function(clusterId) {
      this.getClusterFeatures(clusterId, -1, (clusterFeatures) => {
        let latMin,latMax,lngMin,lngMax;
        for (const feature of clusterFeatures) {
          const [lng,lat] = feature.geometry.coordinates;
          if (latMin === undefined || lat < latMin) {
            latMin = lat;
          }
          if (latMax === undefined || lat > latMax) {
            latMax = lat;
          }
          if (lngMin === undefined || lng < lngMin) {
            lngMin = lng;
          }
          if (lngMax === undefined || lng > lngMax) {
            lngMax = lng;
          }
        }

        this.selectClusterByCoords(latMin, latMax, lngMin, lngMax);
      });
    },
    unselectCluster: function() {
      const aClusterIsSelected = this.getSelectedClusterFromUrl() !== undefined;
      if (aClusterIsSelected) {
        // it shouldn't matter wether a cluster was closed by pressing the back
        // button on a browser or the x-button on the dialog. We therefore make
        // both actions do the exact same thing: navigate backwards
        this.$router.go(-1);
      }
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

      // Show navigation control.
      this.map.addControl(new maplibregl.NavigationControl({
        visualizePitch: true,
        showZoom: true,
        showCompass: true
      }), controlPos);

      // Show terrain control, if supported.
      if (this.terrain[this.style]) {
        this.map.addControl(new maplibregl.TerrainControl({
          source: this.terrain[this.style],
          exaggeration: 1
        }));
      }

      // Show fullscreen control.
      this.map.addControl(new maplibregl.FullscreenControl({container: document.querySelector('body')}), controlPos);

      // Show locate control.
      this.map.addControl(new maplibregl.GeolocateControl({
        positionOptions: {
          enableHighAccuracy: true
        },
        trackUserLocation: true
      }), controlPos);

      this.map.on("load", () => this.onMapLoad());
    },
    getClusterFeatures(clusterId, limit, callback) {
      this.map.getSource('photos').getClusterLeaves(clusterId, limit, undefined, (error, clusterFeatures) => {
        callback(clusterFeatures);
      });
    },
    getMultipleClusterFeatures(clusterIds, callback) {
      const result = {};
      let handledClusterLeaveResultCount = 0;
      for (const clusterId of clusterIds) {
        this.getClusterFeatures(clusterId, 100, (clusterFeatures) => {
          result[clusterId] = clusterFeatures;
          handledClusterLeaveResultCount += 1;

          if (handledClusterLeaveResultCount === clusterIds.length) {
            callback(result);
          }
        });
      }
    },
    getClusterRadiusFromItemCount(itemCount) {
      // see config of cluster-layer for these values
      if (itemCount >= 750) {
        return 50;
      }
      if (itemCount >= 100) {
        return 40;
      }

      return 30;
    },
    updateMarkers() {
      if (this.loading) return;
      let newMarkers = {};
      let features = this.map.querySourceFeatures("photos");
      const clusterIds = features
        .filter(feature => feature.properties.cluster)
        .map(feature => feature.properties.cluster_id);

      this.getMultipleClusterFeatures(clusterIds, (clusterFeaturesById) => {
        for (let i = 0; i < features.length; i++) {
          let coords = features[i].geometry.coordinates;
          let props = features[i].properties;
          let id = features[i].id;

          let marker = this.markers[id];
          let token = this.$config.previewToken;
          if (!marker) {
            let el = document.createElement('div');
            if (props.cluster) {
              const radius = this.getClusterRadiusFromItemCount(props.point_count);
              el.style.width = `${radius * 2}px`;
              el.style.height = `${radius * 2}px`;

              const imageContainer = document.createElement('div');
              imageContainer.className = 'marker cluster-marker';

              const clusterFeatures = clusterFeaturesById[props.cluster_id];
              const previewImageCount = clusterFeatures.length > 3 ? 4 : 2;
              const images = Array(previewImageCount)
                .fill()
                .map((a,i) => {
                  const feature = clusterFeatures[Math.floor(clusterFeatures.length * i / previewImageCount)];
                  const imageHash = feature.properties.Hash;
                  const image = document.createElement('div');
                  image.style.backgroundImage = `url(${this.$config.contentUri}/t/${imageHash}/${token}/tile_${50})`;
                  return image;
                });
              imageContainer.append(...images);

              const counterBubble = document.createElement('div');
              counterBubble.className = 'counter-bubble primary-button theme--light';
              counterBubble.innerText = clusterFeatures.length > 99 ? '99+' : clusterFeatures.length;

              el.append(imageContainer);
              el.append(counterBubble);
              el.addEventListener('click', () => {
                this.selectClusterById(props.cluster_id);
              });
            } else {
              el.className = 'marker';
              el.title = props.Title;
              el.style.backgroundImage = `url(${this.$config.contentUri}/t/${props.Hash}/${token}/tile_50)`;
              el.style.width = '50px';
              el.style.height = '50px';

              el.addEventListener('click', () => this.openPhoto(props.UID));
            }
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
      });
    },
    onMapLoad() {
      this.map.addSource('photos', {
        type: 'geojson',
        data: null,
        cluster: true,
        clusterMaxZoom: 18, // Max zoom to cluster points on
        clusterRadius: 80 // Radius of each cluster when clustering points (defaults to 50)
      });

      // TODO: can this rendering of empty colored circles be removed?
      this.map.addLayer({
        id: 'clusters',
        type: 'circle',
        source: 'photos',
        filter: ['has', 'point_count'],
        paint: {
          'circle-color': [
            'step',
            ['get', 'point_count'],
            'transparent',
            100,
            'transparent',
            750,
            'transparent'
          ],
          'circle-radius': [
            'step',
            ['get', 'point_count'],
            30,
            100,
            40,
            750,
            50
          ]
        }
      });

      this.map.on('idle', this.updateMarkers);

      this.search();
    },
  },
};
</script>

