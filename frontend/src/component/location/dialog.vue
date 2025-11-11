<template>
  <v-dialog
    ref="dialog"
    :model-value="visible"
    :max-width="900"
    :fullscreen="$vuetify.display.xs"
    persistent
    scrim
    scrollable
    class="p-location-dialog"
    @keydown.esc="close"
    @after-enter="afterEnter"
    @after-leave="afterLeave"
  >
    <v-card :tile="$vuetify.display.xs">
      <v-toolbar v-if="$vuetify.display.xs" flat color="navigation" class="mb-4" density="compact">
        <v-btn icon @click.stop="close">
          <v-icon>mdi-close</v-icon>
        </v-btn>
        <v-toolbar-title>
          {{ $gettext("Adjust Location") }}
        </v-toolbar-title>
      </v-toolbar>
      <v-card-title v-else class="d-flex justify-start align-center ga-3">
        <v-icon size="28" color="primary">mdi-map-marker</v-icon>
        <h6 class="text-h6">{{ $gettext("Adjust Location") }}</h6>
      </v-card-title>
      <v-card-text class="pb-3">
        <div class="d-flex flex-column flex-md-row ga-5">
          <div class="flex-grow-1 position-relative mb-4 mb-md-0">
            <p-map
              ref="map"
              :latlng="[currentLat, currentLng]"
              :zoom="12"
              :style="style"
              :interactive="true"
              :draggable="true"
              :show-controls="true"
              :clickable="true"
              @marker-moved="onMarkerMoved"
              @map-clicked="onMapClicked"
            />
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
                item-title="name"
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
                    <v-list-item-title>{{ $gettext("No results") }}</v-list-item-title>
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
              <p-location-input
                :latlng="[currentLat, currentLng]"
                density="comfortable"
                :enable-undo="true"
                :auto-apply="true"
                :label="locationLabel"
                @update:latlng="onLatLngUpdate"
                @changed="onLocationChanged"
                @cleared="onLocationCleared"
              ></p-location-input>
            </div>

            <div class="action-buttons">
              <v-btn variant="flat" color="button" class="action-cancel" min-width="120" @click.stop="close">
                {{ $gettext("Cancel") }}
              </v-btn>
              <v-btn
                color="highlight"
                min-width="120"
                class="action-confirm"
                :disabled="!(currentLat !== null && currentLng !== null) || locationLoading"
                :loading="locationLoading"
                @click="confirm"
              >
                {{ $gettext("Confirm") }}
              </v-btn>
            </div>
          </div>
        </div>
      </v-card-text>
    </v-card>
  </v-dialog>
</template>

<script>
import PLocationInput from "component/location/input.vue";
import PMap from "component/map.vue";

export default {
  name: "PLocationDialog",
  components: {
    PLocationInput,
    PMap,
  },
  props: {
    visible: {
      type: Boolean,
      default: false,
    },
    latlng: {
      type: Array,
      default: () => [0, 0],
    },
    style: {
      type: String,
      default: "embedded",
    },
  },
  emits: ["update:latlng", "close", "confirm"],
  data() {
    return {
      currentLat: this.latlng[0],
      currentLng: this.latlng[1],
      location: null,
      locationLoading: false,
      searchQuery: "",
      searchResults: [],
      searchLoading: false,
      searchTimeout: null,
      selectedPlace: null,
    };
  },
  computed: {
    locationLabel() {
      if (!this.location || !this.location?.place?.label) {
        return "";
      }

      return this.location.place.label;
    },
  },
  watch: {
    visible(show) {
      if (show) {
        this.currentLat = this.latlng[0];
        this.currentLng = this.latlng[1];
      }
    },
  },
  methods: {
    close() {
      this.$emit("close");
    },
    confirm() {
      if (this.currentLat !== null && this.currentLng !== null) {
        this.$emit("update:latlng", [this.currentLat, this.currentLng]);
        this.$emit("confirm", {
          lat: this.currentLat,
          lng: this.currentLng,
          location: this.location,
        });
      }
    },
    afterEnter() {
      if (this.currentLat && this.currentLng && !(this.currentLat === 0 && this.currentLng === 0)) {
        this.fetchLocationInfo(this.currentLat, this.currentLng);
      }
    },
    afterLeave() {
      this.location = null;
      this.locationLoading = false;
      this.resetSearchState();
    },
    onMarkerMoved(event) {
      this.setPositionAndFetchInfo(event.lat, event.lng);
    },
    onMapClicked(event) {
      this.setPositionAndFetchInfo(event.lat, event.lng);
    },
    onLocationChanged(data) {
      if (data.lat && data.lng && !(data.lat === 0 && data.lng === 0)) {
        this.fetchLocationInfo(data.lat, data.lng);
      }
    },
    onLatLngUpdate(latlng) {
      this.currentLat = latlng[0];
      this.currentLng = latlng[1];
    },
    onLocationCleared() {
      this.location = null;
      this.locationLoading = false;

      // Use the map component's removeMarker method
      if (this.$refs.map) {
        this.$refs.map.removeMarker();
        this.$refs.map.flyTo(20, 0, 2); // lat, lng, zoom
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
      this.fetchLocationInfo(lat, lng);
    },
    fetchLocationInfo(lat, lng) {
      this.locationLoading = true;
      this.$api
        .get(`places/reverse?lat=${lat}&lng=${lng}`)
        .then((response) => {
          if (response.data && response.data?.place?.label) {
            this.location = response.data;
          } else {
            this.location = null;
          }
        })
        .catch((error) => {
          console.error("Reverse geocoding error:", error);
          this.location = null;
        })
        .finally(() => {
          this.locationLoading = false;
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
          if (response.data && Array.isArray(response.data)) {
            this.searchResults = response.data;
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
