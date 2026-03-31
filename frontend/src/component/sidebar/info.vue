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
          <div v-tooltip="$gettext('Caption')" class="text-body-2 meta-caption meta-scrollable" v-html="captionHtml"></div>
        </v-list-item>
        <v-divider v-if="model.Title || model.Caption" class="my-4"></v-divider>
        <v-list-item v-tooltip="$gettext(`Taken`)" :title="formatTime(model)" prepend-icon="mdi-calendar" class="metadata__item"> </v-list-item>

        <v-list-item v-tooltip="$gettext(`Size`)" :title="model.getTypeInfo()" :prepend-icon="model.getTypeIcon()" class="metadata__item"> </v-list-item>

        <v-list-item v-if="cameraInfo" v-tooltip="$gettext('Camera')" :title="cameraInfo" prepend-icon="mdi-camera" class="metadata__item"> </v-list-item>

        <v-list-item v-if="lensInfo" v-tooltip="$gettext('Lens')" :title="lensInfo" prepend-icon="mdi-camera-iris" class="metadata__item"> </v-list-item>

        <v-list-item v-if="exifInfo" v-tooltip="$gettext('Exposure')" :title="exifInfo" prepend-icon="mdi-tune" class="metadata__item"> </v-list-item>

        <template v-if="model.Lat && model.Lng">
          <v-divider class="my-4"></v-divider>
          <v-list-item v-if="placeName" v-tooltip="$gettext('Location')" :title="placeName" prepend-icon="mdi-map-marker" class="metadata__item"> </v-list-item>
          <v-list-item
            v-tooltip="$gettext(`Coordinates`)"
            prepend-icon="mdi-crosshairs-gps"
            :title="model.getLatLng()"
            class="clickable metadata__item"
            @click.stop="model.copyLatLng()"
          >
          </v-list-item>
          <v-list-item v-if="featPlaces" class="mx-0 px-0">
            <p-map :latlng="[model.Lat, model.Lng]" :animate-duration="0"></p-map>
          </v-list-item>
        </template>

        <template v-if="people.length > 0">
          <v-divider class="my-4"></v-divider>
          <v-list-item class="metadata__item">
            <div class="text-subtitle-2">{{ $gettext("People") }}</div>
          </v-list-item>
          <v-list-item v-for="m in people" :key="m.UID || m.CropID" class="metadata__item metadata__person-row" :class="{ clickable: m.Name }" @click.stop.prevent="m.Name ? navigateToPerson(m) : null">
            <template #prepend>
              <img :src="m.thumbnailUrl('tile_160')" :alt="m.Name" class="meta-person__avatar" />
            </template>
            <v-list-item-title v-if="m.Name" class="meta-person__name">{{ m.Name }}</v-list-item-title>
          </v-list-item>
        </template>

        <template v-if="labels.length > 0">
          <v-divider class="my-4"></v-divider>
          <v-list-item class="metadata__item">
            <div class="text-subtitle-2">{{ $gettext("Labels") }}</div>
          </v-list-item>
          <v-list-item class="metadata__item metadata__chips">
            <div class="d-flex flex-wrap ga-1">
              <span v-for="l in labels" :key="l.Label.UID" class="meta-chip meta-chip--primary" @click.stop.prevent="navigateToLabel(l.Label)">
                {{ l.Label.Name }}
              </span>
            </div>
          </v-list-item>
        </template>

        <template v-if="albums.length > 0">
          <v-divider class="my-4"></v-divider>
          <v-list-item class="metadata__item">
            <div class="text-subtitle-2">{{ $gettext("Albums") }}</div>
          </v-list-item>
          <v-list-item class="metadata__item metadata__chips">
            <div class="d-flex flex-wrap ga-1">
              <span v-for="a in albums" :key="a.UID" class="meta-chip meta-chip--primary" @click.stop.prevent="navigateToAlbum(a)">
                {{ a.Title }}
              </span>
            </div>
          </v-list-item>
        </template>

        <template v-if="keywords">
          <v-divider class="my-4"></v-divider>
          <v-list-item class="metadata__item">
            <div class="text-subtitle-2">{{ $gettext("Keywords") }}</div>
          </v-list-item>
          <v-list-item class="metadata__item">
            <div class="text-body-2 meta-keywords">{{ keywords }}</div>
          </v-list-item>
        </template>

        <template v-if="subject || artist || copyright || license">
          <v-divider class="my-4"></v-divider>
          <v-list-item v-if="subject" v-tooltip="$gettext('Subject')" :title="subject" prepend-icon="mdi-text-short" class="metadata__item"> </v-list-item>
          <v-list-item v-if="artist" v-tooltip="$gettext('Artist')" :title="artist" prepend-icon="mdi-account-edit" class="metadata__item"> </v-list-item>
          <v-list-item v-if="copyright" v-tooltip="$gettext('Copyright')" :title="copyright" prepend-icon="mdi-copyright" class="metadata__item"> </v-list-item>
          <v-list-item v-if="license" v-tooltip="$gettext('License')" :title="license" prepend-icon="mdi-license" class="metadata__item"> </v-list-item>
        </template>

        <template v-if="fileName">
          <v-divider class="my-4"></v-divider>
          <v-list-item v-tooltip="$gettext('File')" :title="fileName" prepend-icon="mdi-file" class="metadata__item"> </v-list-item>
          <v-list-item
            v-if="originalName && originalName !== fileName"
            v-tooltip="$gettext('Original Name')"
            :title="originalName"
            prepend-icon="mdi-file-outline"
            class="metadata__item"
          >
          </v-list-item>
        </template>

        <template v-if="notesHtml">
          <v-divider class="my-4"></v-divider>
          <v-list-item class="metadata__item">
            <div class="text-subtitle-2">{{ $gettext("Notes") }}</div>
          </v-list-item>
          <v-list-item class="metadata__item">
            <div class="text-body-2 meta-notes meta-scrollable" v-html="notesHtml"></div>
          </v-list-item>
        </template>
      </v-list>
    </div>
  </div>
