<template>
  <v-dialog
    v-model="show"
    :max-width="900"
    :fullscreen="$vuetify.display.mdAndDown"
    :persistent="false"
    class="p-photo-map-dialog"
    @keydown.esc="close"
    @after-leave="onDialogClosed"
  >
    <v-card :tile="$vuetify.display.mdAndDown">
      <v-toolbar flat color="navigation" :density="$vuetify.display.smAndDown ? 'compact' : 'default'" class="px-4">
        <v-btn v-if="$vuetify.display.mdAndDown" icon @click.stop="close">
          <v-icon>mdi-close</v-icon>
        </v-btn>
        <v-icon v-else color="primary" class="mr-3">mdi-map-marker</v-icon>
        <v-toolbar-title>
          {{ $gettext("Set Location") }}
        </v-toolbar-title>
        <v-spacer></v-spacer>
        <v-btn v-if="!$vuetify.display.mdAndDown" icon class="action-close" @click.stop="close">
          <v-icon>mdi-close</v-icon>
        </v-btn>
      </v-toolbar>
      <v-card-text class="pa-0">
        <div class="d-flex flex-column flex-md-row py-4 px-2">
          <div class="flex-grow-1 position-relative mb-4 mb-md-0">
            <div ref="map" class="p-map" style="height: 50vh; min-height: 300px; width: 100%; border-radius: 4px"></div>
          </div>

          <div
            class="map-sidebar ml-0 ml-md-4"
            :style="{
              width: $vuetify.display.smAndDown ? '100%' : '300px',
              maxWidth: $vuetify.display.smAndDown ? '100%' : '300px',
              minWidth: 0,
            }"
          >
            <v-card border class="pa-3 mb-3">
              <div class="text-subtitle-2 mb-2">{{ $gettext("Search Places") }}</div>
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
                    :label="$gettext('Search for a place')"
                    prepend-inner-icon="mdi-magnify"
                    :append-inner-icon="searchLoading ? 'mdi-loading mdi-spin' : searchQuery ? 'mdi-delete' : ''"
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
            </v-card>

            <v-card border class="pa-3 mb-3">
              <div class="text-subtitle-2 mb-2">{{ $gettext("Coordinates") }}</div>
              <v-text-field
                v-model="coordinateInput"
                :label="$gettext('Latitude, Longitude')"
                prepend-inner-icon="mdi-map-marker"
                :append-inner-icon="locationWasCleared ? 'mdi-undo' : coordinateInput ? 'mdi-delete' : ''"
                density="compact"
                variant="outlined"
                placeholder="e.g., 52.5208, 13.4049"
                persistent-hint
                @keydown.enter="applyCoordinates"
                @update:model-value="onCoordinateInputChange"
                @click:append-inner="locationWasCleared ? undoClearLocation() : clearLocation()"
              ></v-text-field>
            </v-card>

            <v-card v-if="locationInfo" border class="pa-3 mb-3">
              <div class="text-subtitle-2 mb-2">{{ $gettext("Location Details") }}</div>
              <div class="text-body-2">
                {{ simplifiedLocationDisplay }}
              </div>
            </v-card>

            <v-card border class="pa-3">
              <div class="text-subtitle-2 mb-2">{{ $gettext("Instructions") }}</div>
              <div class="text-body-2">
                {{ $gettext("Click on the map to set a location. Drag the marker for precise positioning.") }}
              </div>
              <div class="mt-3">
                <div class="d-flex flex-wrap ga-2">
                  <v-btn
                    variant="flat"
                    color="button"
                    class="action-cancel flex-grow-1"
                    style="min-width: 120px"
                    @click.stop="close"
                  >
                    {{ $gettext("Cancel") }}
                  </v-btn>
                  <v-btn
                    color="primary"
                    class="flex-grow-1"
                    style="min-width: 120px"
                    :disabled="!(currentLat !== null && currentLng !== null)"
                    @click="confirm"
                  >
                    {{ $gettext("Apply") }}
                  </v-btn>
                </div>
              </div>
            </v-card>
          </div>
        </div>
      </v-card-text>
    </v-card>
  </v-dialog>
