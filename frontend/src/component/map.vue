<template>
  <div class="p-map" :class="{ 'is-marker-clickable': markerClickable }">
    <div v-if="mapError" class="pa-3" role="status">{{ mapError }}</div>
    <div v-show="!mapError" ref="map" class="p-map-canvas"></div>
  </div>
</template>

<script>
import { markRaw } from "vue";
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
    // When true, attaches a click listener to the marker's DOM element
    // that emits `marker-clicked`. The marker's cursor is gated on the
    // same prop so it only reads as pointer when the consumer is
    // actually listening (see `places.css` `.p-map.is-marker-clickable`).
    markerClickable: {
      type: Boolean,
      default: false,
    },
    // Override animation duration (ms). -1 = use global setting, 0 = no animation.
    animateDuration: {
      type: Number,
      default: -1,
    },
  },
  emits: ["update:latlng", "marker-moved", "marker-clicked", "map-clicked"],
  // data initializes the map, marker, and animation state.
  data() {
    const settings = this.$config.getSettings();

    return {
      map: null,
      marker: null,
      position: [0.0, 0.0],
      animate: this.$util.mapAnimateDuration(settings, this.animateDuration),
      loaded: false,
      markerDragged: false,
      mapError: "",
    };
  },
  watch: {
    // latlng applies coordinate changes from the parent.
    latlng() {
      this.updatePosition();
    },
  },
  // mounted loads the renderer only when this browser supports maps.
  mounted() {
    this._mapUnmounted = false;
    if (!map.supportsWebGL2()) {
      this.showMapError();
      return;
    }
    map.load().then((m) => {
      if (this._mapUnmounted) {
        return;
      }
      maplibregl = m;
      this.initMap();
    }).catch(this.showMapError);
  },
  // beforeUnmount releases the map and prevents delayed initialization.
  beforeUnmount() {
    this._mapUnmounted = true;
    if (this.map) {
      this.map.remove();
      this.map = null;
    }
  },
  methods: {
    // showMapError displays a map-only fallback without blocking location controls.
    showMapError() {
      this.loaded = false;
      this.mapError = this.$gettext("Maps are unavailable. Try another browser or device.");
    },
    // initMap creates the renderer, controls, and coordinate event handlers.
    initMap() {
      if (this.map || !this.$refs.map || !maplibregl) {
        return;
      }

      try {
        const position = this.lngLat(this.latlng);

        // MapLibre stays raw, since a reactive proxy breaks the read-only values it
        // memoizes on its own objects, e.g. the channels of parsed style colors.
        this.map = markRaw(
          new maplibregl.Map({
            container: this.$refs.map,
            // Styles can be edited with https://maplibre.org/maputnik/; to test one, put it in
            // /assets/static/maps and use e.g. "/static/maps/embedded.json" as style.
            style: `https://cdn.photoprism.app/maps/${this.style}.json`,
            glyphs: `https://cdn.photoprism.app/maps/font/{fontstack}/{range}.pbf`,
            center: position || [0, 20],
            zoom: position ? this.zoom : 2,
            interactive: this.interactive,
            attributionControl: false,
            locale: map.locale(),
          })
        );

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

        this.map.setMissingStyleImageResolver((id) => {
          this.map.addImage(id, new ImageData(1, 1));
        });

        this.map.on("load", () => {
          this.loaded = true;
          this.updatePosition();
          this.map.resize();
        });

        // Add click handler for interactive mode
        if (this.clickable) {
          this.map.on("mousedown", this.onMapPointerStart);
          this.map.on("touchstart", this.onMapPointerStart);
          this.map.on("click", (e) => {
            if (this.markerDragged) {
              return;
            }

            const lat = e.lngLat.lat;
            const lng = e.lngLat.lng;
            this.map.stop();
            this.updatePosition([lat, lng], false);
            this.$emit("map-clicked", { lat, lng });
            this.$emit("update:latlng", [lat, lng]);
          });
        }
      } catch (error) {
        console.error("map: initialization failed", error);
        this.showMapError();
      }
    },
    // onMapPointerStart allows placement clicks from a new pointer gesture.
    onMapPointerStart() {
      this.markerDragged = false;
    },
    // lngLat converts [lat, lng] to MapLibre's [lng, lat] order, or returns null for
    // missing or out-of-range coordinates; [0, 0] means no location has been set.
    lngLat(latlng) {
      const lat = parseFloat(latlng?.[0]);
      const lng = parseFloat(latlng?.[1]);

      if (!(Math.abs(lat) <= 90 && Math.abs(lng) <= 180) || (lat === 0 && lng === 0)) {
        return null;
      }

      return [lng, lat];
    },
    // updatePosition places the marker and optionally recenters external coordinate changes.
    updatePosition(latlng = this.latlng, recenter = true) {
      if (!this.map || !this.loaded) {
        return;
      }

      const position = this.lngLat(latlng);

      if (!position) {
        this.removeMarker();
        return;
      }

      if (this.marker && this.position[0] === position[0] && this.position[1] === position[1]) {
        return;
      }

      this.position = position;

      if (recenter) {
        // Only interactive maps zoom in. Animations are not essential, so MapLibre
        // skips them when the system prefers reduced motion.
        const camera = { center: this.position, zoom: this.interactive ? this.zoom : undefined };

        if (this.animate > 0) {
          this.map.flyTo({ ...camera, duration: this.animate, essential: false });
        } else {
          this.map.jumpTo(camera);
        }
      }

      if (this.marker) {
        this.marker.setLngLat(this.position);
      } else {
        this.marker = markRaw(
          new maplibregl.Marker({
            color: "#3fb4df",
            draggable: this.draggable,
          })
            .setLngLat(this.position)
            .addTo(this.map)
        );

        // Add drag event listener for draggable markers
        if (this.draggable) {
          this.marker.on("dragstart", () => {
            // Camera changes can reset MapLibre's click suppression during a drag.
            this.markerDragged = true;
            this.map.stop();
          });
          this.marker.on("dragend", () => {
            const lngLat = this.marker.getLngLat();
            this.updatePosition([lngLat.lat, lngLat.lng], false);
            this.$emit("marker-moved", { lat: lngLat.lat, lng: lngLat.lng });
            this.$emit("update:latlng", [lngLat.lat, lngLat.lng]);
          });
        }

        // MapLibre renders the default marker as a <button>, so keyboard activation works;
        // stopping propagation keeps the click from reaching a map-level handler.
        if (this.markerClickable) {
          const el = this.marker.getElement();
          if (el) {
            el.addEventListener("click", (ev) => {
              ev.stopPropagation();
              const lngLat = this.marker.getLngLat();
              this.$emit("marker-clicked", { lat: lngLat.lat, lng: lngLat.lng });
            });
          }
        }
      }
    },
    // removeMarker removes the location marker from the map, if any.
    removeMarker() {
      if (this.marker) {
        this.marker.remove();
        this.marker = null;
      }
    },
    // flyTo moves the map to the given coordinates, animated unless animations are off.
    flyTo(lat, lng, zoom = this.zoom) {
      if (this.map) {
        if (this.animate > 0) {
          this.map.flyTo({
            center: [lng, lat],
            zoom: zoom,
            duration: this.animate,
            essential: false,
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
