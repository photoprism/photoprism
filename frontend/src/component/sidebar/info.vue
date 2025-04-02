<template>
  <div class="p-sidebar-info metadata">
    <v-toolbar density="comfortable" color="navigation">
      <v-btn :icon="$isRtl ? 'mdi-chevron-left' : 'mdi-chevron-right'" @click.stop="close()"></v-btn>
      <v-toolbar-title class="text-h6 ms-2">{{ $gettext("Info") }}</v-toolbar-title>
    </v-toolbar>
    <div v-if="model.UID">
      <v-list nav slim tile density="compact" class="metadata__list mt-2">
        <v-list-item v-if="model.Title" class="metadata__item">
          <div class="text-subtitle-1 font-weight-bold">{{ model.Title }}</div>
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
          <div class="text-body-2">{{ model.Caption }}</div>
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
          prepend-icon="mdi-calendar"
          :title="$util.formatDate(model.TakenAtLocal, 'date_med_tz', model.TimeZone)"
          class="metadata__item"
        >
          <!-- template #append>
            <v-icon icon="mdi-pencil" size="20"></v-icon>
          </template -->
        </v-list-item>

        <v-list-item
          v-if="model.Type === 'image'"
          prepend-icon="mdi-image"
          :title="`${((model.Width * modelValue.Height) / 1000000).toFixed(1)}MP ${model.Width}×${model.Height}`"
          class="metadata__item"
        >
        </v-list-item>
        <v-list-item
          v-else-if="model.Type === 'raw'"
          prepend-icon="mdi-camera"
          :title="
            $gettext('RAW') + ` ${((model.Width * model.Height) / 1000000).toFixed(1)}MP ${model.Width}×${model.Height}`
          "
          class="metadata__item"
        >
        </v-list-item>
        <v-list-item
          v-else-if="modelValue.Type === 'live'"
          prepend-icon="mdi-play-circle-outline"
          :title="
            $gettext('Live') +
            ` ${((model.Width * model.Height) / 1000000).toFixed(1)}MP ${model.Width}×${model.Height}`
          "
          class="metadata__item"
        >
        </v-list-item>
        <v-list-item
          v-else-if="model.Type === 'document'"
          prepend-icon="mdi-file-pdf-box"
          :title="$gettext('Document')"
          class="metadata__item"
        >
        </v-list-item>

        <template v-if="model.Lat && model.Lng">
          <v-divider class="my-4"></v-divider>
          <v-list-item
            prepend-icon="mdi-map-marker"
            :title="`${model.Lat.toFixed(5)}°N ${model.Lng.toFixed(5)}°E`"
            class="clickable metadata__item"
            @click.stop="$util.copyText(`${model.Lat},${model.Lng}`)"
          >
          </v-list-item>
        </template>
      </v-list>
    </div>
  </div>
</template>
<script>
export default {
  name: "PSidebarInfo",
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
      return this.modelValue;
    },
  },
  methods: {
    close() {
      this.$emit("close");
    },
  },
};
</script>
