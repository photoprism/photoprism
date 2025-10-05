<template>
  <v-autocomplete
    v-model="selectedLocation"
    :items="locationSuggestions"
    :loading="searchLoading"
    :search="locationInput"
    :disabled="disabled"
    :hide-details="hideDetails"
    :label="label"
    :placeholder="placeholder"
    :density="density"
    :validate-on="validateOn"
    :rules="[() => !locationInput || isValidInput]"
    autocomplete="off"
    autocorrect="off"
    autocapitalize="none"
    class="input-location"
    item-title="displayName"
    item-value="coordinates"
    return-object
    clearable
    no-filter
    menu-icon=""
    :menu-props="{ maxHeight: 300 }"
    @update:search="onLocationInputChange"
    @update:model-value="onLocationSelected"
    @click:clear="clearLocation"
    @keydown.enter="applyLocation"
  >
    <template v-if="icon" #prepend-inner>
      <v-icon
        v-if="showMapButton"
        variant="plain"
        :icon="icon"
        :title="mapButtonTitle"
        :disabled="mapButtonDisabled"
        class="action-map"
        @click.stop="$emit('open-map')"
      >
      </v-icon>
      <v-icon v-else variant="plain" :icon="icon" class="text-disabled"> </v-icon>
    </template>
    <template #append-inner>
      <v-icon
        v-if="showUndoButton"
        variant="plain"
        icon="mdi-undo"
        class="action-undo"
        @click.stop="undoClear"
      ></v-icon>
      <v-icon
        v-else-if="locationInput"
        variant="plain"
        icon="mdi-close-circle"
        class="action-clear"
        @click.stop="clearLocation"
      ></v-icon>
    </template>
    <template #item="{ props, item }">
      <v-list-item v-bind="props" density="compact">
        <template #prepend>
          <v-icon>mdi-map-marker</v-icon>
        </template>
        <template #title>
          <div class="d-flex flex-column">
            <span class="text-body-2">{{ item.raw.name }}</span>
            <span v-if="item.raw.coordinates" class="text-caption text-medium-emphasis">
              {{ item.raw.coordinates[0].toFixed(6) }}, {{ item.raw.coordinates[1].toFixed(6) }}
            </span>
          </div>
        </template>
      </v-list-item>
    </template>
    <template #no-data>
      <v-list-item v-if="locationInput && locationInput.length >= 2 && !searchLoading && locationSuggestions.length === 0">
        <v-list-item-title>{{ $gettext("No locations found") }}</v-list-item-title>
      </v-list-item>
    </template>
  </v-autocomplete>
</template>

