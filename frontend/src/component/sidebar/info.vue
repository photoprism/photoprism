<template>
  <div class="p-sidebar-info metadata">
    <v-toolbar density="comfortable" color="navigation">
      <v-btn
        :icon="$isRtl ? 'mdi-chevron-left' : 'mdi-chevron-right'"
        :title="$gettext('Close')"
        @click.stop="close()"
      ></v-btn>
      <v-toolbar-title>{{ $gettext("Information") }}</v-toolbar-title>
    </v-toolbar>
    <div v-if="model.UID">
      <v-list nav slim tile density="compact" class="metadata__list mt-2">
        <v-list-item v-if="model.Title" class="metadata__item">
          <div v-tooltip="$gettext('Title')" class="text-subtitle-2 meta-title">{{ model.Title }}</div>
          <!-- v-text-field
        :model-value="modelValue.Title"
        :placeholder="$gettext('Add a title')"
        density="comfortable"
        variant="solo-filled"
        hide-details
        class="pa-0 font-weight-bold"
      ></v-text-field -->
        </v-list-item>
        <v-list-item v-if="model.Caption" class="metadata__item">
          <div v-tooltip="$gettext('Caption')" class="text-body-2 meta-caption">{{ model.Caption }}</div>
          <!-- v-textarea
        :model-value="modelValue.Caption"
        :placeholder="$gettext('Add a caption')"
        density="comfortable"
        variant="solo-filled"
        hide-details
        autocomplete="off"
        auto-grow
        :rows="1"
        class="pa-0"
      ></v-textarea -->
        </v-list-item>
        <v-divider v-if="model.Title || model.Caption" class="my-4"></v-divider>
        <v-list-item
          v-tooltip="$gettext('Taken')"
          :title="formatTime(model)"
          prepend-icon="mdi-calendar"
          class="metadata__item"
        >
          <!-- template #append>
            <v-icon icon="mdi-pencil" size="20"></v-icon>
          </template -->
        </v-list-item>

        <v-list-item
          v-tooltip="$gettext('Size')"
          :title="model.getTypeInfo()"
          :prepend-icon="model.getTypeIcon()"
          class="metadata__item"
        >
        </v-list-item>

        <template v-if="model.Lat && model.Lng">
          <v-divider class="my-4"></v-divider>
          <v-list-item
            v-tooltip="$gettext('Location')"
            prepend-icon="mdi-map-marker"
            :title="model.getLatLng()"
            class="clickable metadata__item"
            @click.stop="model.copyLatLng()"
          >
          </v-list-item>
          <v-list-item v-if="featPlaces" class="mx-0 px-0">
            <p-map :lat="model.Lat" :lng="model.Lng"></p-map>
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
    album: {
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
    };
  },
  computed: {
    model() {
      return this.modelValue;
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
  },
};
</script>
