<template>
  <div ref="map" class="p-map"></div>
</template>

<script>
import * as map from "common/map";

let maplibregl = null;

export default {
  name: "PMap",
  props: {
    latlng: {
      type: Array,
      default: () => [0.0, 0.0],
      validator: (value) => Array.isArray(value) && value.length === 2,
    },
    zoom: {
      type: Number,
      default: 9,
    },
    style: {
      type: String,
      default: "embedded",
    },
    // Interactive mode props
    interactive: {
      type: Boolean,
      default: false,
    },
    draggable: {
      type: Boolean,
      default: false,
    },
    showControls: {
      type: Boolean,
      default: false,
    },
    clickable: {
      type: Boolean,
      default: false,
    },
  },
  emits: ["update:latlng", "marker-moved", "map-clicked"],
  data() {
    const settings = this.$config.getSettings();

    return {
      map: null,
      marker: null,
      position: [0.0, 0.0],
      animate: settings.maps.animate,
      options: {
        container: null,
        // Styles can be edited/created with https://maplibre.org/maputnik/.
        // To test new styles, put the style file in /assets/static/maps
        // and include it from there e.g. "/static/maps/embedded.json":
        // style: "/static/maps/embedded.json",
        style: `https://cdn.photoprism.app/maps/${this.style}.json`,
        glyphs: `https://cdn.photoprism.app/maps/font/{fontstack}/{range}.pbf`,
        zoom: this.zoom,
        interactive: this.interactive,
        attributionControl: false,
      },
      loaded: false,
    };
  },
  watch: {
    latlng() {
      this.updatePosition();
    },
  },
  mounted() {
    map.load().then((m) => {
      maplibregl = m;
      this.initMap();
    });
  },
  beforeUnmount() {
    if (this.map) {
      this.map.remove();
    }
  },
  methods: {
    initMap() {
      if (this.map || !this.$refs.map || !maplibregl) {
        return;
      }

      try {
        this.options.container = this.$refs.map;

        // Set center based on coordinates or default
        if (!(this.latlng[0] && this.latlng[1] && !(this.latlng[0] === 0 && this.latlng[1] === 0))) {
          this.options.zoom = 2;
          this.options.center = [0, 20];
        } else {
          this.options.center = [this.latlng[1], this.latlng[0]]; // Convert [lat, lng] to [lng, lat] for MapLibre
        }

        this.map = new maplibregl.Map(this.options);

        // Add controls if requested
        if (this.showControls) {
          this.map.addControl(
            new maplibregl.NavigationControl({
              showCompass: true,
              showZoom: true,
              visualizePitch: false,
            }),
            "top-right"
          );

          this.map.addControl(new maplibregl.ScaleControl({ maxWidth: 80, unit: "metric" }), "bottom-left");

          this.map.addControl(
            new maplibregl.GeolocateControl({
              positionOptions: {
                enableHighAccuracy: true,
              },
              trackUserLocation: true,
            }),
            "top-right"
          );
        }

        this.map.on("error", (e) => {
          console.error("map:", e);
        });

        // Handle missing style images
        this.map.on("styleimagemissing", (e) => {
          const emptyImage = new ImageData(1, 1);
          if (e && e.id) {
            this.map.addImage(e.id, emptyImage);
          }
        });

        this.map.on("load", () => {
          this.loaded = true;
          this.updatePosition();
          this.map.resize();
        });

        // Add click handler for interactive mode
        if (this.clickable) {
          this.map.on("click", (e) => {
            const lat = e.lngLat.lat;
            const lng = e.lngLat.lng;
            this.$emit("map-clicked", { lat, lng });
            this.$emit("update:latlng", [lat, lng]);
          });
        }
      } catch (error) {
        console.error("map: initialization failed", error);
        this.loaded = false;
      }
    },
    updatePosition() {
      if (!this.map || !this.loaded) {
        return;
      }

      if (this.position[0] === this.latlng[1] && this.position[1] === this.latlng[0] && this.marker) {
        return;
      }

      // Skip invalid or empty coordinates
      if (!(this.latlng[0] && this.latlng[1] && !(this.latlng[0] === 0 && this.latlng[1] === 0))) {
        if (this.marker) {
          this.marker.remove();
          this.marker = null;
        }
        return;
      }

      this.position = [this.latlng[1], this.latlng[0]]; // Convert [lat, lng] to [lng, lat] for MapLibre

      if (this.animate > 0) {
        this.map.flyTo({
          center: this.position,
          zoom: this.interactive ? this.zoom : undefined, // Only set zoom in interactive mode
          duration: this.animate,
          essential: true, // Respects prefers-reduced-motion
        });
      } else {
        // Use setCenter for instant positioning (no animation)
        if (this.interactive) {
          this.map.setCenter(this.position, {
            zoom: this.zoom,
            animate: false,
          });
        } else {
          this.map.setCenter(this.position);
        }
      }

      if (this.marker) {
        this.marker.setLngLat(this.position);
      } else {
        this.marker = new maplibregl.Marker({
          color: "#3fb4df",
          draggable: this.draggable,
        })
          .setLngLat(this.position)
          .addTo(this.map);

        // Add drag event listener for draggable markers
        if (this.draggable) {
          this.marker.on("dragend", () => {
            const lngLat = this.marker.getLngLat();
            this.$emit("marker-moved", { lat: lngLat.lat, lng: lngLat.lng });
            this.$emit("update:latlng", [lngLat.lat, lngLat.lng]);
          });
        }
      }
    },
    // Public method to remove marker
    removeMarker() {
      if (this.marker) {
        this.marker.remove();
        this.marker = null;
      }
    },
    // Public method to fly to coordinates
    flyTo(lat, lng, zoom = this.zoom) {
      if (this.map) {
        if (this.animate > 0) {
          this.map.flyTo({
            center: [lng, lat],
            zoom: zoom,
            duration: this.animate,
            essential: true,
          });
        } else {
          this.map.jumpTo({
            center: [lng, lat],
            zoom: zoom,
          });
        }
      }
    },
  },
};
</script>
