<template>
  <div class="p-sidebar-info metadata">
    <v-toolbar density="comfortable" color="navigation">
      <v-btn :icon="$isRtl ? 'mdi-chevron-left' : 'mdi-chevron-right'" @click.stop="close()"></v-btn>
      <v-toolbar-title>{{ $gettext("Details") }}</v-toolbar-title>
    </v-toolbar>
    <div v-if="model.UID">
      <v-list nav slim tile density="compact" class="metadata__list mt-2">
        <v-list-item v-if="model.Title" class="metadata__item">
          <div class="text-subtitle-2 font-weight-bold">{{ model.Title }}</div>
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

        <v-list-item :prepend-icon="model.getTypeIcon()" :title="model.getTypeInfo()" class="metadata__item">
        </v-list-item>

        <template v-if="model.Lat && model.Lng">
          <v-divider class="my-4"></v-divider>
          <!-- Clickable version commented out
          <v-list-item
            prepend-icon="mdi-map-marker"
            :title="model.getLatLng()"
            class="clickable metadata__item"
            @click.stop="model.copyLatLng"
          >
          </v-list-item>
          -->
          <v-list-item prepend-icon="mdi-map-marker" :title="model.getLatLng()" class="metadata__item"> </v-list-item>
          <div id="metadata-map" ref="mapContainer" class="metadata__map"></div>
        </template>
      </v-list>
    </div>
  </div>
</template>

<script>
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
