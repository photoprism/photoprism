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

        <template v-if="model.Type === 'image'">
          <v-list-item
            prepend-icon="mdi-image"
            :title="`${megapixels}MP\u2003${model.Width}×${model.Height}`"
            class="metadata__item"
          >
          </v-list-item>
        </template>
        <template v-else-if="model.Type === 'raw'">
          <v-list-item
            prepend-icon="mdi-raw"
            :title="`${megapixels}MP\u2003${model.Width}×${model.Height}`"
            class="metadata__item"
          >
          </v-list-item>
        </template>
        <template v-else-if="model.Type === 'video'">
          <v-list-item
            prepend-icon="mdi-video"
            :title="`${megapixels}MP\u2003${model.Width}×${model.Height}${model.Duration ? '\u2003' + $util.formatDuration(model.Duration) : ''}${model.Codec ? '\u2003' + $util.formatCodec(model.Codec) : ''}`"
            class="metadata__item"
          >
          </v-list-item>
        </template>
        <template v-else-if="model.Type === 'live'">
          <v-list-item
            prepend-icon="mdi-play-circle-outline"
            :title="`${megapixels}MP\u2003${model.Width}×${model.Height}${model.Duration ? '\u2003' + $util.formatDuration(model.Duration) : ''}`"
            class="metadata__item"
          >
          </v-list-item>
        </template>
        <template v-else-if="model.Type === 'animated'">
          <v-list-item
            prepend-icon="mdi-file-gif-box"
            :title="`${megapixels}MP\u2003${model.Width}×${model.Height}`"
            class="metadata__item"
          >
          </v-list-item>
        </template>
        <template v-else-if="model.Type === 'vector'">
          <v-list-item
            prepend-icon="mdi-vector-polyline"
            :title="`${megapixels}MP\u2003${model.Width}×${model.Height}`"
            class="metadata__item"
          >
          </v-list-item>
        </template>
        <template v-else-if="model.Type === 'document'">
          <v-list-item
            prepend-icon="mdi-file-pdf-box"
            :title="$gettext('Document')"
            class="metadata__item"
          >
          </v-list-item>
        </template>

        <template v-if="model.Lat && model.Lng">
          <v-divider class="my-4"></v-divider>
          <!-- Clickable version commented out
          <v-list-item
            prepend-icon="mdi-map-marker"
            :title="`${model.Lat.toFixed(5)}°N ${model.Lng.toFixed(5)}°E`"
            class="clickable metadata__item"
            @click.stop="$util.copyText(`${model.Lat},${model.Lng}`)"
          >
          </v-list-item>
          -->
          <v-list-item
            prepend-icon="mdi-map-marker"
            :title="`${model.Lat.toFixed(5)}°N\u2003${model.Lng.toFixed(5)}°E`"
            class="metadata__item"
          >
          </v-list-item>
          <div id="metadata-map" ref="mapContainer" class="metadata__map"></div>
        </template>
      </v-list>
    </div>
  </div>
</template>

<script>
import model from "../../model/model";

let maplibregl;

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
      map: null,
      marker: null,
      mapLoaded: false,
      loadingMapLibre: false,
    };
  },
  computed: {
    model() {
      return this.modelValue;
    },
    megapixels() {
      if (!this.model.Width || !this.model.Height) return 0;
      return ((this.model.Width * this.model.Height) / 1000000).toFixed(1);
    },
  },
  watch: {
    "model.Lat"() {
      this.loadMapAndInit();
    },
    "model.Lng"() {
      this.loadMapAndInit();
    },
  },
  mounted() {
    if (this.model.Lat && this.model.Lng) {
      this.loadMapAndInit();
    }
  },
  beforeUnmount() {
    if (this.map) {
      this.map.remove();
    }
  },
  methods: {
    close() {
      this.$emit("close");
    },
    loadMapAndInit() {
      if (this.loadingMapLibre) {
        return;
      }

      this.loadingMapLibre = true;

      import("../../common/maplibregl.js")
        .then((module) => {
          maplibregl = module.default;
          // Wait for next tick to ensure DOM is ready
          this.$nextTick(() => {
            // Double check if the container exists
            if (this.$refs.mapContainer) {
              this.initMap();
            } else {
              // If container doesn't exist yet, wait a bit longer
              setTimeout(() => {
                this.initMap();
              }, 100);
            }
          });
        })
        .catch((error) => {
          console.error("Failed to load maplibregl:", error);
        })
        .finally(() => {
          this.loadingMapLibre = false;
        });
    },
    async initMap() {
      if (!this.model.Lat || !this.model.Lng || !this.$refs.mapContainer || !maplibregl) {
        return;
      }

      try {
        if (this.map) {
          this.map.remove();
        }

        const mapKey = this.$config.has("mapKey") ? this.$config.get("mapKey").replace(/[^a-z0-9]/gi, "") : "";
        const style = this.$config.values.settings.maps.style;
        let styleUrl = "https://cdn.photoprism.app/maps/default.json";

        if (mapKey && style) {
          styleUrl = `https://api.maptiler.com/maps/${style === "streets" ? "streets-v2" : style}/style.json?key=${mapKey}`;
        }

        this.map = new maplibregl.Map({
          container: this.$refs.mapContainer,
          style: styleUrl,
          center: [this.model.Lng, this.model.Lat],
          zoom: 13,
          interactive: true,
          attributionControl: false,
        });

        this.map.on("error", (e) => {
          console.error("Map error:", e);
        });

        this.map.on("load", () => {
          this.mapLoaded = true;

          if (this.marker) {
            this.marker.remove();
          }

          this.marker = new maplibregl.Marker().setLngLat([this.model.Lng, this.model.Lat]).addTo(this.map);
        });
      } catch (error) {
        console.error("Failed to initialize map:", error);
        this.mapLoaded = false;
      }
    },
  },
};
</script>

<style lang="scss">
.p-sidebar-info {
  display: flex;
  flex-direction: column;
  height: 100%;
  overflow: hidden;
}

.metadata__list {
  display: block;
  overflow-y: auto;
  height: 100%;
  padding: 0;
}

.metadata__item {
  margin-bottom: 4px;
  padding: 0 16px;
}

.metadata__map {
  display: block;
  height: 300px;
  margin: 8px 0 16px;
  background: var(--v-background-base, #f5f5f5);
  overflow: hidden;
}
</style>
