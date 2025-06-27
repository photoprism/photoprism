<template>
  <v-dialog
    v-model="show"
    :max-width="900"
    :fullscreen="$vuetify.display.mdAndDown"
    :persistent="false"
    class="p-location-dialog"
    @keydown.esc="close"
    @after-leave="onDialogClosed"
  >
    <v-card :tile="$vuetify.display.mdAndDown">
      <v-toolbar
        v-if="$vuetify.display.mdAndDown"
        flat
        color="navigation"
        class="mb-4"
        :density="$vuetify.display.smAndDown ? 'compact' : 'default'"
      >
        <v-btn icon @click.stop="close">
          <v-icon>mdi-close</v-icon>
        </v-btn>
        <v-toolbar-title>
          {{ $gettext("Set Location") }}
        </v-toolbar-title>
      </v-toolbar>
      <v-card-title v-else class="d-flex justify-start align-center ga-3">
        <v-icon size="28" color="primary">mdi-map-marker</v-icon>
        <h6 class="text-h6">{{ $gettext("Set Location") }}</h6>
      </v-card-title>
      <v-card-text class="pt-4">
        <div class="d-flex flex-column flex-md-row">
          <div class="flex-grow-1 position-relative mb-4 mb-md-0">
            <div ref="map" class="p-map" style="height: 50vh; min-height: 300px; width: 100%; border-radius: 4px"></div>
          </div>

          <div
            class="map-sidebar d-flex flex-column ml-0 ml-md-4"
            :style="{
              width: $vuetify.display.smAndDown ? '100%' : '300px',
              maxWidth: $vuetify.display.smAndDown ? '100%' : '300px',
              minWidth: 0,
            }"
          >
            <div class="d-flex flex-column flex-grow-1 ga-3">
              <div>
                <v-menu
                  v-model="showSearchMenu"
                  :close-on-content-click="false"
                  location="bottom"
                  origin="top"
                  max-height="300"
                >
                  <template #activator="{ props }">
                    <v-text-field
                      v-model="searchQuery"
                      :label="$gettext('Search')"
                      prepend-inner-icon="mdi-magnify"
                      :append-inner-icon="
                        searchLoading ? 'mdi-loading mdi-spin' : searchQuery ? 'mdi-close-circle' : ''
                      "
                      density="compact"
                      variant="outlined"
                      placeholder="e.g., Berlin, New York, Tokyo"
                      v-bind="props"
                      @update:model-value="onSearchQueryChange"
                      @click:append-inner="clearSearch"
                      @focus="onSearchFocus"
                      @blur="onSearchBlur"
                    ></v-text-field>
                  </template>
                  <v-list v-if="searchResults.length > 0" density="compact">
                    <v-list-item
                      v-for="place in searchResults"
                      :key="place.id"
                      :title="place.formatted"
                      @click="onPlaceSelected(place)"
                    >
                      <template #prepend>
                        <v-icon>mdi-map-marker</v-icon>
                      </template>
                    </v-list-item>
                  </v-list>
                  <v-list v-else-if="searchQuery && searchQuery.length >= 2 && !searchLoading">
                    <v-list-item>
                      <v-list-item-title>{{ $gettext("No results found") }}</v-list-item-title>
                    </v-list-item>
                  </v-list>
                </v-menu>
              </div>

              <div v-if="locationInfo">
                <div class="text-subtitle-2 mb-2">{{ $gettext("Location Details") }}</div>
                <div class="text-body-2">
                  {{ simplifiedLocationDisplay }}
                </div>
              </div>

              <div>
                <p-coordinate-input
                  :latitude="currentLat"
                  :longitude="currentLng"
                  density="comfortable"
                  placeholder="e.g., 52.5208, 13.4049"
                  :enable-undo="true"
                  :auto-apply="true"
                  :label="$gettext('Coordinates')"
                  @update:latitude="updateLatitude"
                  @update:longitude="updateLongitude"
                  @coordinates-changed="onCoordinatesChanged"
                  @coordinates-cleared="onCoordinatesCleared"
                ></p-coordinate-input>
              </div>

              <div class="d-flex flex-column ga-3 pa-4 border rounded">
                <div class="text-body-2">
                  {{ $gettext("Click on the map to set a location. Drag the marker for precise positioning.") }}
                </div>
                <div class="d-flex ga-2">
                  <v-btn variant="flat" color="button" class="action-cancel" min-width="120" @click.stop="close">
                    {{ $gettext("Cancel") }}
                  </v-btn>
                  <v-btn
                    color="primary"
                    min-width="120"
                    :disabled="!(currentLat !== null && currentLng !== null)"
                    @click="confirm"
                  >
                    {{ $gettext("Apply") }}
                  </v-btn>
                </div>
              </div>
            </div>
          </div>
        </div>
      </v-card-text>
    </v-card>
  </v-dialog>
