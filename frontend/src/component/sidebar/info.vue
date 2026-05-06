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
            :ref="setInlineEditorRef"
            v-model="photo.Title"
            :placeholder="$pgettext('Photo', 'Title')"
            :rules="[textRule]"
            variant="plain"
            density="compact"
            hide-details="auto"
            autocomplete="off"
            class="meta-inline-edit meta-inline-title"
            @keydown.enter.prevent="confirmField"
            @keydown.escape.prevent="cancelEditing"
            @blur="onInlineFieldBlur"
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
            :ref="setInlineEditorRef"
            v-model="photo.Caption"
            :placeholder="$gettext('Caption')"
            variant="plain"
            density="compact"
            auto-grow
            :max-rows="6"
            hide-details="auto"
            autocomplete="off"
            class="meta-inline-edit meta-inline-caption"
            @keydown.escape.prevent="cancelEditing"
            @blur="onInlineFieldBlur"
          ></v-textarea>
          <!-- eslint-disable-next-line vue/no-v-html -- captionHtml is encode-then-sanitized via $util.sanitizeHtml($util.encodeHTML(raw)); see captionHtml() computed -->
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

        <v-list-item v-if="!restrictedRole && fileName" class="metadata__item metadata__file">
          <div class="meta-filename" :title="fileName">{{ fileName }}</div>
        </v-list-item>

        <v-list-item v-if="fileInfo" v-tooltip="$gettext('File')" :title="fileInfo" :prepend-icon="fileIcon" class="metadata__item"></v-list-item>

        <v-divider v-if="(!restrictedRole && fileName) || fileInfo" class="my-4"></v-divider>

        <v-list-item v-tooltip="$gettext(`Taken`)" :title="formatTime(model)" prepend-icon="mdi-calendar" class="metadata__item">
          <template v-if="isEditable" #append>
            <v-icon icon="mdi-pencil-outline" size="small" class="meta-inline-pencil" @click.stop="dateTimeDialog = true"></v-icon>
          </template>
        </v-list-item>

        <v-list-item
          v-if="!restrictedRole && (cameraInfo || isEditable)"
          v-tooltip="$gettext('Camera')"
          :title="cameraInfo || $gettext('Unknown')"
          prepend-icon="mdi-camera"
          class="metadata__item"
        >
          <template v-if="isEditable" #append>
            <v-icon icon="mdi-pencil-outline" size="small" class="meta-inline-pencil" @click.stop="cameraDialog = true"></v-icon>
          </template>
        </v-list-item>

        <v-list-item v-if="!restrictedRole && lensInfo" v-tooltip="$gettext('Lens')" :title="lensInfo" prepend-icon="mdi-camera-iris" class="metadata__item">
        </v-list-item>

        <template v-if="(model.Lat && model.Lng) || (!restrictedRole && isEditable && featPlaces)">
          <v-divider class="my-4"></v-divider>
          <v-list-item
            v-if="!restrictedRole && (placeName || !(model.Lat && model.Lng))"
            v-tooltip="$gettext('Location')"
            :title="placeName || $gettext('Unknown')"
            prepend-icon="mdi-map-marker"
            class="metadata__item"
          >
            <template v-if="isEditable && featPlaces && !(model.Lat && model.Lng)" #append>
              <v-icon
                icon="mdi-pencil-outline"
                size="small"
                class="meta-inline-pencil meta-inline-pencil--location"
                @click.stop.prevent="locationDialog = true"
              ></v-icon>
            </template>
          </v-list-item>
          <template v-if="model.Lat && model.Lng">
            <v-list-item
              v-tooltip="$gettext(`Coordinates`)"
              :title="altitude && !restrictedRole ? model.getLatLng() + ' \u00b7 ' + altitude : model.getLatLng()"
              class="clickable metadata__item"
              @click.stop="model.copyLatLng()"
            >
              <template v-if="isEditable && featPlaces" #append>
                <v-icon
                  icon="mdi-pencil-outline"
                  size="small"
                  class="meta-inline-pencil meta-inline-pencil--location"
                  @click.stop.prevent="locationDialog = true"
                ></v-icon>
              </template>
            </v-list-item>
            <v-list-item v-if="featPlaces" class="mx-0 px-0">
              <p-map :latlng="[model.Lat, model.Lng]" :animate-duration="0"></p-map>
            </v-list-item>
          </template>
        </template>

        <template v-if="!restrictedRole && featPeople && (people.length > 0 || isEditable)">
          <v-divider class="my-4"></v-divider>
          <v-list-item class="metadata__item">
            <div class="text-subtitle-2">{{ $gettext("People") }}</div>
            <template v-if="isEditable" #append>
              <v-icon
                :icon="markersVisible ? 'mdi-eye-off-outline' : 'mdi-eye-outline'"
                size="small"
                class="meta-markers-toggle"
                :class="{ 'is-active': markersVisible }"
                :title="markersVisible ? $gettext('Hide face markers') : $gettext('Show face markers')"
                :disabled="markersBusy"
                @mousedown.prevent
                @click.stop="onToggleMarkersVisible"
              ></v-icon>
              <v-icon
                :icon="addingMarker ? 'mdi-check' : 'mdi-plus'"
                size="small"
                class="meta-marker-add"
                :class="{ 'is-active': addingMarker }"
                :title="addingMarker ? $gettext('Done') : $gettext('Add face')"
                :disabled="markersBusy"
                @mousedown.prevent
                @click.stop="onToggleAddingMarker"
              ></v-icon>
            </template>
          </v-list-item>
          <v-list-item v-for="m in people" :key="m.UID || m.CropID" :data-marker-uid="m.UID" class="metadata__item metadata__person-row">
            <template #prepend>
              <img
                :src="m.thumbnailUrl('tile_160')"
                :alt="m.Name"
                class="meta-person__avatar"
                :class="{ clickable: m.Name && m.SubjUID }"
                @click.stop="m.Name && m.SubjUID ? navigateToPerson(m) : null"
              />
            </template>
            <v-combobox
              v-if="isEditable"
              :model-value="markerInputValue(m.UID)"
              :search="markerInputSearch(m.UID)"
              :items="knownPeople"
              item-title="Name"
              item-value="Name"
              :placeholder="$gettext('Name')"
              :menu-props="markerMenuProps"
              :readonly="markersBusy || !!m.SubjUID"
              :rules="[markerNameRule]"
              return-object
              hide-no-data
              hide-details="auto"
              single-line
              open-on-clear
              append-icon=""
              autocomplete="off"
              density="compact"
              variant="plain"
              class="meta-inline-edit meta-inline-marker"
              :class="{ 'meta-inline-marker--named': m.SubjUID }"
              @update:model-value="(v) => onPickPerson(m, v)"
              @update:search="(v) => setMarkerInputValue(m.UID, v)"
              @keydown.enter.prevent="confirmMarkerName(m, 'enter')"
              @keydown.escape.prevent="cancelMarkerName(m)"
              @blur="confirmMarkerName(m, 'blur')"
              @click.stop
            >
              <template #append-inner>
                <v-icon
                  v-if="m.SubjUID"
                  icon="mdi-eject"
                  size="x-small"
                  class="meta-marker-eject"
                  :title="$gettext('Remove Name')"
                  :disabled="markersBusy"
                  @mousedown.prevent
                  @click.stop="onEjectMarker(m)"
                ></v-icon>
                <v-icon
                  v-else
                  icon="mdi-close"
                  size="x-small"
                  class="meta-marker-remove"
                  :title="$gettext('Remove')"
                  :disabled="markersBusy"
                  @mousedown.prevent
                  @click.stop="onRemoveMarker(m)"
                ></v-icon>
              </template>
            </v-combobox>
            <v-list-item-title v-else-if="m.Name" class="meta-person__name">{{ m.Name }}</v-list-item-title>
            <v-list-item-title v-else class="meta-person__name meta-person__unnamed">{{ $gettext("Unknown") }}</v-list-item-title>
          </v-list-item>
        </template>

        <template v-if="!restrictedRole && (editingField === 'labels' || labels.length > 0 || isEditable)">
          <v-divider class="my-4"></v-divider>
          <v-list-item class="metadata__item meta-labels">
            <div class="text-subtitle-2">{{ $gettext("Labels") }}</div>
            <template v-if="isEditable" #append>
              <p-sidebar-inline-toolbar :editing="editingField === 'labels'" @confirm="confirmLabels" @start="startChipEditing('labels')" />
            </template>
          </v-list-item>
          <v-list-item v-if="labels.length > 0 || chipState.labels.additions.length > 0" class="metadata__item metadata__chips meta-labels">
            <div class="d-flex flex-wrap ga-1">
              <span
                v-for="l in labels"
                :key="l.Label.UID"
                class="meta-chip meta-chip--primary"
                :class="{ 'meta-chip--pending-remove': isChipPendingRemoval('labels', l.Label.ID) }"
                @click.stop.prevent="editingField !== 'labels' ? navigateToLabel(l.Label) : togglePendingChipRemoval('labels', l.Label.ID)"
              >
                {{ l.Label.Name }}
                <v-icon
                  v-if="editingField === 'labels'"
                  :icon="isChipPendingRemoval('labels', l.Label.ID) ? 'mdi-undo' : 'mdi-close-circle'"
                  size="x-small"
                  class="ml-1"
                ></v-icon>
              </span>
              <span
                v-for="name in chipState.labels.additions"
                :key="'add-' + name"
                class="meta-chip meta-chip--pending-add"
                @click.stop.prevent="removePendingChipAdd('labels', name)"
              >
                {{ name }}
                <v-icon icon="mdi-close-circle" size="x-small" class="ml-1"></v-icon>
              </span>
            </div>
          </v-list-item>
          <v-list-item v-else-if="isEditable && editingField !== 'labels'" class="metadata__item meta-labels">
            <div class="meta-add-prompt" @click.stop="startChipEditing('labels')">{{ $gettext("Add label") }}</div>
          </v-list-item>
          <v-list-item v-if="editingField === 'labels'" class="metadata__item meta-labels">
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
              :menu-props="chipMenuProps"
              class="meta-inline-edit"
              @update:model-value="onLabelSelected"
              @keydown.enter.prevent="onLabelEnter"
            ></v-combobox>
          </v-list-item>
        </template>

        <template v-if="!restrictedRole && (editingField === 'albums' || albums.length > 0 || isEditable)">
          <v-divider class="my-4"></v-divider>
          <v-list-item class="metadata__item meta-albums">
            <div class="text-subtitle-2">{{ $gettext("Albums") }}</div>
            <template v-if="isEditable" #append>
              <p-sidebar-inline-toolbar :editing="editingField === 'albums'" @confirm="confirmAlbums" @start="startChipEditing('albums')" />
            </template>
          </v-list-item>
          <v-list-item v-if="albums.length > 0 || chipState.albums.additions.length > 0" class="metadata__item metadata__chips meta-albums">
            <div class="d-flex flex-wrap ga-1">
              <span
                v-for="a in albums"
                :key="a.UID"
                class="meta-chip meta-chip--primary"
                :class="{ 'meta-chip--pending-remove': isChipPendingRemoval('albums', a.UID) }"
                @click.stop.prevent="editingField !== 'albums' ? navigateToAlbum(a) : togglePendingChipRemoval('albums', a.UID)"
              >
                {{ a.Title }}
                <v-icon
                  v-if="editingField === 'albums'"
                  :icon="isChipPendingRemoval('albums', a.UID) ? 'mdi-undo' : 'mdi-close-circle'"
                  size="x-small"
                  class="ml-1"
                ></v-icon>
              </span>
              <span
                v-for="a in chipState.albums.additions"
                :key="'add-' + a.UID"
                class="meta-chip meta-chip--pending-add"
                @click.stop.prevent="removePendingChipAdd('albums', a.UID)"
              >
                {{ a.Title }}
                <v-icon icon="mdi-close-circle" size="x-small" class="ml-1"></v-icon>
              </span>
            </div>
          </v-list-item>
          <v-list-item v-else-if="isEditable && editingField !== 'albums'" class="metadata__item meta-albums">
            <div class="meta-add-prompt" @click.stop="startChipEditing('albums')">{{ $gettext("Add to album") }}</div>
          </v-list-item>
          <v-list-item v-if="editingField === 'albums'" class="metadata__item meta-albums">
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
              :menu-props="chipMenuProps"
              class="meta-inline-edit"
              @update:model-value="onAlbumSelected"
              @keydown.enter="onAlbumEnter"
            ></v-autocomplete>
          </v-list-item>
        </template>

        <template v-if="showDetailsSection">
          <v-divider class="my-4"></v-divider>
          <v-list-item
            v-for="f in detailsFields"
            v-show="shouldShowFieldRow(f)"
            :key="f.key"
            v-tooltip="f.label"
            :prepend-icon="f.icon"
            class="metadata__item"
            :class="`meta-${f.key}`"
          >
            <v-textarea
              v-if="editingField === f.key"
              :ref="setInlineEditorRef"
              :model-value="f.read(photo)"
              :placeholder="f.label"
              :rules="[textRule]"
              variant="plain"
              density="compact"
              auto-grow
              hide-details="auto"
              autocomplete="off"
              class="meta-inline-edit"
              :class="`meta-inline-${f.key}`"
              @update:model-value="(v) => f.write(photo, v)"
              @keydown.escape.prevent="cancelEditing"
              @blur="onInlineFieldBlur"
            ></v-textarea>
            <div v-else-if="f.read(photo)" class="text-body-2 meta-scrollable">{{ f.read(photo) }}</div>
            <div v-else class="meta-add-prompt" @click.stop="startEditing(f.key)">{{ f.label }}</div>
            <template v-if="isEditable" #append>
              <p-sidebar-inline-toolbar :editing="editingField === f.key" @confirm="confirmField" @start="startEditing(f.key)" />
            </template>
          </v-list-item>
        </template>

        <template v-for="f in textFields" :key="f.key">
          <template v-if="!restrictedRole && shouldShowFieldRow(f)">
            <v-divider class="my-4"></v-divider>
            <v-list-item class="metadata__item" :class="`meta-${f.key}`">
              <div class="text-subtitle-2">{{ f.label }}</div>
              <template v-if="isEditable" #append>
                <p-sidebar-inline-toolbar :editing="editingField === f.key" @confirm="confirmField" @start="startEditing(f.key)" />
              </template>
            </v-list-item>
            <v-list-item class="metadata__item" :class="`meta-${f.key}`">
              <v-textarea
                v-if="editingField === f.key"
                :ref="setInlineEditorRef"
                :model-value="f.read(photo)"
                :placeholder="f.label"
                variant="plain"
                density="compact"
                auto-grow
                hide-details="auto"
                autocomplete="off"
                class="meta-inline-edit"
                :class="`meta-inline-${f.key}`"
                @update:model-value="(v) => f.write(photo, v)"
                @keydown.escape.prevent="cancelEditing"
                @blur="onInlineFieldBlur"
              ></v-textarea>
              <!-- eslint-disable-next-line vue/no-v-html -- f.htmlValue references a sanitized computed (e.g. notesHtml) — encode-then-sanitize via $util.sanitizeHtml($util.encodeHTML(raw)). -->
              <div v-else-if="f.display === 'html' && fieldHtml(f)" class="text-body-2 meta-scrollable" :class="`meta-${f.key}`" v-html="fieldHtml(f)"></div>
              <div v-else-if="f.display !== 'html' && f.read(photo)" class="text-body-2 meta-scrollable" :class="`meta-${f.key}`">
                {{ f.read(photo) }}
              </div>
              <div v-else class="meta-add-prompt" @click.stop="startEditing(f.key)">{{ f.label }}</div>
            </v-list-item>
          </template>
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
    <p-confirm-dialog
      :visible="discardDialog.visible"
      icon=""
      :text="$gettext('Discard unsaved changes?')"
      :action="$gettext('Discard')"
      confirm-color="info"
      @close="onDiscardCancel"
      @confirm="onDiscardConfirm"
    ></p-confirm-dialog>
    <p-confirm-dialog
      :visible="addNameDialog.visible"
      icon="mdi-account-plus"
      :icon-size="42"
      :text="addNameDialog.name ? $gettext('Add %{s}?', { s: addNameDialog.name }) : $gettext('Add person?')"
      confirm-color="info"
      @close="onAddNameCancel"
      @confirm="onAddNameConfirm"
    ></p-confirm-dialog>
  </div>
