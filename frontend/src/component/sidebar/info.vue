<template>
  <div class="p-sidebar-info metadata">
    <v-toolbar density="comfortable" color="navigation">
      <v-btn :icon="$isRtl ? 'mdi-chevron-left' : 'mdi-chevron-right'" :title="$gettext('Close')" @click.stop="close()"></v-btn>
      <v-toolbar-title>{{ $gettext(`Information`) }}</v-toolbar-title>
    </v-toolbar>
    <div v-if="model.UID">
      <v-list nav slim tile density="compact" class="metadata__list mt-2">
        <v-list-item
          v-for="item in metadataItems"
          :key="item.key"
          :prepend-icon="item.icon"
          class="metadata__item"
          :class="{ clickable: item.clickable }"
          @click.stop="onItemClick(item.action)"
        >
          <div class="metadata__body">
            <div :class="item.className">
              {{ item.text }}
            </div>
          </div>
        </v-list-item>

        <template v-if="locationModel.Lat && locationModel.Lng">
          <v-divider class="my-4"></v-divider>
          <v-list-item
            v-tooltip="$gettext(`Coordinates`)"
            prepend-icon="mdi-map-marker"
            :title="locationTitle"
            class="clickable metadata__item"
            @click.stop="copyLocation()"
          >
          </v-list-item>
          <v-list-item v-if="featPlaces" class="mx-0 px-0">
            <p-map :latlng="[locationModel.Lat, locationModel.Lng]"></p-map>
          </v-list-item>
        </template>
      </v-list>
    </div>
  </div>
</template>

<script>
import { hasMetadataText, metadataIcon, metadataLabel, metadataLayout, metadataText, MetadataView } from "common/metadata";
import PMap from "component/map.vue";
import { Photo } from "model/photo";

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
    collection: {
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
      details: new Photo(this.modelValue),
      featPlaces: this.$config.feature("places"),
      loading: false,
      requestUid: "",
    };
  },
  computed: {
    model() {
      return this.details?.UID === this.modelValue?.UID ? this.details : this.modelValue;
    },
    locationModel() {
      if (this.model?.Lat && this.model?.Lng) {
        return this.model;
      }

      return this.modelValue || {};
    },
    locationTitle() {
      if (typeof this.locationModel?.getLatLng === "function") {
        return this.locationModel.getLatLng();
      }

      if (!this.locationModel?.Lat || !this.locationModel?.Lng) {
        return "";
      }

      return `${this.locationModel.Lat.toFixed(5)}°N\u2003${this.locationModel.Lng.toFixed(5)}°E`;
    },
    lightboxLayout() {
      return metadataLayout(this.$config.getSettings(), MetadataView.Lightbox);
    },
    metadataItems() {
      return this.lightboxLayout
        .map((fieldId, index) => {
          if (fieldId === "location" && (!this.model?.Lat || !this.model?.Lng)) {
            return null;
          } else if (!hasMetadataText(this.model, fieldId)) {
            return null;
          }

          return {
            key: `${fieldId}-${index}`,
            label: metadataLabel(fieldId),
            text: metadataText(this.model, fieldId),
            icon: metadataIcon(fieldId, this.model),
            className: this.itemClass(fieldId),
            clickable: fieldId === "location" && !!this.model?.Lat && !!this.model?.Lng,
            action: fieldId === "location" ? "copy-location" : "",
          };
        })
        .filter(Boolean);
    },
  },
  watch: {
    modelValue: {
      handler() {
        this.syncDetails();
      },
      immediate: true,
      deep: false,
    },
  },
  methods: {
    close() {
      this.$emit("close");
    },
    syncDetails() {
      this.details = new Photo(this.modelValue);

      const uid = this.modelValue?.UID;

      if (!uid) {
        this.requestUid = "";
        this.loading = false;
        return;
      }

      this.requestUid = uid;
      this.loading = true;

      new Photo()
        .find(uid)
        .then((photo) => {
          if (this.requestUid !== uid) {
            return;
          }

          this.details = photo;
        })
        .finally(() => {
          if (this.requestUid === uid) {
            this.loading = false;
          }
        });
    },
    itemClass(fieldId) {
      return ["metadata__value", `metadata__value--${fieldId}`, `meta-${fieldId}`].join(" ");
    },
    onItemClick(action) {
      if (action === "copy-location" && this.model?.Lat && this.model?.Lng) {
        this.copyLocation();
      }
    },
    copyLocation() {
      if (typeof this.locationModel?.copyLatLng === "function") {
        this.locationModel.copyLatLng();
        return;
      }

      if (!this.locationModel?.Lat || !this.locationModel?.Lng) {
        return;
      }

      navigator.clipboard?.writeText(`${this.locationModel.Lat.toString()},${this.locationModel.Lng.toString()}`);
    },
  },
};
</script>
