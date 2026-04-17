<template>
  <div class="p-sidebar-info metadata">
    <v-toolbar density="comfortable" color="navigation">
      <v-btn :icon="$isRtl ? 'mdi-chevron-left' : 'mdi-chevron-right'" :title="$gettext('Close')" @click.stop="close()"></v-btn>
      <v-toolbar-title>{{ $gettext(`Information`) }}</v-toolbar-title>
    </v-toolbar>
    <div v-if="model.UID">
      <v-list nav slim tile density="compact" class="metadata__list mt-2">
        <v-list-item v-if="model.Title" class="metadata__item">
          <div v-tooltip="$pgettext(`Photo`, `Title`)" class="text-subtitle-2 meta-title">{{ model.Title }}</div>
        </v-list-item>
        <v-list-item v-if="model.Caption" class="metadata__item">
          <div v-tooltip="$gettext('Caption')" class="text-body-2 meta-caption">{{ model.Caption }}</div>
        </v-list-item>
        <v-divider v-if="model.Title || model.Caption" class="my-4"></v-divider>
        <v-list-item v-tooltip="$gettext(`Taken`)" :title="formatTime(model)" prepend-icon="mdi-calendar" class="metadata__item">
        </v-list-item>

        <v-list-item v-tooltip="$gettext(`Size`)" :title="model.getTypeInfo()" :prepend-icon="model.getTypeIcon()" class="metadata__item"> </v-list-item>

        <template v-if="model.Lat && model.Lng">
          <v-divider class="my-4"></v-divider>
          <v-list-item
            v-tooltip="$gettext(`Location`)"
            prepend-icon="mdi-map-marker"
            :title="model.getLatLng()"
            class="clickable metadata__item"
            @click.stop="model.copyLatLng()"
          >
          </v-list-item>
          <v-list-item v-if="featPlaces" class="mx-0 px-0">
            <p-map :latlng="[model.Lat, model.Lng]"></p-map>
          </v-list-item>
        </template>
      </v-list>

      <div v-if="featPeople" class="metadata__section px-4 pb-6">
        <v-divider class="mb-4"></v-divider>
        <div class="metadata__section-header">
          <div class="text-subtitle-2">{{ $gettext("People") }}</div>
          <div class="metadata__section-actions">
            <v-btn size="small" variant="text" :disabled="busy || photoLoading" @click.stop="$emit('toggle-markers')">
              {{ markersVisible ? $gettext("Hide Markers") : $gettext("Show Markers") }}
            </v-btn>
            <v-btn v-if="canEdit" size="small" variant="flat" color="primary" :disabled="busy || markersBusy || photoLoading" @click.stop="$emit('toggle-adding-marker')">
              {{ addingMarker ? $gettext("Cancel") : $gettext("Add Face") }}
            </v-btn>
          </div>
        </div>

        <div v-if="addingMarker" class="metadata__hint text-body-2 mb-3">
          {{ $gettext("Drag a square over a face in the picture to add a marker.") }}
        </div>

        <v-progress-linear v-if="photoLoading" indeterminate color="primary" class="mb-3"></v-progress-linear>

        <v-alert v-else-if="markers.length === 0" color="surface-variant" icon="mdi-account-search-outline" variant="outlined" class="mb-2">
          {{ $gettext("No face markers yet.") }}
        </v-alert>

        <div v-else class="metadata__markers">
          <div v-for="m in markers" :key="m.UID" class="metadata__marker" :class="{ 'metadata__marker--new': m.UID === newMarkerUid }">
            <v-img :src="m.thumbnailUrl('tile_160')" aspect-ratio="1" class="metadata__marker-thumb"></v-img>
            <div class="metadata__marker-body">
              <v-text-field
                v-if="m.SubjUID"
                v-model="m.Name"
                :disabled="busy || markersBusy"
                :readonly="true"
                autocomplete="off"
                autocorrect="off"
                hide-details
                single-line
                clearable
                persistent-clear
                clear-icon="mdi-eject"
                density="comfortable"
                class="metadata__marker-input"
                @click:clear="onClearSubject(m)"
              ></v-text-field>
              <v-combobox
                v-else
                v-model:search="m.Name"
                :items="$config.values.people"
                item-title="Name"
                item-value="Name"
                :disabled="busy || markersBusy || !canEdit"
                :menu-props="menuProps"
                return-object
                hide-no-data
                hide-details
                single-line
                open-on-clear
                append-icon=""
                prepend-inner-icon="mdi-account-plus"
                density="comfortable"
                class="metadata__marker-input text-selectable"
                @update:model-value="(person) => onSetPerson(m, person)"
                @blur="(ev) => onSetName(m, ev)"
                @keyup.enter="(ev) => onSetName(m, ev)"
              >
              </v-combobox>
            </div>
            <v-btn
              v-if="!m.SubjUID && !m.Invalid"
              icon="mdi-close"
              size="small"
              variant="text"
              density="comfortable"
              class="metadata__marker-action"
              :disabled="busy || markersBusy || !canEdit"
              :title="$gettext('Remove')"
              @click.stop="onReject(m)"
            ></v-btn>
          </div>
        </div>
      </div>
    </div>
    <p-confirm-dialog
      :visible="confirm.visible"
      icon="mdi-account-plus"
      :icon-size="42"
      :text="confirm?.model?.Name ? $gettext('Add %{s}?', { s: confirm.model.Name }) : $gettext('Add person?')"
      @close="onCancelSetName"
      @confirm="onConfirmSetName"
    ></p-confirm-dialog>
  </div>
