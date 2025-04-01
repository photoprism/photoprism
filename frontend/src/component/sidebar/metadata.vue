<template>
  <div class="p-sidebar-metadata metadata">
    <v-toolbar density="comfortable" color="navigation">
      <v-btn :icon="$isRtl ? 'mdi-chevron-left' : 'mdi-chevron-right'" @click.stop="close()"></v-btn>
      <v-toolbar-title class="text-h6 ms-2">{{ $gettext("Info") }}</v-toolbar-title>
    </v-toolbar>
    <div v-if="modelValue.UID">
      <v-list nav slim tile density="compact" class="metadata__list mt-2">
        <v-list-item v-if="modelValue.Title" class="metadata__item">
          <div class="text-subtitle-1 font-weight-bold">{{ modelValue.Title }}</div>
          <!-- v-text-field
        :model-value="modelValue.Title"
        :placeholder="$gettext('Add a title')"
        density="comfortable"
        variant="solo-filled"
        hide-details
        class="pa-0 font-weight-bold"
      ></v-text-field -->
        </v-list-item>
        <v-list-item v-if="modelValue.Caption" class="metadata__item">
          <div class="text-body-2">{{ modelValue.Caption }}</div>
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
        <v-divider v-if="modelValue.Title || modelValue.Caption" class="my-4"></v-divider>
        <v-list-item
          prepend-icon="mdi-calendar"
          :title="$util.formatDate(modelValue.TakenAtLocal, 'date_med_tz', modelValue.TimeZone)"
          class="metadata__item"
        >
          <template #append>
            <v-icon icon="mdi-pencil" size="20"></v-icon>
          </template>
        </v-list-item>

        <v-list-item
          v-if="modelValue.Type === 'image'"
          prepend-icon="mdi-image"
          :title="`${((modelValue.Width * modelValue.Height) / 1000000).toFixed(1)}MP ${modelValue.Width}×${modelValue.Height}`"
          class="metadata__item"
        >
        </v-list-item>
        <v-list-item
          v-else-if="modelValue.Type === 'raw'"
          prepend-icon="mdi-camera"
          :title="
            $gettext('RAW') +
            ` ${((modelValue.Width * modelValue.Height) / 1000000).toFixed(1)}MP ${modelValue.Width}×${modelValue.Height}`
          "
          class="metadata__item"
        >
        </v-list-item>
        <v-list-item
          v-else-if="modelValue.Type === 'live'"
          prepend-icon="mdi-play-circle-outline"
          :title="
            $gettext('Live') +
            ` ${((modelValue.Width * modelValue.Height) / 1000000).toFixed(1)}MP ${modelValue.Width}×${modelValue.Height}`
          "
          class="metadata__item"
        >
        </v-list-item>
        <v-list-item
          v-else-if="modelValue.Type === 'document'"
          prepend-icon="mdi-file-pdf-box"
          :title="$gettext('Document')"
          class="metadata__item"
        >
        </v-list-item>

        <v-divider class="my-4"></v-divider>

        <v-list-item prepend-icon="mdi-map-marker" title="Add location" class="metadata__item"> </v-list-item>
      </v-list>
    </div>
  </div>
</template>
<script>
import model from "../../model/model";

export default {
  name: "PSidebarMetadata",
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
    };
  },
  computed: {
    model() {
      return model;
    },
  },
  methods: {
    close() {
      this.$emit("close");
      console.log("CLOSE");
    },
  },
};
</script>
