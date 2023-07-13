<template>
  <v-dialog :value="show" persistent max-width="500" class="p-photo-set-location-dialog" @keydown.esc="cancel" @keydown.enter="confirm">
    <v-card raised elevation="24">
      <v-card-title primary-title class="pb-0">
        <v-layout row wrap>
          <v-flex xs10>
            <h3 class="headline mb-0">
              <translate>Set Location</translate>
            </h3>
          </v-flex>
          <v-flex xs2 text-xs-right>
            <v-icon>edit_location</v-icon>
          </v-flex>
        </v-layout>
      </v-card-title>
      <v-card-text fluid class="pt-3 px-3">
        <v-layout fluid row wrap>
          <v-layout row wrap>
            <v-flex xs2>
              <v-icon>warning</v-icon>
            </v-flex>
            <v-flex xs10>
              <translate>Change is applied to all currently selected photos</translate>
            </v-flex>
          </v-layout>
          <v-flex fluid text-xs-left align-self-center>
            <v-container fluid fill-height>
              <div id="map" style="width: 300px; height: 300px;">
                <v-sheet fluid></v-sheet>
              </div>
            </v-container>
            <v-text-field
                  v-model="latitude"
                  :label="$gettext('Latitude')"
                  placeholder=""
                  color="secondary-dark"
                  class="input-latitude background-inherit elevation-0"
                  autofocus hide-details
            ></v-text-field>
            <v-text-field
                  v-model="longitude"
                  :label="$gettext('Longitude')"
                  placeholder=""
                  color="secondary-dark"
                  class="iniput-longitude background-inherit elevation-0"
                  hide-details
            ></v-text-field>
          </v-flex>
        </v-layout>
      </v-card-text>
      <v-card-actions class="pt-0 pb-3 px-3">
        <v-layout row wrap class="pa-0">
          <v-flex xs12 text-xs-right>
            <v-btn depressed color="secondary-light" class="action-cancel mx-1" @click.stop="cancel">
              <translate>Cancel</translate>
            </v-btn>
            <v-btn depressed color="primary-button"
                   class="action-confirm white--text compact mx-0"
                   @click.stop="confirm">
              <translate>Apply</translate>
            </v-btn>
          </v-flex>
        </v-layout>
      </v-card-actions>
    </v-card>
  </v-dialog>
</template>
<script>
import maplibregl from "maplibre-gl";