</template>

<script>
import { DateTime } from "luxon";
import * as formats from "options/formats";

import PMap from "component/map.vue";
import PConfirmDialog from "component/confirm/dialog.vue";

export default {
  name: "PSidebarInfo",
  components: {
    PMap,
    PConfirmDialog,
  },
  props: {
    modelValue: {
      type: Object,
      default: () => {},
    },
    photo: {
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
    canEdit: {
      type: Boolean,
      default: false,
    },
    markersVisible: {
      type: Boolean,
      default: false,
    },
    addingMarker: {
      type: Boolean,
      default: false,
    },
    markersBusy: {
      type: Boolean,
      default: false,
    },
    photoLoading: {
      type: Boolean,
      default: false,
    },
    newMarkerUid: {
      type: String,
      default: "",
    },
  },
  emits: ["update:modelValue", "close", "toggle-markers", "toggle-adding-marker", "reload-markers"],
  data() {
    return {
      featPlaces: this.$config.feature("places"),
      featPeople: this.$config.feature("people"),
      busy: false,
      markers: this.photo?.UID && typeof this.photo.getMarkers === "function" ? this.photo.getMarkers(true) : [],
      confirm: {
        visible: false,
        model: null,
      },
      menuProps: {
        openOnClick: false,
        openOnFocus: true,
        closeOnBack: true,
        closeOnContentClick: true,
        disableInitialFocus: true,
        persistent: false,
        scrim: true,
        openDelay: 0,
        closeDelay: 0,
        opacity: 0,
        density: "compact",
        maxHeight: 300,
        locationStrategy: "connected",
        scrollStrategy: "reposition",
        origin: "auto",
      },
    };
  },
  computed: {
    model() {
      return this.modelValue;
    },
  },
  watch: {
    photo() {
      this.refreshMarkers();
    },
    newMarkerUid() {
      this.refreshMarkers();
    },
  },
  methods: {
    close() {
      this.$emit("close");
    },
    refreshMarkers() {
      this.markers = this.photo?.UID && typeof this.photo.getMarkers === "function" ? this.photo.getMarkers(true) : [];

      if (this.confirm?.model?.UID && !this.markers.some((marker) => marker.UID === this.confirm.model.UID)) {
        this.confirm.visible = false;
        this.confirm.model = null;
      }
    },
    formatTime(model) {
      if (!model || !model.TakenAtLocal) {
        return this.$gettext("Unknown");
      }

      const dateTime = DateTime.fromISO(model.TakenAtLocal, { zone: "UTC" });

      if (model.TimeZone && model.TimeZone !== "Local" && model.TimeZone !== "UTC") {
        return dateTime.setZone(model.TimeZone, { keepLocalTime: true }).toLocaleString(formats.DATETIME_MED_TZ);
      } else {
        return dateTime.toLocaleString(formats.DATETIME_MED);
      }
    },
    onReject(model) {
      if (this.busy || this.markersBusy || !model) {
        return;
      }

      this.busy = true;
      this.$notify.blockUI("busy");

      model
        .reject()
        .then(() => {
          this.$emit("reload-markers");
        })
        .finally(() => {
          this.$notify.unblockUI();
          this.busy = false;
        });
    },
    onClearSubject(model) {
      if (this.busy || this.markersBusy || !model) {
        return;
      }

      this.busy = true;
      this.$notify.blockUI("busy");

      model
        .clearSubject()
        .then(() => {
          this.$emit("reload-markers");
        })
        .finally(() => {
          this.$notify.unblockUI();
          this.busy = false;
        });
    },
    onSetPerson(model, person) {
      if (typeof person === "object" && model?.UID && person?.Name) {
        model.Name = person.Name;
        model.SubjUID = person.UID || "";
        this.setName(model);
      }

      return true;
    },
    onSetName(model, ev) {
      if (this.busy || this.markersBusy || !model || !this.canEdit) {
        return;
      }

      if (this.confirm.visible && this.confirm.model && this.confirm.model.UID !== model.UID) {
        return;
      }

      const name = model?.Name;

      if (!name) {
        this.onCancelSetName();
        return;
      }

      this.confirm.model = model;

      const people = this.$config.values?.people;

      if (people) {
        const found = people.find((person) => person.Name.localeCompare(name, "en", { sensitivity: "base" }) === 0);
        if (found) {
          model.Name = found.Name;
          model.SubjUID = found.UID || "";
          this.setName(model);
          return;
        }
      }

      model.Name = name;
      model.SubjUID = "";

      if (ev && ev.key === "Enter" && !ev.isComposing && !ev.repeat) {
        this.setName(model);
      } else {
        this.confirm.visible = true;
      }
    },
    onConfirmSetName() {
      if (!this.confirm?.model?.Name) {
        return;
      }

      this.setName(this.confirm.model);
    },
    onCancelSetName() {
      if (this.confirm && this.confirm.model) {
        this.confirm.model.Name = "";
        this.confirm.model.SubjUID = "";
      }

      this.confirm.visible = false;
      this.confirm.model = null;
    },
    setName(model) {
      if (this.busy || this.markersBusy || !model) {
        return;
      }

      this.busy = true;
      this.$notify.blockUI("busy");

      return model
        .setName()
        .then(() => {
          this.$emit("reload-markers");
        })
        .finally(() => {
          this.$notify.unblockUI();
          this.busy = false;
          this.confirm.model = null;
          this.confirm.visible = false;
        });
    },
  },
};
</script>

<style scoped>
.metadata__section-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  margin-bottom: 12px;
  flex-wrap: wrap;
}

.metadata__section-actions {
  display: flex;
  gap: 8px;
  flex-wrap: wrap;
}

.metadata__markers {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.metadata__marker {
  display: grid;
  grid-template-columns: 72px minmax(0, 1fr) auto;
  gap: 12px;
  align-items: start;
}

.metadata__marker--new {
  padding: 8px;
  border-radius: 12px;
  background: rgba(var(--v-theme-info), 0.08);
}

.metadata__marker-thumb {
  border-radius: 12px;
  overflow: hidden;
  background: rgba(var(--v-theme-surface-variant), 0.35);
}

.metadata__marker-body {
  min-width: 0;
}

.metadata__marker-input {
  margin: 0;
}

.metadata__marker-action {
  margin-top: 4px;
}

.metadata__hint {
  opacity: 0.8;
}

@media (max-width: 420px) {
  .metadata__marker {
    grid-template-columns: 64px minmax(0, 1fr);
  }

  .metadata__marker-action {
    grid-column: 2;
    justify-self: end;
    margin-top: -6px;
  }
}
</style>
