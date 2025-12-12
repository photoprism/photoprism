<template>
  <div class="p-sidebar-info metadata">
    <v-toolbar density="comfortable" color="navigation">
      <v-btn :icon="$isRtl ? 'mdi-chevron-left' : 'mdi-chevron-right'" :title="$gettext('Close')" @click.stop="close()"></v-btn>
      <v-toolbar-title>{{ $gettext(`Information`) }}</v-toolbar-title>
    </v-toolbar>
    <div v-if="model.UID">
      <v-list nav slim tile density="compact" class="metadata__list mt-2">
        <v-list-item v-if="model.Title" class="metadata__item">
          <div v-tooltip="$pgettext(`Photo`, `Title`)" class="text-subtitle-2 meta-title">{{ model.Title }}</div>
        </v-list-item>
        <v-list-item v-if="model.Caption" class="metadata__item">
          <div v-tooltip="$gettext('Caption')" class="text-body-2 meta-caption">{{ model.Caption }}</div>
        </v-list-item>
        <v-divider v-if="model.Title || model.Caption" class="my-4"></v-divider>

        <!-- Date and Time -->
        <v-list-item v-tooltip="$gettext('Taken')" :title="formatTime(model)" prepend-icon="mdi-calendar" class="metadata__item"> </v-list-item>

        <!-- File Info -->
        <v-list-item v-if="typeInfo" v-tooltip="$gettext('Size')" :title="typeInfo" :prepend-icon="typeIcon" class="metadata__item"> </v-list-item>

        <!-- Camera Section (only shown when full photo data is loaded) -->
        <template v-if="hasCamera">
          <v-divider class="my-4"></v-divider>
          <v-list-item v-if="cameraName" v-tooltip="$gettext('Camera')" :title="cameraName" prepend-icon="mdi-camera" class="metadata__item"> </v-list-item>
          <v-list-item v-if="lensInfo" v-tooltip="$gettext('Lens')" :title="lensInfo" prepend-icon="mdi-camera-iris" class="metadata__item"> </v-list-item>
          <v-list-item v-if="exposureInfo" v-tooltip="$gettext('Exposure')" :title="exposureInfo" prepend-icon="mdi-tune" class="metadata__item"> </v-list-item>
        </template>

        <!-- Location Section -->
        <template v-if="model.Lat && model.Lng">
          <v-divider class="my-4"></v-divider>
          <v-list-item v-if="locationLabel" v-tooltip="$gettext('Location')" :title="locationLabel" prepend-icon="mdi-map-marker" class="metadata__item">
          </v-list-item>
          <v-list-item
            v-tooltip="$gettext('Coordinates')"
            prepend-icon="mdi-crosshairs-gps"
            :title="latLng"
            class="clickable metadata__item"
            @click.stop="copyLatLng"
          >
          </v-list-item>
          <v-list-item
            v-if="model.Altitude"
            v-tooltip="$gettext('Altitude')"
            :title="model.Altitude + ' m'"
            prepend-icon="mdi-elevation-rise"
            class="metadata__item"
          >
          </v-list-item>
          <v-list-item v-if="featPlaces" class="mx-0 px-0" style="margin-top: 0.5rem">
            <p-map :latlng="[model.Lat, model.Lng]"></p-map>
          </v-list-item>
        </template>

        <!-- Details Section (only shown when full photo data is loaded) -->
        <template v-if="hasDetails">
          <v-divider class="my-4"></v-divider>
          <v-list-item
            v-if="photo.Details && photo.Details.Artist"
            v-tooltip="$gettext('Artist')"
            :title="photo.Details.Artist"
            prepend-icon="mdi-account"
            class="metadata__item"
          >
          </v-list-item>
          <v-list-item
            v-if="photo.Details && photo.Details.Copyright"
            v-tooltip="$gettext('Copyright')"
            :title="photo.Details.Copyright"
            prepend-icon="mdi-copyright"
            class="metadata__item"
          >
          </v-list-item>
          <v-list-item
            v-if="photo.Details && photo.Details.License"
            v-tooltip="$gettext('License')"
            :title="photo.Details.License"
            prepend-icon="mdi-certificate"
            class="metadata__item"
          >
          </v-list-item>
          <v-list-item
            v-if="photo.Details && photo.Details.Subject"
            v-tooltip="$gettext('Subject')"
            :title="photo.Details.Subject"
            prepend-icon="mdi-tag"
            class="metadata__item"
          >
          </v-list-item>
          <v-list-item
            v-if="photo.Details && photo.Details.Keywords"
            v-tooltip="$gettext('Keywords')"
            :title="photo.Details.Keywords"
            prepend-icon="mdi-tag-multiple"
            class="metadata__item"
          >
          </v-list-item>
          <v-list-item
            v-if="photo.Details && photo.Details.Notes"
            v-tooltip="$gettext('Notes')"
            :title="photo.Details.Notes"
            prepend-icon="mdi-note-text"
            class="metadata__item"
          >
          </v-list-item>
        </template>
      </v-list>
    </div>
  </div>
</template>

<script>
import { DateTime } from "luxon";
import * as formats from "options/formats";
import $util from "common/util";