</template>

<script>
import maplibregl from "common/maplibregl";
import PCoordinateInput from "component/location/coordinate-input.vue";

export default {
  name: "PLocationDialog",
  components: {
    PCoordinateInput,
  },
  props: {
    value: {
      type: Boolean,
      default: false,
    },
    latitude: {
      type: Number,
      default: 0,
    },
    longitude: {
      type: Number,
      default: 0,
    },
  },
  emits: ["update:value", "update:latitude", "update:longitude", "confirm", "close"],
  data() {
    return {
      show: this.value,
      map: null,
      marker: null,
      position: [0.0, 0.0],
      options: {
        container: null,
        style: `https://cdn.photoprism.app/maps/embedded.json`,
        glyphs: `https://cdn.photoprism.app/maps/font/{fontstack}/{range}.pbf`,
        zoom: 12,
        interactive: true,
        attributionControl: { compact: true },
      },
      loaded: false,
      currentLat: this.latitude,
      currentLng: this.longitude,
      locationInfo: null,
      searchQuery: "",
      searchResults: [],
      searchLoading: false,
      searchTimeout: null,
      showSearchMenu: false,
    };
  },
  computed: {
    simplifiedLocationDisplay() {
      if (!this.locationInfo) return "";

      if (this.locationInfo.street && this.locationInfo.formatted) {
        return `${this.locationInfo.street}, ${this.locationInfo.formatted}`;
      } else if (this.locationInfo.street) {
        return this.locationInfo.street;
      } else if (this.locationInfo.formatted) {
        return this.locationInfo.formatted;
      }

      return "";
    },
  },
  watch: {
    value(val) {
      this.show = val;
      if (val) {
        this.currentLat = this.latitude;
        this.currentLng = this.longitude;
        this.$nextTick(() => {
          setTimeout(() => {
            this.initMap();
          }, 100);
        });
      } else {
        // Cleanup map when dialog closes
        this.cleanupMap();
      }
    },
    show(val) {
      this.$emit("update:value", val);
    },
    latitude(val) {
      this.currentLat = val;
      if (this.map && this.loaded) {
        this.updatePosition(val, this.currentLng);
      }
    },
    longitude(val) {
      this.currentLng = val;
      if (this.map && this.loaded) {
        this.updatePosition(this.currentLat, val);
      }
    },
  },
  mounted() {
    if (this.show) {
      this.$nextTick(() => {
        setTimeout(() => {
          this.initMap();
        }, 100);
      });
    }
  },
  beforeUnmount() {
    this.cleanupMap();
  },
  methods: {
    close() {
      this.show = false;
      this.$emit("close");
    },
    confirm() {
      if (this.currentLat !== null && this.currentLng !== null) {
        this.$emit("update:latitude", this.currentLat);
        this.$emit("update:longitude", this.currentLng);
        this.$emit("confirm", {
          latitude: this.currentLat,
          longitude: this.currentLng,
        });
      }
      this.close();
    },
    cleanupMap() {
      if (this.map) {
        this.map.remove();
        this.map = null;
        this.marker = null;
        this.loaded = false;
      }
    },
    onDialogClosed() {
      this.cleanupMap();
      this.locationInfo = null;

      // Clear search state
      this.searchQuery = "";
      this.searchResults = [];
      this.searchLoading = false;
      if (this.searchTimeout) {
        clearTimeout(this.searchTimeout);
        this.searchTimeout = null;
      }
    },
    initMap() {
      if (this.map || !this.$refs.map) {
        return;
      }
      try {
        this.options.container = this.$refs.map;
        if (!this.currentLat || !this.currentLng || (this.currentLat === 0 && this.currentLng === 0)) {
          this.options.zoom = 2;
          this.options.center = [0, 20];
        } else {
          this.options.zoom = 12;
          this.options.center = [this.currentLng, this.currentLat];
        }

        this.map = new maplibregl.Map(this.options);

        this.map.on("styleimagemissing", (e) => {
          const emptyImage = new ImageData(1, 1);
          if (e && e.id) {
            this.map.addImage(e.id, emptyImage);
          }
        });

        this.map.addControl(
          new maplibregl.NavigationControl({
            showCompass: true,
            showZoom: true,
            visualizePitch: false,
          }),
          "top-right"
        );
        this.map.addControl(new maplibregl.ScaleControl({ maxWidth: 80, unit: "metric" }), "bottom-left");
        this.map.addControl(
          new maplibregl.GeolocateControl({
            positionOptions: {
              enableHighAccuracy: true,
            },
            trackUserLocation: true,
          }),
          "top-right"
        );

        this.map.on("error", (e) => {
          console.error("map:", e);
        });

        this.map.on("load", () => {
          this.loaded = true;
          if (this.currentLat && this.currentLng && !(this.currentLat === 0 && this.currentLng === 0)) {
            this.updatePosition(this.currentLat, this.currentLng);
            this.fetchLocationInfo(this.currentLat, this.currentLng);
          }
          this.map.resize();
        });

        this.map.on("click", (e) => {
          this.currentLat = e.lngLat.lat;
          this.currentLng = e.lngLat.lng;
          this.updatePosition(e.lngLat.lat, e.lngLat.lng);
          this.fetchLocationInfo(e.lngLat.lat, e.lngLat.lng);
        });
      } catch (error) {
        console.error("map: initialization failed", error);
        this.loaded = false;
      }
    },
    updatePosition(lat, lng) {
      if (this.map && this.loaded) {
        if (this.position[0] === lng && this.position[1] === lat && this.marker) {
          return;
        }

        // Skip invalid or empty coordinates (0,0)
        if ((lat === 0 && lng === 0) || !lat || !lng) {
          if (this.marker) {
            this.marker.remove();
            this.marker = null;
          }
          return;
        }
        this.position = [lng, lat];

        // Always center map when position changes
        this.map.flyTo({
          center: this.position,
          zoom: 12,
          essential: true,
        });

        if (this.marker) {
          this.marker.setLngLat(this.position);
        } else {
          this.marker = new maplibregl.Marker({
            color: "#3fb4df",
            draggable: true,
          })
            .setLngLat(this.position)
            .addTo(this.map);

          // Update coordinates when marker is dragged
          this.marker.on("dragend", () => {
            const lngLat = this.marker.getLngLat();
            this.currentLat = lngLat.lat;
            this.currentLng = lngLat.lng;
            this.fetchLocationInfo(lngLat.lat, lngLat.lng);
          });
        }
      }
    },
    updateLatitude(lat) {
      this.currentLat = lat;
      this.updatePosition(lat, this.currentLng);
    },
    updateLongitude(lng) {
      this.currentLng = lng;
      this.updatePosition(this.currentLat, lng);
    },
    onCoordinatesChanged(data) {
      if (data.latitude !== 0 || data.longitude !== 0) {
        this.fetchLocationInfo(data.latitude, data.longitude);
      }
    },
    onCoordinatesCleared() {
      this.locationInfo = null;
      if (this.marker) {
        this.marker.remove();
        this.marker = null;
      }
      if (this.map) {
        this.map.flyTo({
          center: [0, 20],
          zoom: 2,
          essential: true,
        });
      }
    },

    fetchLocationInfo(lat, lng) {
      this.$api
        .get(`places/reverse?lat=${lat}&lng=${lng}`)
        .then((response) => {
          if (response.data && response.data.formatted) {
            this.locationInfo = response.data;
          } else {
            this.locationInfo = null;
          }
        })
        .catch((error) => {
          console.error("Reverse geocoding error:", error);
          this.locationInfo = null;
        });
    },

    onSearchQueryChange(query) {
      if (this.searchTimeout) {
        clearTimeout(this.searchTimeout);
      }

      if (!query || query.length < 2) {
        this.searchResults = [];
        this.showSearchMenu = false;
        return;
      }

      this.searchTimeout = setTimeout(() => {
        this.performPlaceSearch(query);
      }, 300); // 300ms delay after user stops typing
    },
    async performPlaceSearch(query) {
      this.searchLoading = true;
      try {
        const response = await this.$api.get("places/search", {
          params: {
            q: query,
            count: 10,
            locale: this.$config.getLanguageLocale() || "en",
          },
        });

        if (response.data && response.data.results) {
          this.searchResults = response.data.results;
          this.showSearchMenu = this.searchResults.length > 0;
        } else {
          this.searchResults = [];
          this.showSearchMenu = false;
        }
      } catch (error) {
        console.error("Place search error:", error);
        this.searchResults = [];
      } finally {
        this.searchLoading = false;
      }
    },
    onPlaceSelected(place) {
      if (place && place.lat && place.lng) {
        this.currentLat = place.lat;
        this.currentLng = place.lng;
        this.updatePosition(place.lat, place.lng);
        this.fetchLocationInfo(place.lat, place.lng);

        // Clear search after selection
        this.showSearchMenu = false;
        this.searchQuery = "";
      }
    },
    clearSearch() {
      this.searchQuery = "";
      this.searchResults = [];
      this.showSearchMenu = false;
    },
    onSearchFocus() {
      if (this.searchResults.length > 0) {
        this.showSearchMenu = true;
      }
    },
    onSearchBlur() {
      // Delay hiding menu to allow for selection
      setTimeout(() => {
        this.showSearchMenu = false;
      }, 200);
    },
  },
};
</script>
