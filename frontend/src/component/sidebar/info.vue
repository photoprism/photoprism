<template>
  <div class="p-sidebar-info metadata">
    <v-toolbar density="comfortable" color="navigation">
      <v-btn :icon="$isRtl ? 'mdi-chevron-left' : 'mdi-chevron-right'" :title="$gettext('Close')" @click.stop="close()"></v-btn>
      <v-toolbar-title>{{ $gettext(`Information`) }}</v-toolbar-title>
    </v-toolbar>
    <div v-if="model.UID">
      <v-list nav slim tile density="compact" class="metadata__list mt-2">
        <!-- Title -->
        <v-list-item v-if="editingField === 'title' || model.Title || isEditable" class="metadata__item">
          <v-text-field
            v-if="editingField === 'title'"
            v-model="p.Title"
            :placeholder="$pgettext('Photo', 'Title')"
            :rules="[textRule]"
            variant="plain"
            density="compact"
            hide-details="auto"
            autocomplete="off"
            class="meta-inline-edit meta-inline-title"
            @keydown.enter.prevent="confirmField"
            @keydown.escape.prevent="cancelEditing"
            @blur="cancelEditing"
          ></v-text-field>
          <div v-else-if="model.Title" class="text-subtitle-2 meta-title">{{ model.Title }}</div>
          <div v-else class="meta-add-prompt" @click.stop="startEditing('title')">{{ $pgettext("Photo", "Title") }}</div>
          <template v-if="isEditable" #append>
            <v-icon
              v-if="editingField === 'title'"
              icon="mdi-check"
              size="small"
              class="meta-inline-confirm"
              @mousedown.prevent
              @click.stop="confirmField"
            ></v-icon>
            <v-icon v-else icon="mdi-pencil-outline" size="small" class="meta-inline-pencil" @click.stop="startEditing('title')"></v-icon>
          </template>
        </v-list-item>

        <!-- Caption -->
        <v-list-item v-if="editingField === 'caption' || model.Caption || isEditable" class="metadata__item">
          <v-textarea
            v-if="editingField === 'caption'"
            v-model="p.Caption"
            :placeholder="$gettext('Caption')"
            variant="plain"
            density="compact"
            auto-grow
            :max-rows="6"
            hide-details="auto"
            autocomplete="off"
            class="meta-inline-edit meta-inline-caption"
            @keydown.escape.prevent="cancelEditing"
            @blur="cancelEditing"
          ></v-textarea>
          <div v-else-if="model.Caption" class="text-body-2 meta-caption meta-scrollable" v-html="captionHtml"></div>
          <div v-else class="meta-add-prompt" @click.stop="startEditing('caption')">{{ $gettext("Caption") }}</div>
          <template v-if="isEditable" #append>
            <v-icon
              v-if="editingField === 'caption'"
              icon="mdi-check"
              size="small"
              class="meta-inline-confirm"
              @mousedown.prevent
              @click.stop="confirmField"
            ></v-icon>
            <v-icon v-else icon="mdi-pencil-outline" size="small" class="meta-inline-pencil" @click.stop="startEditing('caption')"></v-icon>
          </template>
        </v-list-item>

        <v-divider v-if="editingField === 'title' || editingField === 'caption' || model.Title || model.Caption || isEditable" class="my-4"></v-divider>
        <v-list-item v-tooltip="$gettext(`Taken`)" :title="formatTime(model)" prepend-icon="mdi-calendar" class="metadata__item">
          <template v-if="isEditable" #append>
            <v-icon icon="mdi-pencil-outline" size="small" class="meta-inline-pencil" @click.stop="dateTimeDialog = true"></v-icon>
          </template>
        </v-list-item>

        <v-list-item v-tooltip="$gettext(`Size`)" :title="model.getTypeInfo()" :prepend-icon="model.getTypeIcon()" class="metadata__item"> </v-list-item>

        <v-list-item v-if="cameraInfo || isEditable" v-tooltip="$gettext('Camera')" :title="cameraInfo" prepend-icon="mdi-camera" class="metadata__item">
          <template v-if="isEditable" #append>
            <v-icon icon="mdi-pencil-outline" size="small" class="meta-inline-pencil" @click.stop="cameraDialog = true"></v-icon>
          </template>
        </v-list-item>

        <v-list-item v-if="lensInfo" v-tooltip="$gettext('Lens')" :title="lensInfo" prepend-icon="mdi-camera-iris" class="metadata__item"> </v-list-item>

        <template v-if="model.Lat && model.Lng">
          <v-divider class="my-4"></v-divider>
          <v-list-item v-if="placeName" v-tooltip="$gettext('Location')" :title="placeName" prepend-icon="mdi-map-marker" class="metadata__item"> </v-list-item>
          <v-list-item
            v-tooltip="$gettext(`Coordinates`)"
            :title="altitude ? model.getLatLng() + ' \u00b7 ' + altitude : model.getLatLng()"
            class="clickable metadata__item"
            @click.stop="model.copyLatLng()"
          >
            <template v-if="isEditable && featPlaces" #append>
              <v-icon icon="mdi-pencil-outline" size="small" class="meta-inline-pencil" @click.stop.prevent="locationDialog = true"></v-icon>
            </template>
          </v-list-item>
          <v-list-item v-if="featPlaces" class="mx-0 px-0">
            <p-map :latlng="[model.Lat, model.Lng]" :animate-duration="0"></p-map>
          </v-list-item>
        </template>

        <template v-if="people.length > 0">
          <v-divider class="my-4"></v-divider>
          <v-list-item class="metadata__item">
            <div class="text-subtitle-2">{{ $gettext("People") }}</div>
          </v-list-item>
          <v-list-item
            v-for="m in people"
            :key="m.UID || m.CropID"
            class="metadata__item metadata__person-row"
            :class="{ clickable: m.Name && m.SubjUID && !isEditingPerson(m) }"
            @click.stop.prevent="m.Name && m.SubjUID && !isEditingPerson(m) ? navigateToPerson(m) : null"
          >
            <template #prepend>
              <img :src="m.thumbnailUrl('tile_160')" :alt="m.Name" class="meta-person__avatar" />
            </template>
            <v-combobox
              v-if="isEditingPerson(m)"
              v-model:search="m.Name"
              :items="$config.values.people"
              item-title="Name"
              item-value="Name"
              return-object
              hide-no-data
              hide-details
              single-line
              append-icon=""
              density="compact"
              variant="plain"
              :placeholder="$gettext('Name')"
              :menu-props="personMenuProps"
              class="meta-inline-edit meta-person-edit"
              @update:model-value="(person) => onSelectPerson(m, person)"
              @keydown.enter.prevent="confirmField"
              @keydown.escape.prevent="cancelEditing"
              @blur="cancelEditing"
            ></v-combobox>
            <v-list-item-title v-else-if="m.Name" class="meta-person__name">{{ m.Name }}</v-list-item-title>
            <template v-if="isEditable" #append>
              <v-icon
                v-if="isEditingPerson(m)"
                icon="mdi-check"
                size="small"
                class="meta-inline-confirm"
                @mousedown.prevent
                @click.stop="confirmField"
              ></v-icon>
              <v-icon
                v-else-if="m.SubjUID"
                icon="mdi-eject"
                size="small"
                class="meta-inline-pencil"
                :title="$gettext('Detach')"
                @click.stop="onEjectPerson(m)"
              ></v-icon>
              <v-icon v-else icon="mdi-pencil-outline" size="small" class="meta-inline-pencil" @click.stop="startEditingPerson(m)"></v-icon>
            </template>
          </v-list-item>
        </template>

        <template v-if="editingField === 'labels' || labels.length > 0 || isEditable">
          <v-divider class="my-4"></v-divider>
          <v-list-item class="metadata__item">
            <div class="text-subtitle-2">{{ $gettext("Labels") }}</div>
            <template v-if="isEditable" #append>
              <v-icon
                v-if="editingField === 'labels'"
                icon="mdi-check"
                size="small"
                class="meta-inline-confirm"
                @mousedown.prevent
                @click.stop="confirmLabels"
              ></v-icon>
              <v-icon v-else icon="mdi-pencil-outline" size="small" class="meta-inline-pencil" @click.stop="startChipEditing('labels')"></v-icon>
            </template>
          </v-list-item>
          <v-list-item v-if="labels.length > 0 || editingField === 'labels'" class="metadata__item metadata__chips">
            <div class="d-flex flex-wrap ga-1">
              <span
                v-for="l in labels"
                :key="l.Label.UID"
                class="meta-chip meta-chip--primary"
                :class="{ 'meta-chip--pending-remove': isLabelPendingRemoval(l) }"
                @click.stop.prevent="editingField !== 'labels' ? navigateToLabel(l.Label) : toggleLabelRemoval(l)"
              >
                {{ l.Label.Name }}
                <v-icon
                  v-if="editingField === 'labels'"
                  :icon="isLabelPendingRemoval(l) ? 'mdi-undo' : 'mdi-close-circle'"
                  size="x-small"
                  class="ml-1"
                ></v-icon>
              </span>
              <span
                v-for="name in pendingLabelAdditions"
                :key="'add-' + name"
                class="meta-chip meta-chip--pending-add"
                @click.stop.prevent="removePendingLabelAdd(name)"
              >
                {{ name }}
                <v-icon icon="mdi-close-circle" size="x-small" class="ml-1"></v-icon>
              </span>
            </div>
          </v-list-item>
          <v-list-item v-else-if="isEditable" class="metadata__item">
            <div class="meta-add-prompt" @click.stop="startChipEditing('labels')">{{ $gettext("Add label") }}</div>
          </v-list-item>
          <v-list-item v-if="editingField === 'labels'" class="metadata__item">
            <v-combobox
              :key="chipKey"
              v-model="chipInput"
              v-model:search="chipSearch"
              :items="labelOptions"
              item-title="Name"
              item-value="Name"
              return-object
              :placeholder="$gettext('Add label')"
              variant="plain"
              density="compact"
              hide-details
              hide-no-data
              append-icon=""
              autocomplete="off"
              class="meta-inline-edit"
              @update:model-value="onLabelSelected"
              @keydown.enter.prevent="onLabelEnter"
            ></v-combobox>
          </v-list-item>
        </template>

        <template v-if="editingField === 'albums' || albums.length > 0 || isEditable">
          <v-divider class="my-4"></v-divider>
          <v-list-item class="metadata__item">
            <div class="text-subtitle-2">{{ $gettext("Albums") }}</div>
            <template v-if="isEditable" #append>
              <v-icon
                v-if="editingField === 'albums'"
                icon="mdi-check"
                size="small"
                class="meta-inline-confirm"
                @mousedown.prevent
                @click.stop="confirmAlbums"
              ></v-icon>
              <v-icon v-else icon="mdi-pencil-outline" size="small" class="meta-inline-pencil" @click.stop="startChipEditing('albums')"></v-icon>
            </template>
          </v-list-item>
          <v-list-item v-if="albums.length > 0 || editingField === 'albums'" class="metadata__item metadata__chips">
            <div class="d-flex flex-wrap ga-1">
              <span
                v-for="a in albums"
                :key="a.UID"
                class="meta-chip meta-chip--primary"
                :class="{ 'meta-chip--pending-remove': isAlbumPendingRemoval(a) }"
                @click.stop.prevent="editingField !== 'albums' ? navigateToAlbum(a) : toggleAlbumRemoval(a)"
              >
                {{ a.Title }}
                <v-icon
                  v-if="editingField === 'albums'"
                  :icon="isAlbumPendingRemoval(a) ? 'mdi-undo' : 'mdi-close-circle'"
                  size="x-small"
                  class="ml-1"
                ></v-icon>
              </span>
              <span
                v-for="a in pendingAlbumAdditions"
                :key="'add-' + a.UID"
                class="meta-chip meta-chip--pending-add"
                @click.stop.prevent="removePendingAlbumAdd(a)"
              >
                {{ a.Title }}
                <v-icon icon="mdi-close-circle" size="x-small" class="ml-1"></v-icon>
              </span>
            </div>
          </v-list-item>
          <v-list-item v-else-if="isEditable" class="metadata__item">
            <div class="meta-add-prompt" @click.stop="startChipEditing('albums')">{{ $gettext("Add to album") }}</div>
          </v-list-item>
          <v-list-item v-if="editingField === 'albums'" class="metadata__item">
            <v-autocomplete
              :key="chipKey"
              v-model="chipInput"
              v-model:search="chipSearch"
              :items="albumOptions"
              item-title="Title"
              item-value="UID"
              :placeholder="$gettext('Add to album')"
              variant="plain"
              density="compact"
              hide-details
              hide-no-data
              return-object
              append-icon=""
              autocomplete="off"
              class="meta-inline-edit"
              @update:model-value="onAlbumSelected"
              @keydown.enter="onAlbumEnter"
            ></v-autocomplete>
          </v-list-item>
        </template>

        <template
          v-if="
            editingField === 'subject' ||
            editingField === 'artist' ||
            editingField === 'copyright' ||
            editingField === 'license' ||
            subject ||
            artist ||
            copyright ||
            license ||
            isEditable
          "
        >
          <v-divider class="my-4"></v-divider>

          <!-- Subject -->
          <v-list-item v-if="editingField === 'subject' || subject || isEditable" v-tooltip="$gettext('Subject')" prepend-icon="mdi-text-box-outline" class="metadata__item">
            <v-textarea
              v-if="editingField === 'subject'"
              v-model="p.Details.Subject"
              :placeholder="$gettext('Subject')"
              :rules="[textRule]"
              variant="plain"
              density="compact"
              auto-grow
              hide-details="auto"
              autocomplete="off"
              class="meta-inline-edit"
              @keydown.escape.prevent="cancelEditing"
              @blur="cancelEditing"
            ></v-textarea>
            <div v-else-if="subject" class="text-body-2 meta-scrollable">{{ subject }}</div>
            <div v-else class="meta-add-prompt" @click.stop="startEditing('subject')">{{ $gettext("Subject") }}</div>
            <template v-if="isEditable" #append>
              <v-icon
                v-if="editingField === 'subject'"
                icon="mdi-check"
                size="small"
                class="meta-inline-confirm"
                @mousedown.prevent
                @click.stop="confirmField"
              ></v-icon>
              <v-icon v-else icon="mdi-pencil-outline" size="small" class="meta-inline-pencil" @click.stop="startEditing('subject')"></v-icon>
            </template>
          </v-list-item>

          <!-- Artist -->
          <v-list-item v-if="editingField === 'artist' || artist || isEditable" v-tooltip="$gettext('Artist')" prepend-icon="mdi-palette" class="metadata__item">
            <v-textarea
              v-if="editingField === 'artist'"
              v-model="p.Details.Artist"
              :placeholder="$gettext('Artist')"
              :rules="[textRule]"
              variant="plain"
              density="compact"
              auto-grow
              hide-details="auto"
              autocomplete="off"
              class="meta-inline-edit"
              @keydown.escape.prevent="cancelEditing"
              @blur="cancelEditing"
            ></v-textarea>
            <div v-else-if="artist" class="text-body-2 meta-scrollable">{{ artist }}</div>
            <div v-else class="meta-add-prompt" @click.stop="startEditing('artist')">{{ $gettext("Artist") }}</div>
            <template v-if="isEditable" #append>
              <v-icon
                v-if="editingField === 'artist'"
                icon="mdi-check"
                size="small"
                class="meta-inline-confirm"
                @mousedown.prevent
                @click.stop="confirmField"
              ></v-icon>
              <v-icon v-else icon="mdi-pencil-outline" size="small" class="meta-inline-pencil" @click.stop="startEditing('artist')"></v-icon>
            </template>
          </v-list-item>

          <!-- Copyright -->
          <v-list-item v-if="editingField === 'copyright' || copyright || isEditable" v-tooltip="$gettext('Copyright')" prepend-icon="mdi-copyright" class="metadata__item">
            <v-textarea
              v-if="editingField === 'copyright'"
              v-model="p.Details.Copyright"
              :placeholder="$gettext('Copyright')"
              :rules="[textRule]"
              variant="plain"
              density="compact"
              auto-grow
              hide-details="auto"
              autocomplete="off"
              class="meta-inline-edit"
              @keydown.escape.prevent="cancelEditing"
              @blur="cancelEditing"
            ></v-textarea>
            <div v-else-if="copyright" class="text-body-2 meta-scrollable">{{ copyright }}</div>
            <div v-else class="meta-add-prompt" @click.stop="startEditing('copyright')">{{ $gettext("Copyright") }}</div>
            <template v-if="isEditable" #append>
              <v-icon
                v-if="editingField === 'copyright'"
                icon="mdi-check"
                size="small"
                class="meta-inline-confirm"
                @mousedown.prevent
                @click.stop="confirmField"
              ></v-icon>
              <v-icon v-else icon="mdi-pencil-outline" size="small" class="meta-inline-pencil" @click.stop="startEditing('copyright')"></v-icon>
            </template>
          </v-list-item>

          <!-- License -->
          <v-list-item v-if="editingField === 'license' || license || isEditable" v-tooltip="$gettext('License')" prepend-icon="mdi-license" class="metadata__item">
            <v-textarea
              v-if="editingField === 'license'"
              v-model="p.Details.License"
              :placeholder="$gettext('License')"
              :rules="[textRule]"
              variant="plain"
              density="compact"
              auto-grow
              hide-details="auto"
              autocomplete="off"
              class="meta-inline-edit"
              @keydown.escape.prevent="cancelEditing"
              @blur="cancelEditing"
            ></v-textarea>
            <div v-else-if="license" class="text-body-2 meta-scrollable">{{ license }}</div>
            <div v-else class="meta-add-prompt" @click.stop="startEditing('license')">{{ $gettext("License") }}</div>
            <template v-if="isEditable" #append>
              <v-icon
                v-if="editingField === 'license'"
                icon="mdi-check"
                size="small"
                class="meta-inline-confirm"
                @mousedown.prevent
                @click.stop="confirmField"
              ></v-icon>
              <v-icon v-else icon="mdi-pencil-outline" size="small" class="meta-inline-pencil" @click.stop="startEditing('license')"></v-icon>
            </template>
          </v-list-item>
        </template>

        <template v-if="editingField === 'keywords' || keywords || isEditable">
          <v-divider class="my-4"></v-divider>
          <v-list-item class="metadata__item">
            <div class="text-subtitle-2">{{ $gettext("Keywords") }}</div>
            <template v-if="isEditable" #append>
              <v-icon
                v-if="editingField === 'keywords'"
                icon="mdi-check"
                size="small"
                class="meta-inline-confirm"
                @mousedown.prevent
                @click.stop="confirmField"
              ></v-icon>
              <v-icon v-else icon="mdi-pencil-outline" size="small" class="meta-inline-pencil" @click.stop="startEditing('keywords')"></v-icon>
            </template>
          </v-list-item>
          <v-list-item class="metadata__item">
            <v-textarea
              v-if="editingField === 'keywords'"
              v-model="p.Details.Keywords"
              :placeholder="$gettext('Keywords')"
              variant="plain"
              density="compact"
              auto-grow
              hide-details="auto"
              autocomplete="off"
              class="meta-inline-edit"
              @keydown.escape.prevent="cancelEditing"
              @blur="cancelEditing"
            ></v-textarea>
            <div v-else-if="keywords" class="text-body-2 meta-keywords meta-scrollable">{{ keywords }}</div>
            <div v-else class="meta-add-prompt" @click.stop="startEditing('keywords')">{{ $gettext("Keywords") }}</div>
          </v-list-item>
        </template>

        <template v-if="editingField === 'originalname' || originalName || isEditable">
          <v-divider class="my-4"></v-divider>
          <v-list-item v-tooltip="$gettext('Original Name')" prepend-icon="mdi-file-outline" class="metadata__item">
            <v-text-field
              v-if="editingField === 'originalname'"
              v-model="p.OriginalName"
              :placeholder="$gettext('Original Name')"
              variant="plain"
              density="compact"
              hide-details="auto"
              autocomplete="off"
              class="meta-inline-edit"
              @keydown.enter.prevent="confirmField"
              @keydown.escape.prevent="cancelEditing"
              @blur="cancelEditing"
            ></v-text-field>
            <div v-else-if="originalName" class="text-body-2">{{ originalName }}</div>
            <div v-else class="meta-add-prompt" @click.stop="startEditing('originalname')">{{ $gettext("Original Name") }}</div>
            <template v-if="isEditable" #append>
              <v-icon
                v-if="editingField === 'originalname'"
                icon="mdi-check"
                size="small"
                class="meta-inline-confirm"
                @mousedown.prevent
                @click.stop="confirmField"
              ></v-icon>
              <v-icon v-else icon="mdi-pencil-outline" size="small" class="meta-inline-pencil" @click.stop="startEditing('originalname')"></v-icon>
            </template>
          </v-list-item>
        </template>

        <template v-if="editingField === 'notes' || notesHtml || isEditable">
          <v-divider class="my-4"></v-divider>
          <v-list-item class="metadata__item">
            <div class="text-subtitle-2">{{ $gettext("Notes") }}</div>
            <template v-if="isEditable" #append>
              <v-icon
                v-if="editingField === 'notes'"
                icon="mdi-check"
                size="small"
                class="meta-inline-confirm"
                @mousedown.prevent
                @click.stop="confirmField"
              ></v-icon>
              <v-icon v-else icon="mdi-pencil-outline" size="small" class="meta-inline-pencil" @click.stop="startEditing('notes')"></v-icon>
            </template>
          </v-list-item>
          <v-list-item class="metadata__item">
            <v-textarea
              v-if="editingField === 'notes'"
              v-model="p.Details.Notes"
              :placeholder="$gettext('Notes')"
              variant="plain"
              density="compact"
              auto-grow
              hide-details="auto"
              autocomplete="off"
              class="meta-inline-edit"
              @keydown.escape.prevent="cancelEditing"
              @blur="cancelEditing"
            ></v-textarea>
            <div v-else-if="notesHtml" class="text-body-2 meta-notes meta-scrollable" v-html="notesHtml"></div>
            <div v-else class="meta-add-prompt" @click.stop="startEditing('notes')">{{ $gettext("Notes") }}</div>
          </v-list-item>
        </template>
      </v-list>
    </div>
    <p-date-time-dialog :visible="dateTimeDialog" :photo="photo" @close="dateTimeDialog = false" @confirm="confirmDateTime"></p-date-time-dialog>
    <p-camera-dialog :visible="cameraDialog" :photo="photo" @close="cameraDialog = false" @confirm="confirmCamera"></p-camera-dialog>
    <p-location-dialog
      :visible="locationDialog"
      :latlng="[photo ? Number(photo.Lat) || 0 : 0, photo ? Number(photo.Lng) || 0 : 0]"
      @close="locationDialog = false"
      @confirm="confirmLocation"
    ></p-location-dialog>
  </div>
