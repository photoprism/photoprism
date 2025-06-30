<template>
  <v-dialog
    v-model="show"
    :max-width="900"
    :fullscreen="$vuetify.display.xs"
    :persistent="false"
    class="p-position-dialog"
    @keydown.esc="close"
    @after-leave="onDialogClosed"
  >
    <v-card :tile="$vuetify.display.xs">
      <v-toolbar
        v-if="$vuetify.display.xs"
        flat
        color="navigation"
        :density="$vuetify.display.smAndDown ? 'compact' : 'default'"
      >
        <v-btn icon @click.stop="close">
          <v-icon>mdi-close</v-icon>
        </v-btn>
        <v-toolbar-title>
          {{ $gettext("Set Position") }}
        </v-toolbar-title>
      </v-toolbar>
      <v-card-title v-else class="d-flex justify-start align-center ga-3">
        <v-icon size="28" color="primary">mdi-map-marker</v-icon>
        <h6 class="text-h6">{{ $gettext("Set Position") }}</h6>
      </v-card-title>
      <v-card-text class="pb-3">
        <div class="d-flex flex-column flex-md-row ga-5">
          <div class="flex-grow-1 position-relative mb-4 mb-md-0">
            <div ref="map" class="p-map" style="height: 50vh; min-height: 300px; width: 100%; border-radius: 4px"></div>
          </div>

          <div
            class="map-sidebar d-flex flex-column"
            :class="$vuetify.display.xs ? `ga-3` : 'ga-5'"
            :style="{
              width: $vuetify.display.smAndDown ? '100%' : '300px',
              maxWidth: $vuetify.display.smAndDown ? '100%' : '300px',
              minWidth: 0,
            }"
          >
            <div>
              <v-autocomplete
                ref="search"
                v-model="selectedPlace"
                :items="searchResults"
                :loading="searchLoading"
                :search="searchQuery"
                prepend-inner-icon="mdi-magnify"
                density="compact"
                variant="outlined"
                :placeholder="$gettext(`Search`)"
                item-title="formatted"
                item-value="id"
                return-object
                auto-select-first
                clearable
                autocomplete="off"
                no-filter
                menu-icon=""
                :menu-props="{ maxHeight: 300 }"
                @update:search="onSearchQueryChange"
                @update:model-value="onPlaceSelected"
                @click:clear="clearSearch"
              >
                <template #item="{ props }">
                  <v-list-item v-bind="props" density="compact">
                    <template #prepend>
                      <v-icon>mdi-map-marker</v-icon>
                    </template>
                  </v-list-item>
                </template>
                <template #no-data>
                  <v-list-item
                    v-if="searchQuery && searchQuery.length >= 2 && !searchLoading && searchResults.length === 0"
                  >
                    <v-list-item-title>{{ $gettext("No results found") }}</v-list-item-title>
                  </v-list-item>
                </template>
              </v-autocomplete>
            </div>
            <!-- div v-if="locationInfo">
              <div class="text-subtitle-2 mb-2">{{ $gettext("Location Details") }}</div>
              <div class="text-body-2">
                {{ simplifiedLocationDisplay }}
              </div>
            </div -->

            <div class="text-body-2 mt-3">
              {{ $gettext("You can search for a location or move the marker on the map to change the position:") }}
            </div>

            <div class="flex-grow-1">
              <p-position-input
                :latitude="currentLat"
                :longitude="currentLng"
                density="comfortable"
                :placeholder="$gettext(`Position`)"
                :enable-undo="true"
                :auto-apply="true"
                :label="simplifiedLocationDisplay"
                @update:latitude="updateLatitude"
                @update:longitude="updateLongitude"
                @coordinates-changed="onCoordinatesChanged"
                @coordinates-cleared="onCoordinatesCleared"
              ></p-position-input>
            </div>

            <div class="action-buttons">
              <v-btn variant="flat" color="button" class="action-cancel" min-width="120" @click.stop="close">
                {{ $gettext("Cancel") }}
              </v-btn>
              <v-btn
                color="highlight"
                min-width="120"
                :disabled="!(currentLat !== null && currentLng !== null) || locationInfoLoading"
                :loading="locationInfoLoading"
                @click="confirm"
              >
                {{ $gettext("Apply") }}
              </v-btn>
            </div>
          </div>
        </div>
      </v-card-text>
    </v-card>
  </v-dialog>
</template>

<script>
import maplibregl from "common/maplibregl";
import PPositionInput from "component/position/input.vue";

export default {
  name: "PPositionDialog",
  components: {
    PPositionInput,
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
      locationInfoLoading: false,
      searchQuery: "",
      searchResults: [],
      searchLoading: false,
      searchTimeout: null,
      selectedPlace: null,
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
      this.locationInfoLoading = false;
      this.resetSearchState();
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
          this.setPositionAndFetchInfo(e.lngLat.lat, e.lngLat.lng);
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
          duration: 900,
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
            this.setPositionAndFetchInfo(lngLat.lat, lngLat.lng);
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
      this.locationInfoLoading = false;
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

    clearSearchTimeout() {
      if (this.searchTimeout) {
        clearTimeout(this.searchTimeout);
        this.searchTimeout = null;
      }
    },

    resetSearchState() {
      this.searchQuery = "";
      this.searchResults = [];
      this.selectedPlace = null;
      this.searchLoading = false;
      this.clearSearchTimeout();
    },

    setPositionAndFetchInfo(lat, lng) {
      this.currentLat = lat;
      this.currentLng = lng;
      this.updatePosition(lat, lng);
      this.fetchLocationInfo(lat, lng);
    },

    fetchLocationInfo(lat, lng) {
      this.locationInfoLoading = true;
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
        })
        .finally(() => {
          this.locationInfoLoading = false;
        });
    },

    onSearchQueryChange(query) {
      this.searchQuery = query;
      this.clearSearchTimeout();

      if (!query || query.length < 2) {
        this.searchResults = [];
        this.searchLoading = false;
        return;
      }

      this.searchLoading = true;
      this.searchTimeout = setTimeout(() => {
        this.performPlaceSearch(query);
      }, 300); // 300ms delay after user stops typing
    },
    async performPlaceSearch(query) {
      if (!query || query.length < 2) {
        this.searchLoading = false;
        return;
      }

      try {
        const response = await this.$api.get("places/search", {
          params: {
            q: query,
            count: 10,
            locale: this.$config.getLanguageLocale() || "en",
          },
        });

        if (this.searchQuery === query) {
          if (response.data && response.data.results) {
            this.searchResults = response.data.results;
          } else {
            this.searchResults = [];
          }
        }
      } catch (error) {
        console.error("Place search error:", error);
        if (this.searchQuery === query) {
          this.searchResults = [];
        }
      } finally {
        if (this.searchQuery === query) {
          this.searchLoading = false;
        }
      }
    },
    onPlaceSelected(place) {
      if (place && place.lat && place.lng) {
        this.setPositionAndFetchInfo(place.lat, place.lng);

        this.$nextTick(() => {
          this.resetSearchState();
        });
      }
    },
    clearSearch() {
      this.resetSearchState();
    },
  },
};
</script>