import PMap from "component/map.vue";
import Photo from "model/photo";

export default {
  name: "PSidebarInfo",
  components: {
    PMap,
  },
  props: {
    modelValue: {
      type: Object,
      default: () => {},
    },
    collection: {
      type: Object,
      default: () => {},
    },
    context: {
      type: String,
      default: "",
    },
  },
  emits: ["update:modelValue", "close"],
  data() {
    return {
      actions: [],
      featPlaces: this.$config.feature("places"),
      photo: null,
      loading: false,
    };
  },
  computed: {
    model() {
      return this.modelValue;
    },
    typeInfo() {
      if (typeof this.model.getTypeInfo === "function") {
        return this.model.getTypeInfo();
      }
      return "";
    },
    typeIcon() {
      if (typeof this.model.getTypeIcon === "function") {
        return this.model.getTypeIcon();
      }
      return "mdi-image";
    },
    latLng() {
      if (typeof this.model.getLatLng === "function") {
        return this.model.getLatLng();
      }
      if (this.model.Lat && this.model.Lng) {
        return `${this.model.Lat.toFixed(5)}°N\u2003${this.model.Lng.toFixed(5)}°E`;
      }
      return "";
    },
    hasCamera() {
      if (!this.photo) return false;
      return this.photo.CameraID > 1 || this.photo.CameraMake || this.photo.CameraModel || this.photo.Iso || this.photo.Exposure;
    },
    cameraName() {
      if (!this.photo) return "";
      if (this.photo.CameraMake && this.photo.CameraModel) {
        return `${this.photo.CameraMake} ${this.photo.CameraModel}`;
      } else if (this.photo.Camera && this.photo.Camera.Make && this.photo.Camera.Model) {
        return `${this.photo.Camera.Make} ${this.photo.Camera.Model}`;
      }
      return "";
    },
    lensInfo() {
      if (!this.photo) return "";
      const parts = [];
      if (this.photo.LensModel) {
        parts.push(this.photo.LensModel.replace("f/", "ƒ/"));
      } else if (this.photo.Lens && this.photo.Lens.Model) {
        parts.push(this.photo.Lens.Model.replace("f/", "ƒ/"));
      }
      if (this.photo.FocalLength && !parts.some((p) => p.includes("mm"))) {
        parts.push(`${this.photo.FocalLength}mm`);
      }
      if (this.photo.FNumber && !parts.some((p) => p.includes("ƒ/"))) {
        parts.push(`ƒ/${this.photo.FNumber}`);
      }
      return parts.join(", ");
    },
    exposureInfo() {
      if (!this.photo) return "";
      const parts = [];
      if (this.photo.Iso) {
        parts.push(`ISO ${this.photo.Iso}`);
      }
      if (this.photo.FNumber) {
        parts.push(`ƒ/${this.photo.FNumber}`);
      }
      if (this.photo.Exposure) {
        parts.push(this.photo.Exposure);
      }
      if (this.photo.FocalLength) {
        parts.push(`${this.photo.FocalLength}mm`);
      }
      return parts.join(", ");
    },
    locationLabel() {
      if (this.photo) {
        if (this.photo.PlaceLabel) {
          return this.photo.PlaceLabel;
        }
        if (this.photo.Place && this.photo.Place.Label) {
          return this.photo.Place.Label;
        }
      }
      return "";
    },
    hasDetails() {
      if (!this.photo) return false;
      const d = this.photo.Details;
      return d && (d.Artist || d.Copyright || d.License || d.Subject || d.Keywords || d.Notes);
    },
  },
  watch: {
    "model.UID": {
      immediate: true,
      handler(uid) {
        if (uid) {
          this.loadPhoto(uid);
        } else {
          this.photo = null;
        }
      },
    },
  },
  methods: {
    close() {
      this.$emit("close");
    },
    copyLatLng() {
      if (typeof this.model.copyLatLng === "function") {
        this.model.copyLatLng();
      } else if (this.model.Lat && this.model.Lng) {
        $util.copyText(`${this.model.Lat.toString()},${this.model.Lng.toString()}`);
      }
    },
    loadPhoto(uid) {
      if (!uid || this.loading) return;
      this.loading = true;
      new Photo()
        .find(uid)
        .then((p) => {
          this.photo = p;
          this.loading = false;
        })
        .catch(() => {
          this.photo = null;
          this.loading = false;
        });
    },
    formatTime(model) {
      if (!model || !model.TakenAtLocal) {
        return this.$gettext("Unknown");
      }

      // Always parse as UTC to avoid time shifts
      const dateTime = DateTime.fromISO(model.TakenAtLocal, { zone: "UTC" });

      if (model.TimeZone && model.TimeZone !== "Local" && model.TimeZone !== "UTC") {
        // We use the real timezone just for display, but don't shift the time (prevents double timezone offset as backend already applied it)
        return dateTime.setZone(model.TimeZone, { keepLocalTime: true }).toLocaleString(formats.DATETIME_MED_TZ);
      } else {
        return dateTime.toLocaleString(formats.DATETIME_MED);
      }
    },
  },
};
</script>