</template>

<script>
import { DateTime } from "luxon";
import * as formats from "options/formats";

import PMap from "component/map.vue";

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
    photo: {
      type: Object,
      default: null,
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
  emits: ["update:modelValue", "close", "navigate"],
  data() {
    return {
      actions: [],
      featPlaces: this.$config.feature("places"),
    };
  },
  computed: {
    model() {
      return this.modelValue;
    },
    p() {
      return this.photo;
    },
    captionHtml() {
      if (!this.model?.Caption) return "";
      return this.$util.sanitizeHtml(this.$util.encodeHTML(this.model.Caption));
    },
    notesHtml() {
      if (!this.p?.Details?.Notes) return "";
      return this.$util.sanitizeHtml(this.$util.encodeHTML(this.p.Details.Notes));
    },
    cameraInfo() {
      if (!this.p) return "";
      const info = this.p.getCameraInfo();
      return info !== this.$gettext("Unknown") ? info : "";
    },
    lensInfo() {
      if (!this.p) return "";
      const info = this.p.getLensInfo();
      return info !== this.$gettext("Unknown") ? info : "";
    },
    exifInfo() {
      if (!this.p) return "";
      return this.p.getExifInfo();
    },
    people() {
      if (!this.p) return [];
      return this.p.getMarkers(true);
    },
    labels() {
      if (!this.p?.Labels) return [];
      return this.p.Labels.filter((l) => l.Label && l.Label.Name);
    },
    albums() {
      if (!this.p?.Albums) return [];
      return this.p.Albums.filter((a) => a.Title);
    },
    subject() {
      return this.p?.Details?.Subject || "";
    },
    artist() {
      return this.p?.Details?.Artist || "";
    },
    copyright() {
      return this.p?.Details?.Copyright || "";
    },
    license() {
      return this.p?.Details?.License || "";
    },
    keywords() {
      return this.p?.Details?.Keywords || "";
    },
    placeName() {
      if (!this.p) return "";
      return this.p.PlaceLabel || "";
    },
    fileName() {
      if (!this.p) return "";
      return this.p.FileName || "";
    },
    originalName() {
      if (!this.p) return "";
      return this.p.OriginalName || "";
    },
  },
  methods: {
    close() {
      this.$emit("close");
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
    navigateToLabel(label) {
      const slug = label.CustomSlug || label.Slug;
      this.$emit("navigate", { name: "browse", query: { q: "label:" + slug } });
    },
    navigateToAlbum(album) {
      this.$emit("navigate", { name: "album", params: { album: album.UID, slug: "view" } });
    },
    navigateToPerson(marker) {
      if (marker.SubjUID) {
        this.$emit("navigate", { name: "browse", query: { q: "subject:" + marker.SubjUID } });
      } else if (marker.Name) {
        this.$emit("navigate", { name: "browse", query: { q: "person:" + marker.Name } });
      }
    },
  },
};
</script>