</template>

<script>
import maplibregl from "common/maplibregl";

export default {
  name: "PPhotoEditMapDialog",
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
      coordinateInput: "",
      locationInfo: null,
      locationWasCleared: false,
      latBeforeClear: null,
      lngBeforeClear: null,
      coordinateInputTimeout: null,
      searchQuery: "",
      searchResults: [],
      searchLoading: false,
      searchTimeout: null,
      showSearchMenu: false,
    };
  },
  computed: {
    isValidCoordinateInput() {
      if (!this.coordinateInput) return false;

      const parts = this.coordinateInput.split(",").map((part) => part.trim());
      if (parts.length !== 2) return false;

      const lat = parseFloat(parts[0]);
      const lng = parseFloat(parts[1]);

      return !isNaN(lat) && !isNaN(lng) && lat >= -90 && lat <= 90 && lng >= -180 && lng <= 180;
    },
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
        this.locationWasCleared = false;
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
    currentLat() {
      this.updateCoordinateInput();
    },
    currentLng() {
      this.updateCoordinateInput();
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
      this.coordinateInput = "";
      this.locationWasCleared = false;
      this.latBeforeClear = null;
      this.lngBeforeClear = null;

      // Clear any pending timeout
      if (this.coordinateInputTimeout) {
        clearTimeout(this.coordinateInputTimeout);
        this.coordinateInputTimeout = null;
      }

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
          this.updateCoordinateInput();
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
          this.locationWasCleared = false;
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
            this.locationWasCleared = false;
          });
        }
      }
    },
    formatCoordinates(lat, lng) {
      return `${lat.toFixed(6)}, ${lng.toFixed(6)}`;
    },
    updateCoordinateInput() {
      if (this.currentLat && this.currentLng && !(this.currentLat === 0 && this.currentLng === 0)) {
        this.coordinateInput = this.formatCoordinates(this.currentLat, this.currentLng);
      } else {
        this.coordinateInput = "";
      }
    },
    applyCoordinates() {
      if (!this.isValidCoordinateInput) return;
      const parts = this.coordinateInput.split(",").map((part) => part.trim());
      const lat = parseFloat(parts[0]);
      const lng = parseFloat(parts[1]);
      this.currentLat = lat;
      this.currentLng = lng;
      this.updatePosition(lat, lng);
      this.fetchLocationInfo(lat, lng);
      this.locationWasCleared = false;
    },
    fetchLocationInfo(lat, lng) {
      this.$api
        .get(`maps/geocode/reverse?lat=${lat}&lng=${lng}`)
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
    clearLocation() {
      this.latBeforeClear = this.currentLat;
      this.lngBeforeClear = this.currentLng;
      this.currentLat = 0;
      this.currentLng = 0;
      this.coordinateInput = "";
      if (this.marker) {
        this.marker.remove();
        this.marker = null;
      }
      this.locationInfo = null;
      if (this.map) {
        this.map.flyTo({
          center: [0, 20],
          zoom: 2,
          essential: true,
        });
      }
      this.locationWasCleared = true;
    },
    undoClearLocation() {
      if (this.latBeforeClear !== null && this.lngBeforeClear !== null) {
        this.currentLat = this.latBeforeClear;
        this.currentLng = this.lngBeforeClear;
        this.updatePosition(this.currentLat, this.currentLng);
        this.fetchLocationInfo(this.currentLat, this.currentLng);
      }
      this.locationWasCleared = false;
      this.latBeforeClear = null;
      this.lngBeforeClear = null;
    },
    onCoordinateInputChange() {
      this.locationWasCleared = false;

      if (this.coordinateInputTimeout) {
        clearTimeout(this.coordinateInputTimeout);
      }

      this.coordinateInputTimeout = setTimeout(() => {
        if (this.isValidCoordinateInput) {
          this.applyCoordinates();
        }
      }, 1000); // 1 second delay after user stops typing
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
        const response = await this.$api.get("maps/places/search", {
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
        this.locationWasCleared = false;

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