</template>

<script>
import { DateTime } from "luxon";
import * as formats from "options/formats";

import * as media from "common/media";
import { Photo } from "model/photo";
import { Label } from "model/label";
import { Album } from "model/album";
import PMap from "component/map.vue";
import PDateTimeDialog from "component/sidebar/datetime-dialog.vue";
import PCameraDialog from "component/sidebar/camera-dialog.vue";
import PLocationDialog from "component/location/dialog.vue";
import PConfirmDialog from "component/confirm/dialog.vue";
import PSidebarInlineToolbar from "component/sidebar/inline-toolbar.vue";

export default {
  name: "PSidebarInfo",
  components: {
    PMap,
    PDateTimeDialog,
    PCameraDialog,
    PLocationDialog,
    PConfirmDialog,
    PSidebarInlineToolbar,
  },
  props: {
    // UID of the photo currently shown in the parent lightbox. Drives the
    // sidebar lifecycle (re-fetching markers, resetting inline edits) when
    // the user navigates between slides. All other parent state is read
    // through `view` (see data() below) — this matches the pattern used
    // by component/photo/edit/labels.vue.
    uid: {
      type: String,
      default: "",
    },
  },
  emits: ["close", "toggle-markers-visible", "toggle-adding-marker", "remove-marker", "eject-marker", "reload-markers", "naming-started"],
  data() {
    return {
      // Live reactive handle to the parent lightbox's $data, captured once at
      // mount via `$view.getData()`. The lightbox calls `$view.enter(this)`
      // before the sidebar mounts (see lightbox.vue:showDialog), so this is
      // populated by the time data() runs. Mutations through this.view.X
      // write through to the parent and don't trigger vue/no-mutating-props.
      view: this.$view.getData(),
      actions: [],
      featPeople: this.$config.feature("people"),
      featPlaces: this.$config.feature("places"),
      textRule: (v) => !v || v.length <= this.$config.get("clip") || this.$gettext("Text too long"),
      dateTimeDialog: false,
      cameraDialog: false,
      locationDialog: false,
      editingField: null,
      editOriginal: null,
      chipInput: null,
      chipSearch: "",
      chipKey: 0,
      labelOptions: [],
      albumOptions: [],
      // Pending chip mutations staged during edit mode. Labels are keyed by
      // Label.ID for removals and by typed name for additions; albums are
      // keyed by Album.UID for removals and stored as full album objects
      // for additions (the title is read off the object at confirm time).
      chipState: {
        labels: { additions: [], removals: [] },
        albums: { additions: [], removals: [] },
      },
      markerDrafts: {},
      markerNameRule: (v) => !v || v.length <= this.$config.get("clip") || this.$gettext("Text too long"),
      markerMenuProps: {
        openOnFocus: true,
        closeOnContentClick: true,
        maxHeight: 260,
        class: "meta-inline-menu",
      },
      chipMenuProps: {
        class: "meta-inline-menu",
      },
      discardDialog: {
        visible: false,
        resolver: null,
      },
      addNameDialog: {
        visible: false,
        marker: null,
        name: "",
      },
    };
  },
  computed: {
    // Aliases for parent-owned reactive state. These read through `view` so
    // every existing template/script reference (this.model.X, this.photo.X, etc.)
    // keeps working without churn. Mutations are explicit: write to
    // `this.view.photo.X` / `this.view.model.X`, never to `this.photo` / `this.model`.
    model() {
      return this.view?.model;
    },
    photo() {
      return this.view?.photo;
    },
    canEdit() {
      return Boolean(this.view?.canEdit && this.view?.contextAllowsEdit);
    },
    collection() {
      return this.view?.collection;
    },
    context() {
      return this.view?.context;
    },
    markersVisible() {
      return Boolean(this.view?.markersVisible);
    },
    addingMarker() {
      return Boolean(this.view?.addingMarker);
    },
    markersBusy() {
      return Boolean(this.view?.markersBusy);
    },
    newMarkerUid() {
      return this.view?.pendingNameMarkerUid;
    },
    isEditable() {
      return this.canEdit && this.photo && this.photo.Details && !this.restrictedRole;
    },
    restrictedRole() {
      return this.$session.isSidebarRestricted();
    },
    captionHtml() {
      const raw = this.photo?.Caption ?? this.model?.Caption;
      if (!raw) return "";
      return this.$util.sanitizeHtml(this.$util.encodeHTML(raw));
    },
    notesHtml() {
      if (!this.photo?.Details?.Notes) return "";
      return this.$util.sanitizeHtml(this.$util.encodeHTML(this.photo.Details.Notes));
    },
    cameraInfo() {
      if (!this.photo) return "";
      // Backend returns the "Unknown" placeholder camera (CameraID=1,
      // Camera={Make:"", Model:"Unknown"}) when no EXIF camera is set, and
      // formatCamera() happily renders that as " Unknown". Suppress it so
      // the read-only sidebar doesn't surface an empty camera row.
      const hasRealCamera =
        (this.photo.CameraID && this.photo.CameraID > 1) ||
        (this.photo.CameraMake && this.photo.CameraMake.trim()) ||
        (this.photo.CameraModel && this.photo.CameraModel.trim() && this.photo.CameraModel !== "Unknown");
      if (!hasRealCamera) return "";
      // Suppress "Unknown, ISO 100"-style rows when only ISO/exposure are set.
      if (!this.$util.formatCamera(this.photo.Camera, this.photo.CameraID, this.photo.CameraMake, this.photo.CameraModel, false)) return "";
      const info = this.photo.getCameraInfo();
      return info !== this.$gettext("Unknown") ? info : "";
    },
    lensInfo() {
      if (!this.photo) return "";
      const hasLens =
        (this.photo.LensID && this.photo.LensID > 1) || this.photo.LensMake || this.photo.LensModel || this.photo.Lens?.Model || this.photo.Lens?.Make;
      if (!hasLens) return "";
      const info = this.photo.getLensInfo();
      return info !== this.$gettext("Unknown") ? info : "";
    },
    exifInfo() {
      if (!this.photo) return "";
      return this.photo.getExifInfo();
    },
    people() {
      if (!this.photo) return [];
      return this.photo.getMarkers(true);
    },
    knownPeople() {
      const values = this.$config && this.$config.values;
      if (!values || !Array.isArray(values.people)) return [];
      return values.people;
    },
    labels() {
      if (!this.photo?.Labels) return [];
      return this.photo.Labels.filter((l) => l.Label && l.Label.Name && l.Uncertainty < 100);
    },
    albums() {
      if (!this.photo?.Albums) return [];
      return this.photo.Albums.filter((a) => a.Title && !a.Private);
    },
    subject() {
      return this.photo?.Details?.Subject || "";
    },
    artist() {
      return this.photo?.Details?.Artist || "";
    },
    copyright() {
      return this.photo?.Details?.Copyright || "";
    },
    license() {
      return this.photo?.Details?.License || "";
    },
    keywords() {
      return this.photo?.Details?.Keywords || "";
    },
    // Single source of truth for inline-text fields. Each entry knows how to
    // read/write its raw value, what label to render (tooltip, placeholder,
    // add-prompt), and whether the display branch should treat the value as
    // sanitized HTML (Caption, Notes) or plain text (everything else).
    // detailsFields/textFields below select subsets for the two visual layouts.
    fieldRegistry() {
      return {
        title: {
          key: "title",
          label: this.$pgettext("Photo", "Title"),
          read: (p) => p?.Title,
          write: (p, v) => {
            if (p) p.Title = v;
          },
          display: "text",
        },
        caption: {
          key: "caption",
          label: this.$gettext("Caption"),
          read: (p) => p?.Caption,
          write: (p, v) => {
            if (p) p.Caption = v;
          },
          display: "html",
          htmlValue: "captionHtml",
        },
        subject: {
          key: "subject",
          label: this.$gettext("Subject"),
          icon: "mdi-text-box-outline",
          read: (p) => p?.Details?.Subject,
          write: (p, v) => {
            if (p?.Details) p.Details.Subject = v;
          },
          display: "text",
        },
        artist: {
          key: "artist",
          label: this.$gettext("Artist"),
          icon: "mdi-palette",
          read: (p) => p?.Details?.Artist,
          write: (p, v) => {
            if (p?.Details) p.Details.Artist = v;
          },
          display: "text",
        },
        copyright: {
          key: "copyright",
          label: this.$gettext("Copyright"),
          icon: "mdi-copyright",
          read: (p) => p?.Details?.Copyright,
          write: (p, v) => {
            if (p?.Details) p.Details.Copyright = v;
          },
          display: "text",
        },
        license: {
          key: "license",
          label: this.$gettext("License"),
          icon: "mdi-license",
          read: (p) => p?.Details?.License,
          write: (p, v) => {
            if (p?.Details) p.Details.License = v;
          },
          display: "text",
        },
        keywords: {
          key: "keywords",
          label: this.$gettext("Keywords"),
          read: (p) => p?.Details?.Keywords,
          write: (p, v) => {
            if (p?.Details) p.Details.Keywords = v;
          },
          display: "text",
        },
        notes: {
          key: "notes",
          label: this.$gettext("Notes"),
          read: (p) => p?.Details?.Notes,
          write: (p, v) => {
            if (p?.Details) p.Details.Notes = v;
          },
          display: "html",
          htmlValue: "notesHtml",
        },
      };
    },
    detailsFields() {
      return ["subject", "artist", "copyright", "license"].map((k) => this.fieldRegistry[k]);
    },
    textFields() {
      return ["keywords", "notes"].map((k) => this.fieldRegistry[k]);
    },
    showDetailsSection() {
      if (this.restrictedRole) return false;
      if (this.isEditable) return true;
      return this.detailsFields.some((f) => Boolean(f.read(this.photo)));
    },
    placeName() {
      if (!this.photo) return "";
      return this.photo.locationInfo() || "";
    },
    altitude() {
      if (!this.photo || !this.photo.Altitude) return "";
      return this.photo.Altitude + " m";
    },
    fileName() {
      if (!this.photo) return "";
      if (this.photo.FileName) return this.photo.FileName;
      const primary = typeof this.photo.primaryFile === "function" ? this.photo.primaryFile() : null;
      return primary?.Name || "";
    },
    fileInfo() {
      if (this.photo) {
        switch (this.photo.Type) {
          case media.Video:
          case media.Live:
          case media.Animated:
            return this.photo.getVideoInfo();
          case media.Vector:
          case media.Document:
            return this.photo.getVectorInfo();
          default:
            return this.photo.getImageInfo();
        }
      }
      // Fallback for restricted roles: Thumb.getTypeInfo() produces
      // format, megapixels, and dimensions from the viewer endpoint data.
      if (this.model && typeof this.model.getTypeInfo === "function") {
        return this.model.getTypeInfo();
      }
      return "";
    },
    fileIcon() {
      switch (this.photo?.Type || this.model?.Type) {
        case media.Raw:
          return "mdi-raw";
        case media.Video:
          return "mdi-video";
        case media.Live:
          return "mdi-play-circle-outline";
        case media.Animated:
          return "mdi-file-gif-box";
        case media.Vector:
          return "mdi-vector-polyline";
        case media.Document:
          return "mdi-file-pdf-box";
        default:
          return "mdi-image-outline";
      }
    },
  },
  watch: {
    people: {
      immediate: true,
      handler(markers) {
        this.syncMarkerDrafts(Array.isArray(markers) ? markers : []);
      },
    },
    newMarkerUid(uid) {
      if (!uid) return;
      this.$nextTick(() => this.focusMarkerInput(uid));
    },
  },
  methods: {
    close() {
      this.$emit("close");
    },
    getFieldValue(field) {
      const f = this.fieldRegistry[field];
      if (!f) return "";
      const v = f.read(this.photo);
      return v == null ? "" : v;
    },
    setFieldValue(field, value) {
      const f = this.fieldRegistry[field];
      if (!f || !this.view?.photo) return;
      f.write(this.view.photo, value);
    },
    // Function ref shared by every inline editor. Vue invokes it with the
    // mounted component on mount and null on unmount; since each editor is
    // gated by a unique `editingField === '<key>'`, only one is mounted at
    // a time, so the latest non-null call always identifies the active one.
    setInlineEditorRef(el) {
      if (el) this._inlineEditorEl = el;
      else if (!this.editingField) this._inlineEditorEl = null;
    },
    fieldHtml(f) {
      if (!f || f.display !== "html" || !f.htmlValue) return "";
      return this[f.htmlValue] || "";
    },
    shouldShowFieldRow(f) {
      if (!f) return false;
      if (this.editingField === f.key) return true;
      if (this.isEditable) return true;
      if (f.display === "html") return Boolean(this.fieldHtml(f));
      return Boolean(f.read(this.photo));
    },
    startEditing(field) {
      if (this.editingField) {
        this.cancelEditing();
      }

      this.editingField = field;
      this.editOriginal = this.getFieldValue(field);
      this._editStartedAt = Date.now();

      this.$nextTick(() => {
        const editor = this._inlineEditorEl;
        if (editor && typeof editor.focus === "function") editor.focus();
      });
    },
    onToggleMarkersVisible() {
      if (!this.isEditable || this.markersBusy) return;
      this.$emit("toggle-markers-visible");
    },
    onToggleAddingMarker() {
      if (!this.isEditable || this.markersBusy) return;
      this.$emit("toggle-adding-marker");
    },
    onRemoveMarker(marker) {
      if (!this.isEditable || this.markersBusy || !marker || marker.SubjUID) return;
      this.$emit("remove-marker", marker);
    },
    onEjectMarker(marker) {
      if (!this.isEditable || this.markersBusy || !marker || !marker.SubjUID) return;
      this.$emit("eject-marker", marker);
    },
    // Combobox can bind either the typed string or the selected subject object.
    unwrapMarkerName(value) {
      return typeof value === "object" && value !== null ? value.Name || "" : value || "";
    },
    syncMarkerDrafts(markers) {
      const seen = new Set();
      markers.forEach((m) => {
        if (!m || !m.UID) return;
        seen.add(m.UID);
        const original = m.Name || "";
        const existing = this.markerDrafts[m.UID];
        if (!existing) {
          this.markerDrafts[m.UID] = { original, current: original };
        } else if (existing.original !== original) {
          existing.original = original;
          existing.current = original;
        }
      });
      Object.keys(this.markerDrafts).forEach((uid) => {
        if (!seen.has(uid)) delete this.markerDrafts[uid];
      });
    },
    markerInputValue(uid) {
      const d = this.markerDrafts[uid];
      return d ? d.current : "";
    },
    markerInputSearch(uid) {
      return this.unwrapMarkerName(this.markerInputValue(uid));
    },
    setMarkerInputValue(uid, value) {
      if (!uid) return;
      if (!this.markerDrafts[uid]) {
        this.markerDrafts[uid] = { original: "", current: value };
      } else {
        this.markerDrafts[uid].current = value;
      }
    },
    focusMarkerInput(uid) {
      if (!uid) return;
      this.$emit("naming-started");
      this.$nextTick(() => {
        const input = this.$el && this.$el.querySelector(`[data-marker-uid="${uid}"] input`);
        if (input) input.focus();
      });
    },
    // Match a typed name against knownPeople case-insensitively so the backend
    // doesn't create a duplicate subject when the input only differs in case.
    findKnownPerson(name) {
      if (!name) return null;
      return this.knownPeople.find((p) => p && p.Name && p.Name.localeCompare(name, "en", { sensitivity: "base" }) === 0) || null;
    },
    confirmMarkerName(marker, source = "enter") {
      if (!marker || !marker.UID) return;
      const draft = this.markerDrafts[marker.UID];
      if (!draft) return;
      const name = this.unwrapMarkerName(draft.current).trim();
      const original = (draft.original || "").trim();

      if (!name || name === original) return;
      if (typeof marker.setName !== "function") return;

      const match = this.findKnownPerson(name);

      // Blur without Enter on an unnamed marker → ask before committing a new
      // name. Skip the dialog if the person already exists (match) or if the
      // marker is already named (eject/rename path) — both are unambiguous.
      if (source === "blur" && !marker.SubjUID && !match) {
        this.addNameDialog = { visible: true, marker, name };
        return;
      }

      this.commitMarkerName(marker, match, name);
    },
    commitMarkerName(marker, match, name) {
      const draft = this.markerDrafts[marker.UID];
      if (!draft) return;

      if (match) {
        marker.Name = match.Name;
        marker.SubjUID = match.UID;
      } else {
        marker.Name = name;
      }

      // Lock the draft to the saved name so a parallel people-reload watcher
      // tick doesn't snap the input back to the old value mid-request.
      draft.original = marker.Name || "";
      draft.current = marker.Name || "";

      marker
        .setName()
        .then(() => {
          this.$emit("reload-markers", marker);
        })
        .catch(() => {
          this.$notify.error(this.$gettext("Failed to save name"));
        });
    },
    onPickPerson(marker, value) {
      if (!marker || !value || typeof value !== "object" || !value.Name) return;
      this.setMarkerInputValue(marker.UID, value.Name);
      this.confirmMarkerName(marker, "enter");
    },
    onAddNameConfirm() {
      const { marker, name } = this.addNameDialog;
      this.addNameDialog = { visible: false, marker: null, name: "" };
      if (!marker || !name) return;
      this.commitMarkerName(marker, this.findKnownPerson(name), name);
    },
    onAddNameCancel() {
      const { marker } = this.addNameDialog;
      this.addNameDialog = { visible: false, marker: null, name: "" };
      const draft = marker && marker.UID ? this.markerDrafts[marker.UID] : null;
      if (draft) draft.current = draft.original || "";
    },
    cancelMarkerName(marker) {
      if (!marker || !marker.UID) return;
      const draft = this.markerDrafts[marker.UID];
      if (!draft) return;
      draft.current = draft.original;
      // Blur the active input so the user gets a visual cue the edit was
      // dropped; @blur re-fires confirmMarkerName but it's a no-op now that
      // current === original.
      if (typeof document !== "undefined" && document.activeElement && typeof document.activeElement.blur === "function") {
        document.activeElement.blur();
      }
    },
    resetInlineEdits() {
      if (this.editingField) this.cancelEditing();
      Object.keys(this.markerDrafts).forEach((uid) => {
        const d = this.markerDrafts[uid];
        if (d) d.current = d.original;
      });
    },
    // Inline text fields (title/caption/subject/...) are excluded on purpose:
    // onInlineFieldBlur() auto-commits them before any navigation source can
    // fire, so they can never have pending state at nav time. Only the
    // staged editors that do NOT auto-commit on blur belong here.
    hasPendingEdit() {
      for (const uid of Object.keys(this.markerDrafts)) {
        const d = this.markerDrafts[uid];
        if (!d) continue;
        if (this.unwrapMarkerName(d.current).trim() !== (d.original || "").trim()) return true;
      }
      if (Object.values(this.chipState).some((s) => s.additions.length || s.removals.length)) return true;
      return false;
    },
    // Async guard used by the lightbox before closing / hiding / navigating.
    // Returns a Promise<boolean>: true = safe to proceed, false = user canceled.
    confirmDiscardPending() {
      if (!this.hasPendingEdit()) return Promise.resolve(true);
      if (this.discardDialog.visible && this.discardDialog.resolver) {
        // Another request is already waiting on the dialog; reuse it.
        return new Promise((resolve) => {
          const prev = this.discardDialog.resolver;
          this.discardDialog.resolver = (ok) => {
            prev(ok);
            resolve(ok);
          };
        });
      }
      return new Promise((resolve) => {
        this.discardDialog.resolver = resolve;
        this.discardDialog.visible = true;
      });
    },
    onDiscardConfirm() {
      this.discardDialog.visible = false;
      const r = this.discardDialog.resolver;
      this.discardDialog.resolver = null;
      this.resetInlineEdits();
      if (r) r(true);
    },
    onDiscardCancel() {
      this.discardDialog.visible = false;
      const r = this.discardDialog.resolver;
      this.discardDialog.resolver = null;
      if (r) r(false);
    },
    confirmField() {
      if (!this.photo || !this.canEdit) {
        this.editingField = null;
        return;
      }

      const field = this.editingField;
      this.editingField = null;
      this.editOriginal = null;

      if (!this.photo.wasChanged()) {
        return;
      }

      // The inline-edit binding (v-model="photo.X") already mutated the photo
      // optimistically; on success sync the matching Thumb fields, on
      // failure roll back so the user sees the title/caption revert and
      // gets a notification instead of a silent ghost edit.
      this.photo
        .update()
        .then(() => {
          if ((field === "title" || field === "caption") && this.view?.model) {
            this.view.model.Title = this.view.photo.Title;
            this.view.model.Caption = this.view.photo.Caption;
          }
        })
        .catch(() => {
          this.photo.rollback();
          this.$notify.error(this.$gettext("Failed to save changes"));
        });
    },
    cancelEditing() {
      // Guard: clicking a pencil icon triggers blur on the previous field before focus lands
      // on the new input, which would immediately cancel the edit we just started.
      if (this._editStartedAt && Date.now() - this._editStartedAt < 200) {
        return;
      }

      if (this.editingField && this.editOriginal !== null) {
        this.setFieldValue(this.editingField, this.editOriginal);
      }

      this.editingField = null;
      this.editOriginal = null;
      this._editStartedAt = null;
      this.resetChipState();
    },
    // Blur handler for inline text fields (title/caption/subject/artist/
    // copyright/license/keywords/notes). Commits the edit instead of
    // silently reverting so the user does not lose their typing when
    // they click away, swipe to the next photo, or press the nav arrow.
    // Escape still cancels explicitly via cancelEditing().
    onInlineFieldBlur() {
      if (this._editStartedAt && Date.now() - this._editStartedAt < 200) {
        return;
      }
      if (!this.editingField) return;
      this.confirmField();
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
    openInNewTab(route) {
      if (!route) return;
      const resolved = this.$router ? this.$router.resolve(route) : null;
      const href = resolved?.href || "";
      if (!href || typeof window === "undefined" || typeof window.open !== "function") return;
      window.open(href, "_blank", "noopener,noreferrer");
    },
    navigateToLabel(label) {
      if (!label) return;
      const slug = label.CustomSlug || label.Slug;
      if (!slug) return;
      this.openInNewTab({ name: "browse", query: { q: "label:" + slug } });
    },
    navigateToAlbum(album) {
      if (!album || !album.UID) return;
      this.openInNewTab({ name: "album", params: { album: album.UID, slug: "view" } });
    },
    navigateToPerson(marker) {
      if (!marker) return;
      if (marker.SubjUID) {
        this.openInNewTab({ name: "browse", query: { q: "subject:" + marker.SubjUID } });
      } else if (marker.Name) {
        this.openInNewTab({ name: "browse", query: { q: "person:" + marker.Name } });
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
    // Generic chip-state helpers. Field is "labels" or "albums"; the key is
    // whatever uniquely identifies a chip in that field's domain (Label.ID
    // for labels, Album.UID for albums on the removals side; the typed name
    // for label additions, Album.UID for album additions).
    isChipPendingRemoval(field, key) {
      const state = this.chipState[field];
      return Boolean(state && key != null && state.removals.includes(key));
    },
    togglePendingChipRemoval(field, key) {
      const state = this.chipState[field];
      if (!state || key == null) return;
      const idx = state.removals.indexOf(key);
      if (idx >= 0) {
        state.removals.splice(idx, 1);
      } else {
        state.removals.push(key);
      }
    },
    removePendingChipAdd(field, key) {
      const state = this.chipState[field];
      if (!state || key == null) return;
      const idx = field === "labels" ? state.additions.indexOf(key) : state.additions.findIndex((a) => a.UID === key);
      if (idx >= 0) state.additions.splice(idx, 1);
    },
    resetChipState() {
      Object.values(this.chipState).forEach((s) => {
        s.additions = [];
        s.removals = [];
      });
    },
    addPendingLabel(rawName) {
      const name = (rawName || "").trim();
      if (!name) return false;
      if (name.length > this.$config.get("clip")) {
        this.$notify.error(this.$gettext("Name too long"));
        return false;
      }
      const norm = this.$util.normalizeLabelTitle(name);
      if (!norm) return false;
      const additions = this.chipState.labels.additions;
      if (additions.some((n) => this.$util.normalizeLabelTitle(n) === norm)) return false;
      if (this.labels.some((l) => this.$util.normalizeLabelTitle(l?.Label?.Name) === norm)) return false;
      additions.push(name);
      return true;
    },
    albumTitleConflicts(norm) {
      if (!norm) return true;
      if (this.chipState.albums.additions.some((a) => this.$util.normalizeLabelTitle(a?.Title) === norm)) return true;
      if (this.albums.some((a) => this.$util.normalizeLabelTitle(a?.Title) === norm)) return true;
      return false;
    },
    addPendingAlbum(album) {
      if (!album || typeof album !== "object") return false;
      const title = (album.Title || "").trim();
      if (!title) return false;
      if (title.length > this.$config.get("clip")) {
        this.$notify.error(this.$gettext("Name too long"));
        return false;
      }
      const additions = this.chipState.albums.additions;
      if (album.UID) {
        if (additions.some((a) => a.UID === album.UID)) return false;
        if (this.albums.some((a) => a.UID === album.UID)) return false;
      }
      if (this.albumTitleConflicts(this.$util.normalizeLabelTitle(title))) return false;
      additions.push(album);
      return true;
    },
    onLabelSelected(value) {
      if (!value || typeof value !== "object" || !value.Name) return;
      this.addPendingLabel(value.Name);
      this.clearChipInput();
    },
    // Read the typed name from the input DOM as a fallback so that
    // Vuetify clearing `chipSearch` on the same Enter keystroke we
    // handle does not silently swallow the pending addition.
    pendingChipName(ev) {
      if (this.chipSearch) return this.chipSearch;
      const target = ev && ev.target ? ev.target : null;
      return target && typeof target.value === "string" ? target.value : "";
    },
    onLabelEnter(ev) {
      if (this.addPendingLabel(this.pendingChipName(ev))) {
        this.clearChipInput();
      }
    },
    confirmLabels() {
      if (!this.photo) {
        this.editingField = null;
        return;
      }

      const state = this.chipState.labels;
      const removals = state.removals.slice();
      const additions = state.additions.slice();
      this.editingField = null;
      state.removals = [];
      state.additions = [];

      const promises = [];
      removals.forEach((id) => promises.push(this.photo.removeLabel(id)));
      additions.forEach((name) => promises.push(this.photo.addLabel(name)));

      // Cache freshness: photo.addLabel / removeLabel patch this.photo.Labels
      // locally on success, and the backend publishes photos.updated which
      // evicts the cached entry via evictCachedFromEntities — see
      // model/photo.js. confirmAlbums needs an explicit evict + re-find
      // because Album mutations go through raw $api.delete/post and don't
      // patch this.photo.Albums; that asymmetry is intentional.
      if (promises.length) {
        Promise.all(promises).catch(() => {
          this.$notify.error(this.$gettext("Failed to save changes"));
        });
      }
    },
    confirmAlbums() {
      if (!this.photo) {
        this.editingField = null;
        return;
      }

      const state = this.chipState.albums;
      const removals = state.removals.slice();
      const additions = state.additions.slice();
      this.editingField = null;
      state.removals = [];
      state.additions = [];

      const promises = [];
      removals.forEach((uid) => promises.push(this.$api.delete(`albums/${uid}/photos`, { data: { photos: [this.photo.UID] } })));
      additions.forEach((a) => promises.push(this.$api.post(`albums/${a.UID}/photos`, { photos: [this.photo.UID] })));

      if (promises.length) {
        Promise.all(promises)
          .then(() => {
            // Album mutations don't patch this.photo.Albums locally and the
            // backend publishes only albums.updated (not photos.updated) for
            // membership changes, so we evict + re-find here so the sidebar
            // reflects the saved state without waiting for navigation.
            Photo.evictCache(this.photo.UID);
            return this.photo.find(this.photo.UID);
          })
          .then((photo) => {
            this.photo.setValues(photo.getValues());
          })
          .catch(() => {
            this.$notify.error(this.$gettext("Failed to save changes"));
          });
      }
    },
    onAlbumSelected(value) {
      if (!value || typeof value !== "object" || !value.UID) {
        this.clearChipInput();
        return;
      }
      this.addPendingAlbum(value);
      this.clearChipInput();
    },
    onAlbumEnter(ev) {
      const search = this.pendingChipName(ev).trim();
      if (!search) return;

      if (search.length > this.$config.get("clip")) {
        this.$notify.error(this.$gettext("Name too long"));
        return;
      }

      const norm = this.$util.normalizeLabelTitle(search);
      if (!norm) {
        this.clearChipInput();
        return;
      }

      // TODO: this partial-match lookup (startsWith/includes against the
      // typed string) can produce spurious matches for short inputs and
      // should be revisited separately.
      const lower = search.toLowerCase();
      const match =
        this.albumOptions.find((a) => a.Title.toLowerCase().startsWith(lower)) || this.albumOptions.find((a) => a.Title.toLowerCase().includes(lower));

      if (match) {
        this.onAlbumSelected(match);
        return;
      }

      // Skip the API round-trip if a normalized title clash already exists.
      if (this.albumTitleConflicts(norm)) {
        this.clearChipInput();
        return;
      }

      const album = new Album({ Title: search });

      album
        .save()
        .then(() => {
          if (album.UID && this.addPendingAlbum(album)) {
            this.albumOptions.push(album);
          }
        })
        .catch(() => {})
        .finally(() => {
          this.clearChipInput();
        });
    },
    confirmDateTime(data) {
      this.dateTimeDialog = false;

      const photo = this.view?.photo;
      if (!photo || !photo.UID || !this.canEdit) return;

      photo.Day = data.Day;
      photo.Month = data.Month;
      photo.Year = data.Year;
      photo.TimeZone = data.TimeZone;

      const localDate = photo.localDate(data.time);
      if (!localDate.isValid) return;

      const isoTime =
        localDate.toISO({
          suppressMilliseconds: true,
          includeOffset: false,
        }) + "Z";

      photo.TakenAtLocal = isoTime;

      if (photo.currentTimeZoneUTC()) {
        photo.TakenAt = isoTime;
      }

      photo
        .update()
        .then(() => {
          if (!this.view?.model) return;
          this.view.model.TakenAtLocal = photo.TakenAtLocal;
          this.view.model.TimeZone = photo.TimeZone;
        })
        .catch(() => {
          photo.rollback();
          this.$notify.error(this.$gettext("Failed to save changes"));
        });
    },
    confirmCamera(data) {
      this.cameraDialog = false;

      const photo = this.view?.photo;
      if (!photo || !photo.UID || !this.canEdit) return;

      photo.CameraID = data.CameraID;
      photo.LensID = data.LensID;
      photo.Iso = data.Iso;
      photo.Exposure = data.Exposure;
      photo.FNumber = data.FNumber;
      photo.FocalLength = data.FocalLength;

      // photo.update() resets __originalValues only on success; on rejection
      // the snapshot still holds the pre-mutation state, so rollback() puts
      // every field back without per-field bookkeeping here.
      photo.update().catch(() => {
        photo.rollback();
        this.$notify.error(this.$gettext("Failed to save changes"));
      });
    },
    confirmLocation(data) {
      this.locationDialog = false;

      const photo = this.view?.photo;
      if (!photo || !photo.UID || !this.canEdit) return;

      photo.Lat = data.lat;
      photo.Lng = data.lng;
      photo.PlaceSrc = "manual";

      if (data.location?.country) {
        photo.Country = data.location.country;
      }

      photo
        .update()
        .then(() => {
          if (!this.view?.model) return;
          this.view.model.Lat = photo.Lat;
          this.view.model.Lng = photo.Lng;
        })
        .catch(() => {
          photo.rollback();
          this.$notify.error(this.$gettext("Failed to save changes"));
        });
    },
  },
};
</script>