export default {
  name: 'PPhotoSetLocationDialog',
  props: {
    show: Boolean,
  },
  data() {
    return {
      latitude: "",
      longitude: "",

      map: null,
      marker: null,
      stype: "",
      terrain: {
        'topo-v2': 'terrain_rgb',
        'outdoor-v2': 'terrain-rgb',
        '414c531c-926d-4164-a057-455a215c0eee': 'terrain_rgb_virtual',
      },
      attribution: '<a href="https://www.maptiler.com/copyright/" target="_blank">&copy; MapTiler</a> <a href="https://www.openstreetmap.org/copyright" target="_blank">&copy; OpenStreetMap contributors</a>',
      options: {},
      mapFont: ["Open Sans Regular"],
      config: this.$config.values,
      settings: this.$config.values.settings.maps,

    };
  },
  watch: {
    latitude: function(newLatitude, oldLatitude) {
      this.marker.setLngLat([this.longitude, newLatitude]);
    },
    longitude: function(newLongitude, oldLongitude) {
      this.marker.setLngLat([newLongitude, this.latitude]);
    }
  },
  mounted() {
    this.configureMap().then(() => this.renderMap());
  },
  methods: {
    configureMap() {
      return this.$config.load().finally(() => {
        const s = this.$config.values.settings.maps;
        let mapKey = "";

        if (this.$config.has("mapKey")) {
          // Remove non-alphanumeric characters from key.
          mapKey = this.$config.get("mapKey").replace(/[^a-z0-9]/gi, '');
        }

        const settings = this.$config.settings();

        switch (s.style) {
          case "basic":
          case "offline":
            this.style = "";
            break;
          case "hybrid":
            this.style = "414c531c-926d-4164-a057-455a215c0eee";
            break;
          case "outdoor":
            this.style = "outdoor-v2";
            break;
          case "topographique":
            this.style = "topo-v2";
            break;
          default:
            this.style = s.style;
        }

        if (!mapKey && this.style !== "low-resolution") {
          this.style = "";
        }

        let mapOptions = {
          container: "map",
          style: "https://api.maptiler.com/maps/" + this.style + "/style.json?key=" + mapKey,
          glyphs: "https://api.maptiler.com/fonts/{fontstack}/{range}.pbf?key=" + mapKey,
          attributionControl: true,
          customAttribution: this.attribution,
          zoom: 0,
        };

        if (this.style === "") {
          mapOptions = {
            container: "map",
            style: "https://cdn.photoprism.app/maps/default.json",
            glyphs: `https://cdn.photoprism.app/maps/font/{fontstack}/{range}.pbf`,
            attributionControl: true,
            zoom: 0,
          };
        } else if (this.style === "low-resolution") {
          mapOptions = {
            container: "map",
            style: {
              "version": 8,
              "sources": {
                "world": {
                  "type": "geojson",
                  "data": `${this.$config.staticUri}/geo/world.json`,
                  "maxzoom": 6
                }
              },
              "glyphs": `${this.$config.staticUri}/font/{fontstack}/{range}.pbf`,
              "layers": [
                {
                  "id": "background",
                  "type": "background",
                  "paint": {
                    "background-color": "#aadafe"
                  }
                },
                {
                  id: "land",
                  type: "fill",
                  source: "world",
                  // "source-layer": "land",
                  paint: {
                    "fill-color": "#cbe5ca",
                  },
                },
                {
                  "id": "country-abbrev",
                  "type": "symbol",
                  "source": "world",
                  "maxzoom": 3,
                  "layout": {
                    "text-field": "{abbrev}",
                    "text-font": ["Open Sans Semibold"],
                    "text-transform": "uppercase",
                    "text-max-width": 20,
                    "text-size": {
                      "stops": [[3, 10], [4, 11], [5, 12], [6, 16]]
                    },
                    "text-letter-spacing": {
                      "stops": [[4, 0], [5, 1], [6, 2]]
                    },
                    "text-line-height": {
                      "stops": [[5, 1.2], [6, 2]]
                    }
                  },
                  "paint": {
                    "text-halo-color": "#fff",
                    "text-halo-width": 1
                  },
                },
                {
                  "id": "country-border",
                  "type": "line",
                  "source": "world",
                  "paint": {
                    "line-color": "#226688",
                    "line-opacity": 0.25,
                    "line-dasharray": [6, 2, 2, 2],
                    "line-width": 1.2
                  }
                },
                {
                  "id": "country-name",
                  "type": "symbol",
                  "minzoom": 3,
                  "source": "world",
                  "layout": {
                    "text-field": "{name}",
                    "text-font": ["Open Sans Semibold"],
                    "text-max-width": 20,
                    "text-size": {
                      "stops": [[3, 10], [4, 11], [5, 12], [6, 16]]
                    }
                  },
                  "paint": {
                    "text-halo-color": "#fff",
                    "text-halo-width": 1
                  },
                },
              ],
            },
            attributionControl: false,
            customAttribution: '',
            zoom: 0,
          };
        }

        this.options = mapOptions;
      });
    },
    onDragEnd() {
        const lngLat = this.marker.getLngLat();
        this.latitude = lngLat.lat;
        this.longitude = lngLat.lng;
    },
    onMapClick(e) {
      let lngLat = e.lngLat.wrap();
      this.marker.setLngLat([lngLat.lng, lngLat.lat]);
      this.latitude = lngLat.lat;
      this.longitude = lngLat.lng;
    },
    renderMap() {

    let mapOptions = {
        container: "map",
        style: "https://cdn.photoprism.app/maps/default.json",
        glyphs: `https://cdn.photoprism.app/maps/font/{fontstack}/{range}.pbf`,
        attributionControl: true,
        customAttribution: this.attribution,
        zoom: 0,
    };
      this.map = new maplibregl.Map(this.options);
      this.map.setLanguage(this.$config.values.settings.ui.language.split("-")[0]);

      const controlPos = this.$rtl ? 'top-left' : 'top-right';

      // Show navigation control.
      this.map.addControl(new maplibregl.NavigationControl({
        visualizePitch: true,
        showZoom: true,
        showCompass: true
      }), controlPos);

      this.marker = new maplibregl.Marker({draggable: true})
        .setLngLat([0, 0])
        .addTo(this.map);

      this.marker.on('dragend', this.onDragEnd);

      this.map.on("click", this.onMapClick);

      // Show terrain control, if supported.
      if (this.terrain[this.style]) {
        this.map.addControl(new maplibregl.TerrainControl({
          source: this.terrain[this.style],
          exaggeration: 1
        }));
      }

      // Show locate control.
      this.map.addControl(new maplibregl.GeolocateControl({
        positionOptions: {
          enableHighAccuracy: true
        },
        trackUserLocation: true
      }), controlPos);

      this.map.on("load", () => this.onMapLoad());
    },
    onMapLoad() {
      this.map.resize();
    },

    cancel() {
      this.$emit('cancel');
    },
    confirm() {
      this.$emit('confirm', this.latitude, this.longitude);
    },
  },
};
</script>
  