<script>
export default {
  name: "PLocationInput",
  props: {
    latlng: {
      type: Array,
      default: () => [null, null],
      validator: (value) => Array.isArray(value) && value.length === 2,
    },
    disabled: {
      type: Boolean,
      default: false,
    },
    hideDetails: {
      type: Boolean,
      default: true,
    },
    label: {
      type: String,
      default: "",
    },
    placeholder: {
      type: String,
      default: "Enter location name or coordinates (e.g., San Francisco or 37.7749, -122.4194)",
    },
    density: {
      type: String,
      default: "comfortable",
    },
    validateOn: {
      type: String,
      default: "input",
    },
    showMapButton: {
      type: Boolean,
      default: false,
    },
    icon: {
      type: String,
      default: "mdi-map-marker",
    },
    mapButtonTitle: {
      type: String,
      default: "",
    },
    mapButtonDisabled: {
      type: Boolean,
      default: false,
    },
    enableUndo: {
      type: Boolean,
      default: false,
    },
    autoApply: {
      type: Boolean,
      default: true,
    },
    debounceDelay: {
      type: Number,
      default: 1000,
    },
  },
  emits: ["update:latlng", "changed", "cleared", "open-map"],
  data() {
    return {
      locationInput: "",
      selectedLocation: null,
      locationSuggestions: [],
      searchLoading: false,
      searchTimeout: null,
      inputTimeout: null,
      wasCleared: false,
      lastValidLat: null,
      lastValidLng: null,
    };
  },
  computed: {
    isValidInput() {
      if (!this.locationInput) return false;

      // Check if it's valid coordinates
      const parts = this.locationInput.split(",").map((part) => part.trim());
      if (parts.length === 2) {
        const lat = parseFloat(parts[0]);
        const lng = parseFloat(parts[1]);
        return !isNaN(lat) && !isNaN(lng) && lat >= -90 && lat <= 90 && lng >= -180 && lng <= 180;
      }

      // Check if it's a valid location name (at least 2 characters)
      return this.locationInput.length >= 2;
    },
    showUndoButton() {
      return this.enableUndo && this.wasCleared && this.lastValidLat !== null && this.lastValidLng !== null;
    },
  },
  watch: {
    latlng() {
      this.updateLocationInput();
    },
  },
  mounted() {
    this.updateLocationInput();
  },
  beforeUnmount() {
    if (this.inputTimeout) {
      clearTimeout(this.inputTimeout);
    }
    if (this.searchTimeout) {
      clearTimeout(this.searchTimeout);
    }
  },
  methods: {
    updateLocationInput() {
      const lat = this.latlng[0];
      const lng = this.latlng[1];

      if (lat !== null && lng !== null && !(lat === 0 && lng === 0) && !isNaN(lat) && !isNaN(lng)) {
        this.locationInput = `${parseFloat(lat)}, ${parseFloat(lng)}`;
        this.wasCleared = false;
      } else {
        this.locationInput = "";
      }
    },

    onLocationInputChange(value) {
      this.locationInput = value;
      this.wasCleared = false;
      this.selectedLocation = null;

      if (this.inputTimeout) {
        clearTimeout(this.inputTimeout);
      }

      if (this.autoApply) {
        this.inputTimeout = setTimeout(() => {
          if (this.isValidInput) {
            this.applyLocation();
          }
        }, this.debounceDelay);
      }

      // Search for locations if input looks like a location name (not coordinates)
      if (value && value.length >= 2 && !this.isCoordinateInput(value)) {
        this.searchLocations(value);
      } else {
        this.clearSearchTimeout();
        this.locationSuggestions = [];
      }
    },

    isCoordinateInput(input) {
      const parts = input.split(",").map((part) => part.trim());
      if (parts.length !== 2) return false;
      
      const lat = parseFloat(parts[0]);
      const lng = parseFloat(parts[1]);
      return !isNaN(lat) && !isNaN(lng) && lat >= -90 && lat <= 90 && lng >= -180 && lng <= 180;
    },

    async searchLocations(query) {
      this.clearSearchTimeout();
      
      this.searchTimeout = setTimeout(async () => {
        if (!query || query.length < 2 || this.isCoordinateInput(query)) {
          return;
        }

        this.searchLoading = true;
        
        try {
          const response = await this.$api.get("places/search", {
            params: {
              q: query,
              count: 10,
              locale: this.$config.getLanguageLocale() || "en",
            },
          });

          if (response.data && Array.isArray(response.data)) {
            this.locationSuggestions = response.data.map(place => ({
              name: place.name,
              coordinates: [place.lat, place.lng],
              displayName: place.name,
            }));
          } else {
            this.locationSuggestions = [];
          }
        } catch (error) {
          console.error("Location search error:", error);
          this.locationSuggestions = [];
        } finally {
          this.searchLoading = false;
        }
      }, 300);
    },

    onLocationSelected(location) {
      if (location && location.coordinates) {
        this.selectedLocation = location;
        this.applyCoordinates(location.coordinates[0], location.coordinates[1]);
      }
    },

    applyLocation() {
      if (!this.isValidInput) return;

      // If it's coordinates input
      if (this.isCoordinateInput(this.locationInput)) {
        const parts = this.locationInput.split(",").map((part) => part.trim());
        const lat = parseFloat(parts[0]);
        const lng = parseFloat(parts[1]);
        this.applyCoordinates(lat, lng);
      }
      // If it's a selected location
      else if (this.selectedLocation && this.selectedLocation.coordinates) {
        this.applyCoordinates(this.selectedLocation.coordinates[0], this.selectedLocation.coordinates[1]);
      }
    },

    applyCoordinates(lat, lng) {
      this.$emit("update:latlng", [lat, lng]);
      this.$emit("changed", { lat: lat, lng: lng });
    },

    clearLocation() {
      if (this.enableUndo) {
        this.lastValidLat = this.latlng[0];
        this.lastValidLng = this.latlng[1];
      }

      this.locationInput = "";
      this.selectedLocation = null;
      this.locationSuggestions = [];
      this.wasCleared = true;

      this.$emit("update:latlng", [0, 0]);
      this.$emit("changed", { lat: 0, lng: 0 });
      this.$emit("cleared", {
        lat: 0,
        lng: 0,
        previousLatitude: this.lastValidLat,
        previousLongitude: this.lastValidLng,
      });
    },

    clearSearchTimeout() {
      if (this.searchTimeout) {
        clearTimeout(this.searchTimeout);
        this.searchTimeout = null;
      }
    },

    undoClear() {
      if (this.lastValidLat !== null && this.lastValidLng !== null) {
        this.$emit("update:latlng", [this.lastValidLat, this.lastValidLng]);
        this.$emit("changed", {
          lat: this.lastValidLat,
          lng: this.lastValidLng,
        });

        this.wasCleared = false;
        this.lastValidLat = null;
        this.lastValidLng = null;
      }
    },
    pastePosition(event) {
      // Autofill the lat and lng fields if the text in the clipboard contains two float values.
      const clipboard = event.clipboardData ? event.clipboardData : window.clipboardData;

      if (!clipboard) {
        return;
      }

      // Get values from browser clipboard.
      const text = clipboard.getData("text");

      // Trim spaces before splitting by whitespace and/or commas.
      const val = text.trim().split(/[ ,]+/);

      if (val.length >= 2) {
        const lat = parseFloat(val[0]);
        const lng = parseFloat(val[1]);

        if (!isNaN(lat) && lat >= -90 && lat <= 90 && !isNaN(lng) && lng >= -180 && lng <= 180) {
          // Update coordinates
          this.$emit("update:latlng", [lat, lng]);
          this.$emit("changed", { lat: lat, lng: lng });

          // Prevent default action.
          event.preventDefault();
        }
      }
    },
  },
};
</script>