</template>

<script>
import { DateTime } from "luxon";
import * as formats from "options/formats";

import { Photo } from "model/photo";
import { Label } from "model/label";
import { Album } from "model/album";
import PMap from "component/map.vue";
import PDateTimeDialog from "component/sidebar/datetime-dialog.vue";
import PCameraDialog from "component/sidebar/camera-dialog.vue";
import PLocationDialog from "component/location/dialog.vue";

export default {
  name: "PSidebarInfo",
  components: {
    PMap,
    PDateTimeDialog,
    PCameraDialog,
    PLocationDialog,
  },
  props: {
    modelValue: {
      type: Object,
      default: () => {},
    },
    photo: {
      type: Object,
      default: null,
    },
    canEdit: {
      type: Boolean,
      default: false,
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
  emits: ["update:modelValue", "close", "navigate"],
  data() {
    return {
      actions: [],
      featPlaces: this.$config.feature("places"),
      textRule: (v) => !v || v.length <= this.$config.get("clip") || this.$gettext("Text too long"),
      dateTimeDialog: false,
      cameraDialog: false,
      locationDialog: false,
      editingField: null,
      editOriginal: null,
      editMarker: null,
      chipInput: null,
      chipSearch: "",
      chipKey: 0,
      labelOptions: [],
      albumOptions: [],
      pendingLabelRemovals: [],
      pendingLabelAdditions: [],
      pendingAlbumRemovals: [],
      pendingAlbumAdditions: [],
      personMenuProps: {
        openOnClick: false,
        openOnFocus: true,
        closeOnContentClick: true,
        maxHeight: 200,
        density: "compact",
      },
    };
  },
  computed: {
    model() {
      return this.modelValue;
    },
    p() {
      return this.photo;
    },
    isEditable() {
      return this.canEdit && this.p && this.p.Details;
    },
    captionHtml() {
      const raw = this.p?.Caption ?? this.model?.Caption;
      if (!raw) return "";
      return this.$util.sanitizeHtml(this.$util.encodeHTML(raw));
    },
    notesHtml() {
      if (!this.p?.Details?.Notes) return "";
      return this.$util.sanitizeHtml(this.$util.encodeHTML(this.p.Details.Notes));
    },
    cameraInfo() {
      if (!this.p) return "";
      const info = this.p.getCameraInfo();
      return info !== this.$gettext("Unknown") ? info : "";
    },
    lensInfo() {
      if (!this.p) return "";
      const info = this.p.getLensInfo();
      return info !== this.$gettext("Unknown") ? info : "";
    },
    exifInfo() {
      if (!this.p) return "";
      return this.p.getExifInfo();
    },
    people() {
      if (!this.p) return [];
      return this.p.getMarkers(true);
    },
    labels() {
      if (!this.p?.Labels) return [];
      return this.p.Labels.filter((l) => l.Label && l.Label.Name && l.Uncertainty < 100);
    },
    albums() {
      if (!this.p?.Albums) return [];
      return this.p.Albums.filter((a) => a.Title);
    },
    subject() {
      return this.p?.Details?.Subject || "";
    },
    artist() {
      return this.p?.Details?.Artist || "";
    },
    copyright() {
      return this.p?.Details?.Copyright || "";
    },
    license() {
      return this.p?.Details?.License || "";
    },
    keywords() {
      return this.p?.Details?.Keywords || "";
    },
    placeName() {
      if (!this.p) return "";
      return this.p.locationInfo() || "";
    },
    altitude() {
      if (!this.p || !this.p.Altitude) return "";
      return this.p.Altitude + " m";
    },
    originalName() {
      if (!this.p) return "";
      return this.p.OriginalName || "";
    },
  },
  watch: {
    photo() {
      if (this.editingField) {
        this.cancelEditing();
      }
    },
  },
  methods: {
    close() {
      this.$emit("close");
    },
    getFieldValue(field) {
      switch (field) {
        case "title":
          return this.p.Title;
        case "caption":
          return this.p.Caption;
        case "subject":
          return this.p.Details.Subject;
        case "artist":
          return this.p.Details.Artist;
        case "copyright":
          return this.p.Details.Copyright;
        case "license":
          return this.p.Details.License;
        case "keywords":
          return this.p.Details.Keywords;
        case "notes":
          return this.p.Details.Notes;
        case "originalname":
          return this.p.OriginalName;
        default:
          return "";
      }
    },
    setFieldValue(field, value) {
      switch (field) {
        case "title":
          this.p.Title = value;
          break;
        case "caption":
          this.p.Caption = value;
          break;
        case "subject":
          this.p.Details.Subject = value;
          break;
        case "artist":
          this.p.Details.Artist = value;
          break;
        case "copyright":
          this.p.Details.Copyright = value;
          break;
        case "license":
          this.p.Details.License = value;
          break;
        case "keywords":
          this.p.Details.Keywords = value;
          break;
        case "notes":
          this.p.Details.Notes = value;
          break;
        case "originalname":
          this.p.OriginalName = value;
          break;
      }
    },
    startEditing(field) {
      if (this.editingField) {
        this.cancelEditing();
      }

      this.editingField = field;
      this.editOriginal = this.getFieldValue(field);
      this._editStartedAt = Date.now();

      this.$nextTick(() => {
        const input = this.$el.querySelector(".meta-inline-edit input, .meta-inline-edit textarea");
        if (input) input.focus();
      });
    },
    isEditingPerson(marker) {
      return this.editingField === "person-" + (marker.UID || marker.CropID);
    },
    startEditingPerson(marker) {
      if (this.editingField) {
        this.cancelEditing();
      }

      this.editingField = "person-" + (marker.UID || marker.CropID);
      this.editOriginal = marker.Name;
      this.editMarker = marker;
      this._editStartedAt = Date.now();

      this.$nextTick(() => {
        const input = this.$el.querySelector(".meta-person-edit input");
        if (input) input.focus();
      });
    },
    onEjectPerson(marker) {
      marker.clearSubject().then(() => {
        this.startEditingPerson(marker);
      });
    },
    onSelectPerson(marker, person) {
      if (typeof person === "object" && person?.Name && person?.UID) {
        marker.Name = person.Name;
        marker.SubjUID = person.UID;

        this.editingField = null;
        this.editOriginal = null;
        this.editMarker = null;

        marker.setName();
      }
    },
    confirmField() {
      if (this.editMarker) {
        const marker = this.editMarker;
        this.editingField = null;
        this.editOriginal = null;
        this.editMarker = null;

        if (!marker.Name || !marker.Name.trim()) {
          if (marker.SubjUID) {
            marker.clearSubject();
          }

          return;
        }

        // Match typed name against existing people (case-insensitive).
        const people = this.$config.values?.people;

        if (people) {
          const found = people.find((p) => p.Name.localeCompare(marker.Name, "en", { sensitivity: "base" }) === 0);

          if (found) {
            marker.Name = found.Name;
            marker.SubjUID = found.UID;
          }
        }

        marker.setName();
        return;
      }

      if (!this.p || !this.canEdit) {
        this.editingField = null;
        return;
      }

      const field = this.editingField;
      this.editingField = null;
      this.editOriginal = null;

      if (!this.p.wasChanged()) {
        return;
      }

      this.p.update().then(() => {
        if (field === "title" || field === "caption") {
          this.model.Title = this.p.Title;
          this.model.Caption = this.p.Caption;
        }
      });
    },
    cancelEditing() {
      // Guard: clicking a pencil icon triggers blur on the previous field before focus lands
      // on the new input, which would immediately cancel the edit we just started.
      if (this._editStartedAt && Date.now() - this._editStartedAt < 200) {
        return;
      }

      if (this.editingField && this.editOriginal !== null) {
        if (this.editMarker) {
          this.editMarker.Name = this.editOriginal;
        } else {
          this.setFieldValue(this.editingField, this.editOriginal);
        }
      }

      this.editingField = null;
      this.editOriginal = null;
      this.editMarker = null;
      this._editStartedAt = null;
      this.pendingLabelRemovals = [];
      this.pendingLabelAdditions = [];
      this.pendingAlbumRemovals = [];
      this.pendingAlbumAdditions = [];
    },
    formatTime(model) {
      if (!model || !model.TakenAtLocal) {
        return this.$gettext("Unknown");
      }

      // Always parse as UTC to avoid time shifts
      const dateTime = DateTime.fromISO(model.TakenAtLocal, { zone: "UTC" });

      if (model.TimeZone && model.TimeZone !== "Local" && model.TimeZone !== "UTC") {
        // We use the real timezone just for display, but don't shift the time (prevents double timezone offset as backend already applied it)
        return dateTime.setZone(model.TimeZone, { keepLocalTime: true }).toLocaleString(formats.DATETIME_MED_TZ);
      } else {
        return dateTime.toLocaleString(formats.DATETIME_MED);
      }
    },
    navigateToLabel(label) {
      const slug = label.CustomSlug || label.Slug;
      this.$emit("navigate", { name: "browse", query: { q: "label:" + slug } });
    },
    navigateToAlbum(album) {
      this.$emit("navigate", { name: "album", params: { album: album.UID, slug: "view" } });
    },
    navigateToPerson(marker) {
      if (marker.SubjUID) {
        this.$emit("navigate", { name: "browse", query: { q: "subject:" + marker.SubjUID } });
      } else if (marker.Name) {
        this.$emit("navigate", { name: "browse", query: { q: "person:" + marker.Name } });
      }
    },
    startChipEditing(field) {
      if (this.editingField) {
        this.cancelEditing();
      }

      this.editingField = field;
      this._editStartedAt = Date.now();

      if (field === "labels" && !this.labelOptions.length) {
        Label.search({ count: 1000, order: "name", all: true })
          .then((resp) => {
            this.labelOptions = (resp.models || []).map((l) => ({ Name: l.Name, UID: l.UID }));
          })
          .catch(() => {});
      } else if (field === "albums" && !this.albumOptions.length) {
        Album.search({ count: 1000, order: "name", type: "album" })
          .then((resp) => {
            this.albumOptions = resp.models || [];
          })
          .catch(() => {});
      }
    },
    clearChipInput() {
      this.chipInput = null;
      this.chipSearch = "";
      this.chipKey++;
    },
    onLabelSelected(value) {
      if (!value || typeof value !== "object" || !value.Name) return;

      const name = value.Name.trim();

      if (name && !this.pendingLabelAdditions.includes(name) && !this.labels.some((l) => l.Label.Name.toLowerCase() === name.toLowerCase())) {
        this.pendingLabelAdditions.push(name);
      }

      this.clearChipInput();
    },
    onLabelEnter() {
      const name = (this.chipSearch || "").trim();
      if (!name) return;
      if (this.pendingLabelAdditions.includes(name)) return;
      if (this.labels.some((l) => l.Label.Name.toLowerCase() === name.toLowerCase())) return;
      this.pendingLabelAdditions.push(name);
      this.clearChipInput();
    },
    removePendingLabelAdd(name) {
      const idx = this.pendingLabelAdditions.indexOf(name);
      if (idx >= 0) this.pendingLabelAdditions.splice(idx, 1);
    },
    isLabelPendingRemoval(label) {
      return this.pendingLabelRemovals.includes(label.Label.ID);
    },
    toggleLabelRemoval(label) {
      if (!label?.Label?.ID) return;
      const id = label.Label.ID;
      const idx = this.pendingLabelRemovals.indexOf(id);
      if (idx >= 0) {
        this.pendingLabelRemovals.splice(idx, 1);
      } else {
        this.pendingLabelRemovals.push(id);
      }
    },
    confirmLabels() {
      if (!this.p) {
        this.editingField = null;
        return;
      }

      const removals = this.pendingLabelRemovals.slice();
      const additions = this.pendingLabelAdditions.slice();
      this.editingField = null;
      this.pendingLabelRemovals = [];
      this.pendingLabelAdditions = [];

      const promises = [];
      removals.forEach((id) => promises.push(this.p.removeLabel(id)));
      additions.forEach((name) => promises.push(this.p.addLabel(name)));

      if (promises.length) {
        Promise.all(promises).catch(() => {});
      }
    },
    isAlbumPendingRemoval(album) {
      return this.pendingAlbumRemovals.includes(album.UID);
    },
    toggleAlbumRemoval(album) {
      if (!album?.UID) return;
      const uid = album.UID;
      const idx = this.pendingAlbumRemovals.indexOf(uid);
      if (idx >= 0) {
        this.pendingAlbumRemovals.splice(idx, 1);
      } else {
        this.pendingAlbumRemovals.push(uid);
      }
    },
    confirmAlbums() {
      if (!this.p) {
        this.editingField = null;
        return;
      }

      const removals = this.pendingAlbumRemovals.slice();
      const additions = this.pendingAlbumAdditions.slice();
      this.editingField = null;
      this.pendingAlbumRemovals = [];
      this.pendingAlbumAdditions = [];

      const promises = [];
      removals.forEach((uid) => promises.push(this.$api.delete(`albums/${uid}/photos`, { data: { photos: [this.p.UID] } })));
      additions.forEach((a) => promises.push(this.$api.post(`albums/${a.UID}/photos`, { photos: [this.p.UID] })));

      if (promises.length) {
        Promise.all(promises)
          .then(() => {
            Photo.evictCache(this.p.UID);
            return this.p.find(this.p.UID);
          })
          .then((photo) => {
            this.p.setValues(photo.getValues());
          })
          .catch(() => {});
      }
    },
    onAlbumSelected(value) {
      if (!value || typeof value !== "object" || !value.UID) {
        this.clearChipInput();
        return;
      }

      if (!this.pendingAlbumAdditions.find((a) => a.UID === value.UID) && !this.albums.some((a) => a.UID === value.UID)) {
        this.pendingAlbumAdditions.push(value);
      }

      this.clearChipInput();
    },
    onAlbumEnter() {
      const search = (this.chipSearch || "").trim();
      if (!search) return;

      const lower = search.toLowerCase();
      const match =
        this.albumOptions.find((a) => a.Title.toLowerCase().startsWith(lower)) || this.albumOptions.find((a) => a.Title.toLowerCase().includes(lower));

      if (match) {
        this.onAlbumSelected(match);
        return;
      }

      // Create new album with the typed name.
      if (this.pendingAlbumAdditions.find((a) => a.Title.toLowerCase() === lower)) return;

      const album = new Album({ Title: search });

      album
        .save()
        .then(() => {
          if (album.UID) {
            this.pendingAlbumAdditions.push(album);
            this.albumOptions.push(album);
          }
        })
        .catch(() => {})
        .finally(() => {
          this.clearChipInput();
        });
    },
    removePendingAlbumAdd(album) {
      const idx = this.pendingAlbumAdditions.findIndex((a) => a.UID === album.UID);
      if (idx >= 0) this.pendingAlbumAdditions.splice(idx, 1);
    },
    confirmDateTime(data) {
      this.dateTimeDialog = false;

      if (!this.photo || !this.canEdit) return;

      this.photo.Day = data.Day;
      this.photo.Month = data.Month;
      this.photo.Year = data.Year;
      this.photo.TimeZone = data.TimeZone;

      this.photo.TakenAtLocal = this.photo.localDateString(data.time);

      if (this.photo.currentTimeZoneUTC()) {
        this.photo.TakenAt = this.photo.TakenAtLocal;
      }

      this.photo.update().then(() => {
        this.model.TakenAtLocal = this.photo.TakenAtLocal;
        this.model.TimeZone = this.photo.TimeZone;
      });
    },
    confirmCamera(data) {
      this.cameraDialog = false;

      if (!this.photo || !this.canEdit) return;

      this.photo.CameraID = data.CameraID;
      this.photo.LensID = data.LensID;
      this.photo.Iso = data.Iso;
      this.photo.Exposure = data.Exposure;
      this.photo.FNumber = data.FNumber;
      this.photo.FocalLength = data.FocalLength;

      this.photo.update();
    },
    confirmLocation(data) {
      this.locationDialog = false;

      if (!this.photo || !this.canEdit) return;

      this.photo.Lat = data.lat;
      this.photo.Lng = data.lng;
      this.photo.PlaceSrc = "manual";

      if (data.location?.country) {
        this.photo.Country = data.location.country;
      }

      this.photo.update().then(() => {
        this.model.Lat = this.photo.Lat;
        this.model.Lng = this.photo.Lng;
      });
    },
  },
};
</script>
