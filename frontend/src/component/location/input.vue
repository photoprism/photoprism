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
    @keydown.enter.stop="applyCoordinates"
    @update:model-value="onCoordinateInputChange"
    @paste="pastePosition"
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
      default: "37.75267, -122.543",
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
    latlng() {
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
      const lat = this.latlng[0];
      const lng = this.latlng[1];

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

      this.$emit("update:latlng", [lat, lng]);
      this.$emit("changed", { lat: lat, lng: lng });
    },
    clearCoordinates() {
      if (this.enableUndo) {
        this.lastValidLat = this.latlng[0];
        this.lastValidLng = this.latlng[1];
      }

      this.coordinateInput = "";
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
