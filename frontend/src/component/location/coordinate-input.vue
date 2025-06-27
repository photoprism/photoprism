<template>
  <v-text-field
    v-model="coordinateInput"
    :disabled="disabled"
    :hide-details="hideDetails"
    :label="label"
    :placeholder="placeholder"
    :density="density"
    :validate-on="validateOn"
    :rules="[() => !coordinateInput || isValidCoordinateInput]"
    autocomplete="off"
    autocorrect="off"
    autocapitalize="none"
    class="input-coordinates"
    @keydown.enter="applyCoordinates"
    @update:model-value="onCoordinateInputChange"
    @paste="pastePosition"
  >
    <template #prepend-inner>
      <v-icon
        v-if="showMapButton"
        variant="plain"
        icon="mdi-crosshairs-gps"
        :title="mapButtonTitle"
        :disabled="mapButtonDisabled"
        class="action-map"
        @click.stop="$emit('open-map')"
      >
      </v-icon>
      <v-icon v-else variant="plain" icon="mdi-crosshairs-gps" class="text-disabled"> </v-icon>
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
        v-else-if="coordinateInput"
        variant="plain"
        icon="mdi-close-circle"
        class="action-clear"
        @click.stop="clearCoordinates"
      ></v-icon>
    </template>
  </v-text-field>
</template>

<script>
export default {
  name: "PCoordinateInput",
  props: {
    latitude: {
      type: Number,
      default: null,
    },
    longitude: {
      type: Number,
      default: null,
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
      default: "e.g., 52.5208, 13.4049",
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
  emits: ["update:latitude", "update:longitude", "coordinates-changed", "coordinates-cleared", "open-map"],
  data() {
    return {
      coordinateInput: "",
      inputTimeout: null,
      wasCleared: false,
      lastValidLat: null,
      lastValidLng: null,
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
    showUndoButton() {
      return this.enableUndo && this.wasCleared && this.lastValidLat !== null && this.lastValidLng !== null;
    },
  },
  watch: {
    latitude() {
      this.updateCoordinateInput();
    },
    longitude() {
      this.updateCoordinateInput();
    },
  },
  mounted() {
    this.updateCoordinateInput();
  },
  beforeUnmount() {
    if (this.inputTimeout) {
      clearTimeout(this.inputTimeout);
    }
  },
  methods: {
    updateCoordinateInput() {
      const lat = this.latitude;
      const lng = this.longitude;

      if (lat !== null && lng !== null && !(lat === 0 && lng === 0) && !isNaN(lat) && !isNaN(lng)) {
        this.coordinateInput = `${parseFloat(lat)}, ${parseFloat(lng)}`;
        this.wasCleared = false;
      } else {
        this.coordinateInput = "";
      }
    },

    onCoordinateInputChange(value) {
      this.coordinateInput = value;
      this.wasCleared = false;

      if (this.inputTimeout) {
        clearTimeout(this.inputTimeout);
      }

      if (this.autoApply) {
        this.inputTimeout = setTimeout(() => {
          if (this.isValidCoordinateInput) {
            this.applyCoordinates();
          }
        }, this.debounceDelay);
      }
    },
    applyCoordinates() {
      if (!this.isValidCoordinateInput) return;

      const parts = this.coordinateInput.split(",").map((part) => part.trim());
      const lat = parseFloat(parts[0]);
      const lng = parseFloat(parts[1]);

      this.$emit("update:latitude", lat);
      this.$emit("update:longitude", lng);
      this.$emit("coordinates-changed", { latitude: lat, longitude: lng });
    },
    clearCoordinates() {
      if (this.enableUndo) {
        this.lastValidLat = this.latitude;
        this.lastValidLng = this.longitude;
      }

      this.coordinateInput = "";
      this.wasCleared = true;

      this.$emit("update:latitude", 0);
      this.$emit("update:longitude", 0);
      this.$emit("coordinates-changed", { latitude: 0, longitude: 0 });
      this.$emit("coordinates-cleared", {
        latitude: 0,
        longitude: 0,
        previousLatitude: this.lastValidLat,
        previousLongitude: this.lastValidLng,
      });
    },
    undoClear() {
      if (this.lastValidLat !== null && this.lastValidLng !== null) {
        this.$emit("update:latitude", this.lastValidLat);
        this.$emit("update:longitude", this.lastValidLng);
        this.$emit("coordinates-changed", {
          latitude: this.lastValidLat,
          longitude: this.lastValidLng,
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
          this.$emit("update:latitude", lat);
          this.$emit("update:longitude", lng);
          this.$emit("coordinates-changed", { latitude: lat, longitude: lng });

          // Prevent default action.
          event.preventDefault();
        }
      }
    },
  },
};
</script>
