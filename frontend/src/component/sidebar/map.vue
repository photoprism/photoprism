<template>
  <div v-if="lat && lng && !isOfflineStyle" class="metadata__map" ref="mapContainer"></div>
</template>

<script>
let maplibregl;

export default {
  name: "PMap",
  props: {
    lat: {
      type: Number,
      required: true,
    },
    lng: {
      type: Number,
      required: true,
    },
  },
  data() {
    return {
      map: null,
      marker: null,
      mapLoaded: false,
      loadingMapLibre: false,
    };
  },
  computed: {
    isOfflineStyle() {
      return this.$config.values.settings.maps.style === "low-resolution";
    },
  },
  watch: {
    lat() {
      this.updateMapPosition();
    },
    lng() {
      this.updateMapPosition();
    },
  },
  async mounted() {
    if (!this.isOfflineStyle) {
      await this.loadMapAndInit();
    }
  },
  beforeUnmount() {
    if (this.map) {
      this.map.remove();
    }
  },
  methods: {
    async loadMapAndInit() {
      if (this.loadingMapLibre) {
        return;
      }

      this.loadingMapLibre = true;

      try {
        const module = await import("../../common/maplibregl.js");
        maplibregl = module.default;
        await this.$nextTick();
        await this.initMap();
      } catch (error) {
        console.error("Failed to load maplibregl:", error);
      } finally {
        this.loadingMapLibre = false;
      }
    },
    async initMap() {
      if (!this.$refs.mapContainer || !maplibregl) {
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
          center: [this.lng, this.lat],
          zoom: 13,
          interactive: true,
          attributionControl: { compact: true },
        });

        // Add zoom controls
        this.map.addControl(new maplibregl.NavigationControl({
          showCompass: false,
          showZoom: true,
          visualizePitch: false,
        }), 'top-right');

        this.map.on("error", (e) => {
          console.error("Map error:", e);
        });

        this.map.on("load", () => {
          this.mapLoaded = true;

          if (this.marker) {
            this.marker.remove();
          }

          this.marker = new maplibregl.Marker().setLngLat([this.lng, this.lat]).addTo(this.map);
          
          // Minimize attribution control by default
          const attrCtrl = this.$refs.mapContainer.querySelector(".maplibregl-ctrl-attrib");
          if (attrCtrl) {
            attrCtrl.classList.remove("maplibregl-compact-show");
            attrCtrl.removeAttribute("open");
          }
        });
      } catch (error) {
        console.error("Failed to initialize map:", error);
        this.mapLoaded = false;
      }
    },
    updateMapPosition() {
      if (this.map && this.mapLoaded) {
        this.map.setCenter([this.lng, this.lat]);
        if (this.marker) {
          this.marker.setLngLat([this.lng, this.lat]);
        }
      }
    },
  },
};
</script> 