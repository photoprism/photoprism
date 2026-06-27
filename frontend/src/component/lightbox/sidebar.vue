<template>
  <div
    class="p-lightbox-sidebar bg-background metadata"
    :class="{ 'hide-edit-pencils': hideEditPencils, 'hide-edit-undo': hideEditUndo, 'hide-edit-save': hideEditSave }"
  >
    <v-toolbar density="comfortable" color="background">
      <v-btn :icon="$isRtl ? 'mdi-chevron-left' : 'mdi-chevron-right'" :title="$gettext('Close')" @click.stop="close()"></v-btn>
      <v-toolbar-title>{{ $gettext(`Information`) }}</v-toolbar-title>
    </v-toolbar>
    <div v-if="model.UID">
      <v-list nav slim tile density="compact" bg-color="background" class="metadata__list mt-1">
        <!-- Title -->
        <v-list-item
          v-if="editingField === 'title' || model.Title || isEditable"
          :class="['metadata__item', { clickable: editingField !== 'title' && (isEditable || model.Title) }]"
          @click.stop="onTextRowClick($event, 'title', model.Title)"
        >
          <v-text-field
            v-if="editingField === 'title'"
            :ref="setInlineEditorRef"
            v-model="photo.Title"
            :rules="rules.text(false, 0, fieldRegistry.title.maxLength, fieldRegistry.title.label)"
            density="compact"
            hide-details="auto"
            autocomplete="off"
            class="meta-inline-edit meta-inline-title"
            @keydown.enter.stop.prevent="confirmField"
            @keydown.escape.stop.prevent="cancelEditing"
            @blur="onInlineFieldBlur"
          ></v-text-field>
          <div v-else-if="model.Title" class="text-subtitle-2 meta-title">{{ model.Title }}</div>
          <div v-else class="meta-add-prompt" @click.stop="startEditing('title')">{{ $pgettext("Photo", "Add a Title") }}</div>
          <template v-if="isEditable" #append>
            <p-lightbox-sidebar-toolbar
              :editing="editingField === 'title'"
              :can-undo="editingField === 'title'"
              :undo-disabled="!inlineEditDirty"
              @confirm="confirmField"
              @undo="undoInlineEdit"
              @start="startEditing('title')"
            />
          </template>
        </v-list-item>

        <!-- Caption -->
        <v-list-item
          v-if="editingField === 'caption' || model.Caption || isEditable"
          :class="['metadata__item', { clickable: editingField !== 'caption' && (isEditable || model.Caption) }]"
          @click.stop="onTextRowClick($event, 'caption', model.Caption)"
        >
          <v-textarea
            v-if="editingField === 'caption'"
            :ref="setInlineEditorRef"
            v-model="photo.Caption"
            :rows="1"
            :max-rows="14"
            density="compact"
            auto-grow
            hide-details="auto"
            autocomplete="off"
            class="meta-inline-edit meta-inline-caption"
            @keydown.enter.stop
            @keydown.escape.stop.prevent="cancelEditing"
            @blur="onInlineFieldBlur"
          ></v-textarea>
          <!-- eslint-disable-next-line vue/no-v-html -- captionHtml is encode-then-sanitized via $util.sanitizeHtml($util.encodeHTML(raw)); see captionHtml() computed -->
          <div v-else-if="model.Caption" class="text-body-2 meta-caption meta-scrollable text-html" v-html="captionHtml"></div>
          <div v-else class="meta-add-prompt" @click.stop="startEditing('caption')">{{ $gettext("Add a Caption") }}</div>
          <template v-if="isEditable" #append>
            <p-lightbox-sidebar-toolbar
              :editing="editingField === 'caption'"
              :can-undo="editingField === 'caption'"
              :undo-disabled="!inlineEditDirty"
              @confirm="confirmField"
              @undo="undoInlineEdit"
              @start="startEditing('caption')"
            />
          </template>
        </v-list-item>

        <v-divider v-if="editingField === 'title' || editingField === 'caption' || model.Title || model.Caption || isEditable" class="my-3"></v-divider>

        <v-list-item
          v-if="fileInfo"
          v-tooltip="fileTypeName"
          :prepend-icon="fileIcon"
          :title="fileInfo"
          :subtitle="fileName"
          :lines="fileName ? 'two' : 'one'"
          :class="['metadata__item', 'meta-file', { clickable: !!fileName }]"
          @click.stop.prevent="fileName && $util.copyText(fileName)"
        >
        </v-list-item>

        <v-divider v-if="fileName || fileInfo" class="my-3"></v-divider>

        <v-list-item
          v-tooltip="$gettext(`Taken`)"
          :title="formatTime(model)"
          prepend-icon="mdi-calendar"
          :class="['metadata__item', { clickable: isEditable }]"
          @click.stop="openDateTimeDialog"
        >
          <template v-if="isEditable" #append>
            <v-btn
              icon="mdi-pencil-outline"
              density="compact"
              variant="plain"
              size="x-small"
              class="meta-inline-pencil"
              :title="$gettext('Edit')"
              @click.stop="openDateTimeDialog"
            ></v-btn>
          </template>
        </v-list-item>

        <v-list-item
          v-if="canViewLibrary && (cameraInfo || isEditable)"
          v-tooltip="$gettext('Camera')"
          :title="cameraInfo || $gettext('Unknown')"
          prepend-icon="mdi-camera"
          :class="['metadata__item', { clickable: isEditable }]"
          @click.stop="openCameraDialog"
        >
          <template v-if="isEditable" #append>
            <v-btn
              icon="mdi-pencil-outline"
              density="compact"
              variant="plain"
              size="x-small"
              class="meta-inline-pencil"
              :title="$gettext('Edit')"
              @click.stop="openCameraDialog"
            ></v-btn>
          </template>
        </v-list-item>

        <v-list-item
          v-if="canViewLibrary && lensInfo"
          v-tooltip="$gettext('Lens')"
          :title="lensInfo"
          prepend-icon="mdi-camera-iris"
          :class="['metadata__item', { clickable: isEditable }]"
          @click.stop="openCameraDialog"
        ></v-list-item>

        <template v-if="locationRowVisible">
          <!-- v-divider class="my-3"></v-divider -->
          <v-list-item
            v-tooltip="$gettext('Location')"
            :title="locationTitle"
            :subtitle="locationSubtitle"
            :lines="locationSubtitle ? 'two' : 'one'"
            prepend-icon="mdi-map-marker"
            :class="['metadata__item', 'meta-location', { 'meta-coordinates': model.Lat && model.Lng, 'clickable': locationRowClickable }]"
            @click.stop.prevent="onLocationRowClick"
          >
            <template v-if="isEditable && featPlaces" #append>
              <v-btn
                icon="mdi-pencil-outline"
                density="compact"
                variant="plain"
                size="x-small"
                class="meta-inline-pencil meta-inline-pencil--location"
                :title="$gettext('Edit')"
                @click.stop.prevent="openLocationDialog"
              ></v-btn>
            </template>
          </v-list-item>
          <v-list-item v-if="featPlaces && model.Lat && model.Lng" class="mx-0 px-0">
            <p-map :latlng="[model.Lat, model.Lng]" :animate-duration="0" :marker-clickable="isEditable" @marker-clicked="openLocationDialog"></p-map>
          </v-list-item>
        </template>

        <template v-if="canViewPeople && featPeople && (people.length > 0 || isEditable)">
          <v-divider class="my-3"></v-divider>
          <v-list-item class="metadata__item">
            <div class="text-subtitle-2">{{ $gettext("People") }}</div>
            <template v-if="isEditable || people.length > 0" #append>
              <!--
                Editable users get the pencil toggle (edit mode is a
                superset of display, so no separate display toggle);
                read-only users get the eye toggle. Both modes hide
                lightbox chrome and gate keyboard shortcuts — see
                lightbox.vue isShortcutDisabledInFaceMarkerMode.
              -->
              <v-btn
                v-if="isEditable"
                :icon="markersEdit ? 'mdi-pencil-off-outline' : 'mdi-pencil-outline'"
                density="compact"
                variant="plain"
                size="x-small"
                class="meta-markers-toggle meta-faces-edit"
                :class="{ 'is-active': markersEdit }"
                :title="markersEdit ? $gettext('Done') : $gettext('Edit')"
                :disabled="markersBusy"
                @mousedown.prevent
                @click.stop="onToggleFaceMarkerEdit"
              ></v-btn>
              <v-btn
                v-else
                :icon="markersVisible ? 'mdi-eye-off-outline' : 'mdi-eye-outline'"
                density="compact"
                variant="plain"
                size="x-small"
                class="meta-markers-toggle meta-faces-display"
                :class="{ 'is-active': markersVisible }"
                :title="markersVisible ? $gettext('Hide face markers') : $gettext('Show face markers')"
                :disabled="markersBusy"
                @mousedown.prevent
                @click.stop="onToggleFaceMarkerMode"
              ></v-btn>
            </template>
          </v-list-item>
          <v-list-item
            v-for="m in people"
            :key="m.UID || m.CropID"
            :data-marker-uid="m.UID"
            class="metadata__item metadata__person-row"
            @mouseenter="faceMarkers.active && faceMarkers.setHoveredMarkerUid(m.UID)"
            @mouseleave="faceMarkers.active && faceMarkers.setHoveredMarkerUid('')"
          >
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
              :list-props="chipListProps"
              :readonly="markersBusy || !!m.SubjUID"
              :rules="rules.text(false, 0, SubjectMaxLength.Name, $gettext('Name'))"
              return-object
              hide-no-data
              hide-details="auto"
              single-line
              open-on-clear
              append-icon=""
              :menu-icon="null"
              density="compact"
              bg-color="background"
              autocomplete="off"
              class="meta-inline-edit meta-inline-marker"
              :class="{ 'meta-inline-marker--named': m.SubjUID }"
              @focus="loadChipOptions('people')"
              @update:model-value="(v) => onPickPerson(m, v)"
              @update:search="(v) => setMarkerInputValue(m.UID, v)"
              @keydown.enter.stop.prevent="confirmMarkerName(m, 'enter')"
              @keydown.escape.stop.prevent="cancelMarkerName(m)"
              @blur="confirmMarkerName(m, 'blur')"
              @click.stop
            >
              <template v-if="m.SubjUID && markersEdit" #append-inner>
                <v-btn
                  :disabled="markersBusy"
                  :title="$gettext('Unassign')"
                  icon="mdi-eject"
                  density="compact"
                  variant="plain"
                  size="x-small"
                  class="meta-marker-clear-subject"
                  @mousedown.prevent
                  @click.stop="onClearSubject(m)"
                ></v-btn>
              </template>
            </v-combobox>
            <v-list-item-title v-else-if="m.Name" class="meta-person__name">{{ m.Name }}</v-list-item-title>
            <v-list-item-title v-else class="meta-person__name meta-person__unnamed">{{ $gettext("Unknown") }}</v-list-item-title>
          </v-list-item>
        </template>

        <template v-if="canViewAlbums && (albums.length > 0 || isEditable)">
          <v-divider class="my-3"></v-divider>
          <v-list-item class="metadata__item meta-albums">
            <div class="text-subtitle-2">{{ $gettext("Albums") }}</div>
            <template v-if="isEditable && chipState.albums.removals.length > 0" #append>
              <p-lightbox-sidebar-toolbar :editing="true" :can-undo="true" chip-mode @confirm="confirmAlbums" @undo="undoChipRemovals('albums')" />
            </template>
          </v-list-item>
          <v-list-item v-if="visibleAlbums.length > 0" class="metadata__item metadata__chips meta-albums">
            <div class="d-flex flex-wrap ga-1">
              <span
                v-for="a in visibleAlbums"
                :key="a.UID"
                tabindex="0"
                class="meta-chip meta-chip--primary"
                :title="a.Title"
                @click.stop.prevent="onChipActivate('albums', a)"
                @keydown.enter.stop.prevent="onChipActivate('albums', a)"
                @keydown.delete.stop.prevent="onChipDelete('albums', a)"
              >
                <span class="meta-chip__label text-truncate">{{ a.Title }}</span>
                <button
                  v-if="isEditable"
                  type="button"
                  tabindex="-1"
                  class="ms-1 meta-chip__remove meta-icon-btn"
                  :title="$gettext('Remove')"
                  :aria-label="$gettext('Remove')"
                  @mousedown.prevent
                  @click.stop.prevent="onChipDelete('albums', a)"
                >
                  <v-icon icon="mdi-close-circle" size="x-small"></v-icon>
                </button>
              </span>
            </div>
          </v-list-item>
          <v-list-item v-if="isEditable" class="metadata__item meta-albums">
            <v-combobox
              :key="chipState.albums.key"
              v-model="chipState.albums.input"
              v-model:search="chipState.albums.search"
              :items="availableAlbumOptions"
              item-title="Title"
              item-value="Title"
              return-object
              :placeholder="$gettext('Select or create albums')"
              hide-details
              hide-no-data
              single-line
              append-icon=""
              :menu-icon="null"
              density="compact"
              bg-color="background"
              autocomplete="off"
              :menu-props="chipMenuProps"
              :list-props="chipListProps"
              class="meta-inline-edit"
              @focus="loadChipOptions('albums')"
              @update:model-value="onAlbumSelected"
              @keydown.enter.stop.prevent="onAlbumEnter"
              @keydown.escape.stop.prevent="onChipEscape('albums')"
            ></v-combobox>
          </v-list-item>
        </template>

        <template v-if="canViewLabels && (labels.length > 0 || isEditable)">
          <v-divider class="my-3"></v-divider>
          <v-list-item class="metadata__item meta-labels">
            <div class="text-subtitle-2">{{ $gettext("Labels") }}</div>
            <template v-if="isEditable && chipState.labels.removals.length > 0" #append>
              <p-lightbox-sidebar-toolbar :editing="true" :can-undo="true" chip-mode @confirm="confirmLabels" @undo="undoChipRemovals('labels')" />
            </template>
          </v-list-item>
          <v-list-item v-if="visibleLabels.length > 0" class="metadata__item metadata__chips meta-labels">
            <div class="d-flex flex-wrap ga-1">
              <span
                v-for="l in visibleLabels"
                :key="l.Label.UID"
                tabindex="0"
                class="meta-chip meta-chip--primary"
                :title="l.Label.Name"
                @click.stop.prevent="onChipActivate('labels', l)"
                @keydown.enter.stop.prevent="onChipActivate('labels', l)"
                @keydown.delete.stop.prevent="onChipDelete('labels', l)"
              >
                <span class="meta-chip__label text-truncate">{{ l.Label.Name }}</span>
                <button
                  v-if="isEditable"
                  type="button"
                  tabindex="-1"
                  class="ms-1 meta-chip__remove meta-icon-btn"
                  :title="$gettext('Remove')"
                  :aria-label="$gettext('Remove')"
                  @mousedown.prevent
                  @click.stop.prevent="onChipDelete('labels', l)"
                >
                  <v-icon icon="mdi-close-circle" size="x-small"></v-icon>
                </button>
              </span>
            </div>
          </v-list-item>
          <v-list-item v-if="isEditable" class="metadata__item meta-labels">
            <v-combobox
              :key="chipState.labels.key"
              v-model="chipState.labels.input"
              v-model:search="chipState.labels.search"
              :items="availableLabelOptions"
              item-title="Name"
              item-value="Name"
              return-object
              :placeholder="$gettext('Select or create labels')"
              hide-details
              hide-no-data
              single-line
              append-icon=""
              :menu-icon="null"
              density="compact"
              bg-color="background"
              autocomplete="off"
              :menu-props="chipMenuProps"
              :list-props="chipListProps"
              class="meta-inline-edit"
              @focus="loadChipOptions('labels')"
              @update:model-value="onLabelSelected"
              @keydown.enter.stop.prevent="onLabelEnter"
              @keydown.escape.stop.prevent="onChipEscape('labels')"
            ></v-combobox>
          </v-list-item>
        </template>

        <template v-if="showDetailsSection">
          <v-divider class="my-3"></v-divider>
          <template v-for="f in detailsFields" :key="f.key">
            <v-divider v-if="f.key === 'notes' && showRightsDivider" class="my-3 meta-rights-divider"></v-divider>
            <v-list-item
              v-show="shouldShowFieldRow(f)"
              v-tooltip="{ text: f.label, disabled: !f.icon || $isMobile }"
              :prepend-icon="f.icon"
              :class="['metadata__item', `meta-${f.key}`, { clickable: editingField !== f.key && (isEditable || f.read(photo)) }]"
              @click.stop="onTextRowClick($event, f.key, f.read(photo))"
            >
              <v-textarea
                v-if="editingField === f.key"
                :ref="setInlineEditorRef"
                :model-value="f.read(photo)"
                :rules="rules.text(false, 0, f.maxLength, f.label)"
                :rows="1"
                density="compact"
                auto-grow
                hide-details="auto"
                autocomplete="off"
                class="meta-inline-edit"
                :class="`meta-inline-${f.key}`"
                @update:model-value="(v) => f.write(photo, v)"
                @keydown.enter.stop="(ev) => onInlineEnter(ev, f)"
                @keydown.escape.stop.prevent="cancelEditing"
                @blur="onInlineFieldBlur"
              ></v-textarea>
              <!-- eslint-disable-next-line vue/no-v-html -- f.htmlValue points at a sanitized computed (e.g. notesHtml) which runs $util.sanitizeHtml($util.encodeHTML(raw)). -->
              <div v-else-if="f.display === 'html' && fieldHtml(f)" class="text-body-2 meta-scrollable text-html" :class="`meta-${f.key}`" v-html="fieldHtml(f)"></div>
              <div v-else-if="f.display !== 'html' && f.read(photo)" class="text-body-2 meta-scrollable" :class="`meta-${f.key}`">{{ f.read(photo) }}</div>
              <div v-else class="meta-add-prompt" @click.stop="startEditing(f.key)">{{ f.placeholder ? f.placeholder : f.label }}</div>
              <template v-if="isEditable" #append>
                <p-lightbox-sidebar-toolbar
                  :editing="editingField === f.key"
                  :can-undo="editingField === f.key"
                  :undo-disabled="!inlineEditDirty"
                  @confirm="confirmField"
                  @undo="undoInlineEdit"
                  @start="startEditing(f.key)"
                />
              </template>
            </v-list-item>
          </template>
        </template>
      </v-list>
    </div>
    <p-meta-datetime-dialog :visible="dateTimeDialog" :photo="photo" @close="dateTimeDialog = false" @confirm="confirmDateTime"></p-meta-datetime-dialog>
    <p-meta-camera-dialog :visible="cameraDialog" :photo="photo" @close="cameraDialog = false" @confirm="confirmCamera"></p-meta-camera-dialog>
    <p-meta-location-dialog
      :visible="locationDialog"
      :latlng="[photo ? Number(photo.Lat) || 0 : 0, photo ? Number(photo.Lng) || 0 : 0]"
      @close="locationDialog = false"
      @confirm="confirmLocation"
    ></p-meta-location-dialog>
    <p-confirm-dialog
      :visible="discardDialog.visible"
      icon="mdi-alert-circle-outline"
      :text="discardDialogText"
      :action="$gettext('Discard')"
      @close="onDiscardCancel"
      @confirm="onDiscardConfirm"
    ></p-confirm-dialog>
    <p-confirm-dialog
      :visible="addNameDialog.visible"
      icon="mdi-account-plus"
      :icon-size="42"
      :text="addNameDialog.name ? $gettext('Add %{s}?', { s: addNameDialog.name }) : $gettext('Add person?')"
      @close="onAddNameCancel"
      @confirm="onAddNameConfirm"
    ></p-confirm-dialog>
  </div>
</template>

<script>
import { DateTime } from "luxon";
import * as formats from "options/formats";
import { $faceMarkers } from "common/face-markers";

import * as media from "common/media";
import { is360Equirectangular } from "common/sphere";
import typeaheadCache from "common/typeahead-cache";
import { rules } from "common/form";
import { Album, MaxLength as AlbumMaxLength } from "model/album";
import { MaxLength as LabelMaxLength } from "model/label";
import { MaxLength as PhotoMaxLength } from "model/photo";
import { MaxLength as SubjectMaxLength } from "model/subject";
import PMap from "component/map.vue";
import PMetaDatetimeDialog from "component/meta/datetime/dialog.vue";
import PMetaCameraDialog from "component/meta/camera/dialog.vue";
import PMetaLocationDialog from "component/meta/location/dialog.vue";
import PConfirmDialog from "component/confirm/dialog.vue";
import PLightboxSidebarToolbar from "component/lightbox/sidebar/toolbar.vue";

export default {
  name: "PLightboxSidebar",
  components: {
    PMap,
    PMetaDatetimeDialog,
    PMetaCameraDialog,
    PMetaLocationDialog,
    PConfirmDialog,
    PLightboxSidebarToolbar,
  },
  props: {
    // UID of the photo currently in the parent lightbox; drives the
    // slide-change lifecycle. All other parent state flows through
    // `view` (data() below).
    uid: {
      type: String,
      default: "",
    },
  },
  emits: ["close", "toggle-face-marker-mode", "toggle-face-marker-edit", "clear-subject", "reload-markers", "naming-started"],
  data() {
    return {
      // Reactive handle to the parent lightbox's $data via `$view.getData()`.
      // Mutations through `this.view.X` write through to the parent.
      view: this.$view.getData(),
      // Shared face-marker state singleton. The lightbox owns policy
      // (transitions, API writes); the sidebar emits `toggle-face-marker-*`
      // / `clear-subject` / `reload-markers`.
      faceMarkers: $faceMarkers,
      actions: [],
      featPeople: this.$config.feature("people"),
      featPlaces: this.$config.feature("places"),
      rules,
      SubjectMaxLength,
      dateTimeDialog: false,
      cameraDialog: false,
      locationDialog: false,
      // CSS toggles for the inline pencil / Undo / Save affordances —
      // hidden by default since row click + keyboard cover everything.
      // Flip per-user / A/B to surface the mouse-driven affordances.
      hideEditPencils: true,
      hideEditUndo: true,
      hideEditSave: true,
      editingField: null,
      editOriginal: null,
      // Per-field combobox state: input/search drive v-model; `search`
      // doubles as the typed-but-not-yet-Enter pending detector; `key`
      // force-remounts to clear stale dropdown state; `options` is the
      // typeahead list; `removals` queues IDs for the toolbar ✓ (adds
      // skip this and take the instant-save path).
      chipState: {
        labels: { input: null, search: "", key: 0, options: [], removals: [] },
        albums: { input: null, search: "", key: 0, options: [], removals: [] },
        // People is read-only suggestions (no removals/typed-text editing); the
        // full shape keeps the chipState iterations (pending-edit checks) uniform.
        people: { input: null, search: "", key: 0, options: [], removals: [] },
      },
      markerDrafts: {},
      markerMenuProps: {
        openOnFocus: true,
        closeOnContentClick: true,
        maxHeight: 260,
        class: "meta-inline-menu",
      },
      chipMenuProps: {
        class: "meta-inline-menu",
      },
      // Forwarded to the inner v-list of the combobox/autocomplete dropdown.
      // density="compact" on the input itself only sizes the trigger field —
      // list-props is the documented way to size the menu items themselves.
      chipListProps: {
        density: "compact",
      },
      discardDialog: {
        visible: false,
        resolver: null,
      },
      // Stores the marker UID, not the Marker instance: the live Marker
      // is re-derived from photo.getMarkers(true) at commit time so a
      // slide-nav between open and Add/Cancel can't write through a
      // stale reference.
      addNameDialog: {
        visible: false,
        markerUid: "",
        name: "",
      },
    };
  },
  computed: {
    // Aliases for parent-owned reactive state. Reads go through these;
    // mutations must target `this.view.photo.X` / `this.view.model.X`
    // directly, never the aliases.
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
    // Booleans derived from the face-marker singleton
    // (`common/face-markers.js`): null = off, 'display' = read-only,
    // 'edit' = drag-to-create + click-to-remove.
    markersVisible() {
      return this.faceMarkers.active;
    },
    markersEdit() {
      return this.faceMarkers.isEdit;
    },
    markersBusy() {
      return this.faceMarkers.busy;
    },
    newMarkerUid() {
      return this.faceMarkers.pendingNameMarkerUid;
    },
    isEditable() {
      return this.canEdit && this.photo && this.photo.Details && this.canViewLibrary;
    },
    // True for full library roles (admin/user/manager/viewer), false
    // for visitor/guest/contributor. Gates photographer-EXIF privacy
    // fields (camera, lens, file name) and the rights/notes cluster.
    canViewLibrary() {
      return this.$config.allow("photos", "access_library");
    },
    // canViewPeople gates the People section on the backend's people
    // resource grant. People is admin/client-only in CE rules (see
    // internal/auth/acl/rules.go); Plus/Pro/Portal may extend.
    canViewPeople() {
      return this.$config.allow("people", "search");
    },
    // canViewLabels gates the Labels section on the backend's labels
    // resource grant. Same shape as canViewPeople.
    canViewLabels() {
      return this.$config.feature("labels") && this.$config.allow("labels", "search");
    },
    // Gates the Albums section. Visitors/guests hold `access_shared`,
    // so they see albums on shared photos even without people/labels.
    canViewAlbums() {
      return this.$config.feature("albums") && this.$config.allow("albums", "search");
    },
    // Gates the place name / altitude / location row. Visitors/guests
    // hold view access via GrantViewShared / GrantReactShared, matching
    // the backend's redaction policy on shared photos.
    canViewPlaces() {
      return this.$config.allow("places", "view");
    },
    captionHtml() {
      // `||` not `??`: limited-access sessions skip the Photo fetch, leaving
      // `photo.Caption` as the empty-string default — must fall through to
      // `model.Caption`.
      const raw = this.photo?.Caption || this.model?.Caption;
      if (!raw) {
        return "";
      }
      return this.$util.sanitizeHtml(this.$util.encodeHTML(raw));
    },
    notesHtml() {
      if (!this.photo?.Details?.Notes) {
        return "";
      }
      return this.$util.sanitizeHtml(this.$util.encodeHTML(this.photo.Details.Notes));
    },
    cameraInfo() {
      if (!this.photo) {
        return "";
      }
      // Backend returns the "Unknown" placeholder camera (CameraID=1)
      // when no EXIF camera is set; suppress so the row doesn't render
      // as " Unknown".
      const hasRealCamera =
        (this.photo.CameraID && this.photo.CameraID > 1) ||
        (this.photo.CameraMake && this.photo.CameraMake.trim()) ||
        (this.photo.CameraModel && this.photo.CameraModel.trim() && this.photo.CameraModel !== "Unknown");
      if (!hasRealCamera) {
        return "";
      }
      // Suppress "Unknown, ISO 100"-style rows when only ISO/exposure are set.
      if (!this.$util.formatCamera(this.photo.Camera, this.photo.CameraID, this.photo.CameraMake, this.photo.CameraModel, false)) {
        return "";
      }
      const info = this.photo.getCameraInfo();
      return info !== this.$gettext("Unknown") ? info : "";
    },
    lensInfo() {
      if (!this.photo) {
        return "";
      }
      const hasLens =
        (this.photo.LensID && this.photo.LensID > 1) || this.photo.LensMake || this.photo.LensModel || this.photo.Lens?.Model || this.photo.Lens?.Make;
      if (!hasLens) {
        return "";
      }
      const info = this.photo.getLensInfo();
      return info !== this.$gettext("Unknown") ? info : "";
    },
    exifInfo() {
      if (!this.photo) {
        return "";
      }
      return this.photo.getExifInfo();
    },
    people() {
      if (!this.photo) {
        return [];
      }
      return this.photo.getMarkers(true);
    },
    // Name suggestions for the marker combobox, loaded from the shared people
    // cache via loadChipOptions("people") and sorted there; the cache evicts on
    // people.* / subjects.* WS events so the list stays current.
    knownPeople() {
      return this.chipState.people.options;
    },
    labels() {
      if (!this.photo?.Labels) {
        return [];
      }
      // Sort by name — the backend orders by uncertainty/topicality but
      // the sidebar doesn't surface those scores.
      return this.photo.Labels.filter((l) => l.Label && l.Label.Name && l.Uncertainty < 100)
        .slice()
        .sort((a, b) => (a.Label.Name || "").localeCompare(b.Label.Name || "", undefined, { sensitivity: "base", numeric: true }));
    },
    albums() {
      if (!this.photo?.Albums) {
        return [];
      }
      return this.photo.Albums.filter((a) => a.Title && !a.Private)
        .slice()
        .sort((a, b) => (a.Title || "").localeCompare(b.Title || "", undefined, { sensitivity: "base", numeric: true }));
    },
    // Typeahead options minus items already on the photo — matches
    // batch-edit / Edit Dialog. add*Immediate still consults the
    // unfiltered options for canonical-name matching.
    availableLabelOptions() {
      const normalize = (s) => this.$util.normalizeTitle(s || "");
      const assigned = new Set(this.labels.map((l) => normalize(l?.Label?.Name)).filter(Boolean));
      return this.chipState.labels.options.filter((opt) => !assigned.has(normalize(opt.Name)));
    },
    availableAlbumOptions() {
      const normalize = (s) => this.$util.normalizeTitle(s || "");
      const assigned = new Set(this.albums.map((a) => normalize(a?.Title)).filter(Boolean));
      return this.chipState.albums.options.filter((opt) => !assigned.has(normalize(opt.Title)));
    },
    // Visible chips = primary list minus soft-removed entries. The
    // chip-row wrapper gates on these so it collapses once every chip
    // is removed; Undo restores by clearing `chipState.<field>.removals`.
    visibleLabels() {
      return this.labels.filter((l) => !this.isChipPendingRemoval("labels", l?.Label?.ID));
    },
    visibleAlbums() {
      return this.albums.filter((a) => !this.isChipPendingRemoval("albums", a?.UID));
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
    // fieldRegistry is the single source of truth for inline-text fields:
    // raw read/write, labels, per-field maxLength (from PhotoMaxLength), and
    // whether the display branch treats the value as sanitized HTML.
    fieldRegistry() {
      return {
        title: {
          key: "title",
          label: this.$pgettext("Photo", "Title"),
          read: (p) => p?.Title,
          write: (p, v) => {
            if (p) {
              p.Title = v;
            }
          },
          display: "text",
          maxLength: PhotoMaxLength.Title,
        },
        caption: {
          key: "caption",
          label: this.$gettext("Caption"),
          read: (p) => p?.Caption,
          write: (p, v) => {
            if (p) {
              p.Caption = v;
            }
          },
          display: "html",
          htmlValue: "captionHtml",
          maxLength: PhotoMaxLength.Caption,
        },
        subject: {
          key: "subject",
          label: this.$gettext("Subject"),
          icon: "mdi-flower-tulip",
          read: (p) => p?.Details?.Subject,
          write: (p, v) => {
            if (p?.Details) {
              p.Details.Subject = v;
            }
          },
          display: "text",
          maxLength: PhotoMaxLength.Subject,
          commitOnEnter: true,
        },
        copyright: {
          key: "copyright",
          label: this.$gettext("Copyright"),
          icon: "mdi-copyright",
          read: (p) => p?.Details?.Copyright,
          write: (p, v) => {
            if (p?.Details) {
              p.Details.Copyright = v;
            }
          },
          display: "text",
          maxLength: PhotoMaxLength.Copyright,
          commitOnEnter: true,
        },
        artist: {
          key: "artist",
          label: this.$gettext("Artist"),
          icon: "mdi-account-tie",
          read: (p) => p?.Details?.Artist,
          write: (p, v) => {
            if (p?.Details) {
              p.Details.Artist = v;
            }
          },
          display: "text",
          maxLength: PhotoMaxLength.Artist,
          commitOnEnter: true,
        },
        license: {
          key: "license",
          label: this.$gettext("License"),
          icon: "mdi-scale-balance",
          read: (p) => p?.Details?.License,
          write: (p, v) => {
            if (p?.Details) {
              p.Details.License = v;
            }
          },
          display: "text",
          maxLength: PhotoMaxLength.License,
          commitOnEnter: true,
        },
        keywords: {
          key: "keywords",
          label: this.$gettext("Keywords"),
          icon: "mdi-tag-multiple-outline",
          read: (p) => p?.Details?.Keywords,
          write: (p, v) => {
            if (p?.Details) {
              p.Details.Keywords = v;
            }
          },
          display: "text",
          maxLength: PhotoMaxLength.Keywords,
          commitOnEnter: true,
        },
        notes: {
          key: "notes",
          label: this.$gettext("Notes"),
          placeholder: this.$gettext("Add Notes"),
          icon: null,
          read: (p) => p?.Details?.Notes,
          write: (p, v) => {
            if (p?.Details) {
              p.Details.Notes = v;
            }
          },
          display: "html",
          htmlValue: "notesHtml",
          maxLength: PhotoMaxLength.Notes,
        },
      };
    },
    detailsFields() {
      return ["subject", "copyright", "artist", "license", "keywords", "notes"].map((k) => this.fieldRegistry[k]);
    },
    showDetailsSection() {
      if (!this.canViewLibrary) {
        return false;
      }
      if (this.isEditable) {
        return true;
      }
      return this.detailsFields.some((f) => Boolean(f.read(this.photo)));
    },
    // True when the active inline editor's value differs from the
    // editOriginal snapshot — gates the Undo button's disabled state.
    inlineEditDirty() {
      if (!this.editingField) {
        return false;
      }
      return this.getFieldValue(this.editingField) !== (this.editOriginal ?? "");
    },
    // Discard-dialog body text: shifts to "Discard invalid changes?"
    // when the only pending state is an overlength inline edit, so the
    // user knows which kind of change they're abandoning.
    discardDialogText() {
      if (this.hasPendingInlineOverflow() && !this.hasPendingNonOverflowEdit()) {
        return this.$gettext("Discard invalid changes?");
      }
      return this.$gettext("Discard unsaved changes?");
    },
    // Renders the divider between the rights cluster and the Notes row.
    // For read-only users we drop it when either side is empty so it
    // doesn't appear as an orphan line.
    showRightsDivider() {
      if (this.isEditable) {
        return true;
      }
      const hasAbove = ["subject", "artist", "copyright", "license", "keywords"].some((k) => this.shouldShowFieldRow(this.fieldRegistry[k]));
      const hasBelow = this.shouldShowFieldRow(this.fieldRegistry.notes);
      return hasAbove && hasBelow;
    },
    placeName() {
      // Empty-string contract lives in the Photo model so other views
      // can reuse it; see `Photo.placeName()`.
      return this.photo?.placeName?.() || "";
    },
    altitude() {
      if (!this.photo || !this.photo.Altitude) {
        return "";
      }
      return this.photo.Altitude + " m";
    },
    // Returns the lat/lng (shortened, with optional altitude) for the
    // combined Location row. Users without places-view ACL see only the
    // lat/lng so altitude isn't leaked through the sidebar.
    coordinatesLine() {
      if (!this.model?.Lat || !this.model?.Lng) {
        return "";
      }
      const coords = this.model.getLatLngShort();
      if (this.altitude && this.canViewPlaces) {
        return `${coords}\u2002${this.altitude}`;
      }
      return coords;
    },
    // True when the combined Location row should render — i.e., the row
    // has coordinates to display, OR the user holds places-view ACL and
    // has a place name / can edit a missing location.
    locationRowVisible() {
      if (this.model?.Lat && this.model?.Lng) {
        return true;
      }
      if (!this.canViewPlaces) {
        return false;
      }
      if (this.placeName) {
        return true;
      }
      return this.isEditable && this.featPlaces;
    },
    // Returns the merged row's title: place name when the user holds
    // places-view ACL, else coordinates so the row never renders empty.
    locationTitle() {
      if (this.canViewPlaces && this.placeName) {
        return this.placeName;
      }
      if (this.coordinatesLine) {
        return this.coordinatesLine;
      }
      return this.$gettext("Unknown");
    },
    // Returns the merged row's subtitle: the coordinates line, but only
    // when the title already shows the place name (so we don't render
    // the coordinates twice on a no-places-ACL / no-placeName row).
    locationSubtitle() {
      if (!this.canViewPlaces) {
        return null;
      }
      if (this.placeName && this.coordinatesLine) {
        return this.coordinatesLine;
      }
      return null;
    },
    // True when the combined Location row has any click action: editing
    // the location, copying the coordinates, or copying the place name.
    // Drives the .clickable cursor class on the row.
    locationRowClickable() {
      if (this.isEditable && this.featPlaces) {
        return true;
      }
      if (this.model?.Lat && this.model?.Lng) {
        return true;
      }
      return this.canViewPlaces && !!this.placeName;
    },
    // Prefers originalFile() so video / Live / Animated show .mp4 etc.
    // instead of the JPEG cover. Returns null (not "") so Vuetify's
    // `:subtitle != null` gate hides the slot cleanly.
    fileName() {
      if (!this.canViewLibrary || !this.photo) {
        return null;
      }
      if (typeof this.photo.originalFile === "function") {
        const original = this.photo.originalFile();
        if (original && original !== this.photo && original.Name) {
          return original.Name;
        }
      }
      if (this.photo.FileName) {
        return this.photo.FileName;
      }
      const primary = typeof this.photo.primaryFile === "function" ? this.photo.primaryFile() : null;
      return primary?.Name || null;
    },
    fileInfo() {
      // Gate on UID so the limited-access empty-Photo placeholder falls
      // through to the Thumb helper instead of rendering "Unknown".
      if (this.photo && this.photo.UID) {
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
      // Thumb.getTypeInfo returns "" when no codec/dimensions/duration
      // are set, so a truthy check hides the row cleanly.
      if (this.model && typeof this.model.getTypeInfo === "function") {
        return this.model.getTypeInfo() || "";
      }
      return "";
    },
    // mediaType returns the active media type for icon / label lookup.
    // Gate on UID: Photo.getDefaults() seeds an empty Photo with Type=image,
    // which would mask the Thumb's real type for limited-access sessions.
    mediaType() {
      if (this.photo && this.photo.UID && this.photo.Type) {
        return this.photo.Type;
      }
      return this.model?.Type || "";
    },
    // mediaIs360 reports whether the active media is equirectangular 360° content,
    // reusing the lightbox slide-routing discriminator (is360Equirectangular) so the
    // file icon matches what opens in the sphere viewer. Checks both the thumb model
    // and the photo since either may carry the projection / dimensions.
    mediaIs360() {
      return is360Equirectangular(this.model) || is360Equirectangular(this.photo);
    },
    fileIcon() {
      if (this.mediaIs360) {
        return "mdi-panorama-variant-outline";
      }
      switch (this.mediaType) {
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
    // Localized media-type label for the file row tooltip, with a
    // generic "File" fallback when Type is missing.
    fileTypeName() {
      return this.$util.typeName(this.mediaType, this.$gettext("File"));
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
      if (!uid) {
        return;
      }
      this.$nextTick(() => this.focusMarkerInput(uid));
    },
  },
  mounted() {
    // Warm the typeahead cache so chip comboboxes are populated when
    // focused. The shared cache (common/typeahead-cache.js) dedupes
    // concurrent callers, so remounts during batch-edit cost nothing.
    if (this.isEditable) {
      if (this.canViewLabels) {
        this.loadChipOptions("labels");
      }
      if (this.canViewAlbums) {
        this.loadChipOptions("albums");
      }
      if (this.canViewPeople) {
        this.loadChipOptions("people");
      }
    }
  },
  methods: {
    close() {
      this.$emit("close");
    },
    // openDateTimeDialog mounts the date-and-time-edit dialog when the
    // session is editable. No-op otherwise, so callers (row @click and
    // pencil button) don't need to inline the gate.
    openDateTimeDialog() {
      if (!this.isEditable) {
        return;
      }
      this.dateTimeDialog = true;
    },
    // openCameraDialog mounts the camera-and-lens-edit dialog when the
    // session is editable. Shared by the Camera and Lens row icons and
    // the Camera pencil button.
    openCameraDialog() {
      if (!this.isEditable) {
        return;
      }
      this.cameraDialog = true;
    },
    // openLocationDialog mounts the location-edit dialog when the
    // session is editable and the `places` feature is enabled. Shared by
    // the Location row pencil and the row click in edit mode.
    openLocationDialog() {
      if (!this.isEditable || !this.featPlaces) {
        return;
      }
      this.locationDialog = true;
    },
    // onLocationRowClick: edit mode opens the location dialog;
    // read-only copies coordinates if present, else the place name.
    // Coordinates-first preserves the existing copy-coords gesture.
    onLocationRowClick() {
      if (this.isEditable && this.featPlaces) {
        this.openLocationDialog();
      } else if (this.model?.Lat && this.model?.Lng) {
        this.model.copyLatLng();
      } else if (this.placeName) {
        this.$util.copyText(this.placeName);
      }
    },
    // onTextRowClick enters the inline editor (edit mode) or copies the
    // value to the clipboard (read-only). Bails on link clicks so the
    // browser follows the href in v-html fields like caption and notes.
    onTextRowClick(ev, field, value) {
      if (ev?.target?.closest?.("a") || this.editingField === field) {
        return;
      }

      if (this.isEditable) {
        this.startEditing(field);
      } else if (value) {
        this.$util.copyText(value);
      }
    },
    getFieldValue(field) {
      const f = this.fieldRegistry[field];
      if (!f) {
        return "";
      }
      const v = f.read(this.photo);
      return v == null ? "" : v;
    },
    setFieldValue(field, value) {
      const f = this.fieldRegistry[field];
      if (!f || !this.view?.photo) {
        return;
      }
      f.write(this.view.photo, value);
    },
    // Function ref shared by every inline editor. Only one is mounted
    // at a time (gated by `editingField === '<key>'`), so the latest
    // non-null call identifies the active editor.
    setInlineEditorRef(el) {
      if (el) {
        this._inlineEditorEl = el;
      } else if (!this.editingField) {
        this._inlineEditorEl = null;
      }
    },
    fieldHtml(f) {
      if (!f || f.display !== "html" || !f.htmlValue) {
        return "";
      }
      return this[f.htmlValue] || "";
    },
    shouldShowFieldRow(f) {
      if (!f) {
        return false;
      }
      if (this.isEditable || this.editingField === f.key) {
        return true;
      }
      if (f.display === "html") {
        return Boolean(this.fieldHtml(f));
      }
      return Boolean(f.read(this.photo));
    },
    // Plain Enter commits when the field opts in via `commitOnEnter`
    // (single-line fields); otherwise it falls through to insert a
    // newline. Shift+Enter always inserts a newline. Propagation is
    // stopped by the `.stop` modifier on the binding, so the keystroke
    // never reaches the dialog's Enter handler in either branch.
    onInlineEnter(ev, f) {
      if (!f?.commitOnEnter || ev.shiftKey) {
        return;
      }
      ev.preventDefault();
      this.confirmField();
    },
    // Reverts the active inline editor to editOriginal without exiting
    // edit mode (cancelEditing/Escape also exits). Wired to the toolbar
    // Undo button for mouse parity with the keyboard cancel.
    undoInlineEdit() {
      if (!this.editingField || this.editOriginal === null) {
        return;
      }
      this.setFieldValue(this.editingField, this.editOriginal);
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
        if (editor && typeof editor.focus === "function") {
          editor.focus();
        }
      });
    },
    // Eye-icon handler: flips between `null` and `FaceMarkerDisplay`.
    // Gates only on `markersBusy`, not `isEditable`, because display
    // mode doesn't require edit permission.
    onToggleFaceMarkerMode() {
      if (this.markersBusy) {
        return;
      }
      this.$emit("toggle-face-marker-mode");
    },
    // Pencil-icon handler: flips between `null` and `FaceMarkerEdit`,
    // which adds drag-to-create + click-to-remove.
    onToggleFaceMarkerEdit() {
      if (!this.isEditable || this.markersBusy) {
        return;
      }
      this.$emit("toggle-face-marker-edit");
    },
    onClearSubject(marker) {
      if (!this.isEditable || this.markersBusy || !marker || !marker.SubjUID) {
        return;
      }
      // Without this timestamp, the implicit @blur from the combobox
      // re-render after clearSubject would commit the typed name the
      // user just rejected. confirmMarkerName checks it and bails.
      this._lastDestructiveMarkerActionAt = Date.now();
      this.$emit("clear-subject", marker);
    },
    // Combobox can bind either the typed string or the selected subject object.
    unwrapMarkerName(value) {
      return typeof value === "object" && value !== null ? value.Name || "" : value || "";
    },
    // Reconciles markerDrafts with the latest markers on every
    // photo-cache mutation. The `editing` flag prevents an unrelated
    // WS update from clobbering text the user is currently typing;
    // confirmMarkerName / cancelMarkerName / onAddNameConfirm clear it.
    syncMarkerDrafts(markers) {
      const seen = new Set();
      markers.forEach((m) => {
        if (!m || !m.UID) {
          return;
        }

        seen.add(m.UID);

        const original = m.Name || "";
        const existing = this.markerDrafts[m.UID];

        if (!existing) {
          this.markerDrafts[m.UID] = { original, current: original, editing: false };
        } else if (existing.original !== original) {
          existing.original = original;
          if (!existing.editing) {
            existing.current = original;
          }
        }
      });
      Object.keys(this.markerDrafts).forEach((uid) => {
        if (!seen.has(uid)) {
          delete this.markerDrafts[uid];
        }
      });
    },
    markerInputValue(uid) {
      const d = this.markerDrafts[uid];
      return d ? d.current : "";
    },
    markerInputSearch(uid) {
      return this.unwrapMarkerName(this.markerInputValue(uid));
    },
    // Records typed text into the per-marker draft and sets `editing`
    // so a concurrent WS update can't snap the input back mid-keystroke.
    // commit / cancel / onAddName* clear the flag at settled states.
    setMarkerInputValue(uid, value) {
      if (!uid) {
        return;
      }
      if (!this.markerDrafts[uid]) {
        this.markerDrafts[uid] = { original: "", current: value, editing: true };
      } else {
        this.markerDrafts[uid].current = value;
        this.markerDrafts[uid].editing = true;
      }
    },
    focusMarkerInput(uid) {
      if (!uid) {
        return;
      }
      this.$emit("naming-started");
      this.$nextTick(() => {
        const input = this.$el && this.$el.querySelector(`[data-marker-uid="${uid}"] input`);
        if (input) {
          input.focus();
        }
      });
    },
    // Locale-aware case-insensitive match so the backend doesn't create
    // a duplicate when the typed name differs only in case (handles
    // Turkish i, German ß, Cyrillic, Hebrew, etc.).
    findKnownPerson(name) {
      if (!name) {
        return null;
      }
      return this.knownPeople.find((p) => p && p.Name && p.Name.localeCompare(name, undefined, { sensitivity: "base" }) === 0) || null;
    },
    // Resolves a marker by UID from the current photo. Used by Add-name
    // confirm so a stale Marker reference can't write through after the
    // marker was rejected or the slide moved on.
    findMarker(uid) {
      if (!uid || !this.photo || typeof this.photo.getMarkers !== "function") {
        return null;
      }
      return this.photo.getMarkers(true).find((m) => m && m.UID === uid) || null;
    },
    // Commits typed text from the per-marker draft. "blur" routes
    // unnamed markers through Add-name confirmation. Bails when busy,
    // invalid, or a destructive icon fired in the last 200 ms.
    confirmMarkerName(marker, source = "enter") {
      if (!marker || !marker.UID) {
        return;
      }
      if (this.markersBusy || marker.Invalid) {
        return;
      }
      if (this._lastDestructiveMarkerActionAt && Date.now() - this._lastDestructiveMarkerActionAt < 200) {
        return;
      }
      const draft = this.markerDrafts[marker.UID];
      if (!draft) {
        return;
      }
      const name = this.unwrapMarkerName(draft.current).trim();
      const original = (draft.original || "").trim();

      if (!name || name === original) {
        // Reaching here means the user blurred without changing anything
        // (or restored the original). The draft is settled; clear the
        // editing flag so a concurrent WS update can re-sync.
        draft.editing = false;
        return;
      }

      if (typeof marker.setName !== "function") {
        return;
      }

      const match = this.findKnownPerson(name);

      // Blur without Enter on an unnamed marker → ask before committing a new
      // name. Skip the dialog if the person already exists (match) or if the
      // marker is already named (eject/rename path) — both are unambiguous.
      if (source === "blur" && !marker.SubjUID && !match) {
        this.addNameDialog = { visible: true, markerUid: marker.UID, name };
        return;
      }

      this.commitMarkerName(marker, match, name);
    },
    commitMarkerName(marker, match, name) {
      const draft = this.markerDrafts[marker.UID];
      if (!draft) {
        return;
      }

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
      draft.editing = false;

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
      if (!marker || !value || typeof value !== "object" || !value.Name) {
        return;
      }
      this.setMarkerInputValue(marker.UID, value.Name);
      this.confirmMarkerName(marker, "enter");
    },
    // Resolves the stored markerUid against the live photo and drops
    // the commit silently if the marker is gone (rejected, slide moved,
    // etc.) — the dialog already closed.
    onAddNameConfirm() {
      const { markerUid, name } = this.addNameDialog;
      this.addNameDialog = { visible: false, markerUid: "", name: "" };
      if (markerUid && name) {
        const marker = this.findMarker(markerUid);
        if (marker) {
          this.commitMarkerName(marker, this.findKnownPerson(name), name);
        }
      }
      this.resolveAddNameNav();
    },
    onAddNameCancel() {
      const { markerUid } = this.addNameDialog;
      this.addNameDialog = { visible: false, markerUid: "", name: "" };
      const draft = markerUid ? this.markerDrafts[markerUid] : null;
      if (draft) {
        draft.current = draft.original || "";
        draft.editing = false;
      }
      this.resolveAddNameNav();
    },
    // Resolves the nav promise parked by `confirmDiscardPending` when
    // Add-name was open — both Add and Cancel settle the draft.
    resolveAddNameNav() {
      const resolve = this._addNameNavResolver;
      this._addNameNavResolver = null;
      if (resolve) {
        resolve(true);
      }
    },
    cancelMarkerName(marker) {
      if (!marker || !marker.UID) {
        return;
      }
      const draft = this.markerDrafts[marker.UID];
      if (!draft) {
        return;
      }
      draft.current = draft.original;
      draft.editing = false;
      // Blur the marker's own input (scoped, not document.activeElement)
      // for visual feedback. The @blur re-fires confirmMarkerName but
      // it's a no-op since current === original and editing is false.
      const input = this.$el && this.$el.querySelector(`[data-marker-uid="${marker.UID}"] input`);
      if (input && typeof input.blur === "function") {
        input.blur();
      }
    },
    resetInlineEdits() {
      if (this.editingField) {
        this.cancelEditing();
      }

      this.resetChipState();
      this.clearChipInput();

      Object.keys(this.markerDrafts).forEach((uid) => {
        const d = this.markerDrafts[uid];
        if (d) {
          d.current = d.original;
          d.editing = false;
        }
      });

      if (this.addNameDialog && this.addNameDialog.visible) {
        this.addNameDialog = { visible: false, markerUid: "", name: "" };
      }
    },
    // True when an edit is staged that the navigation guard should
    // warn about. Split into overflow (open editor at maxLength+) vs
    // non-overflow so the discard dialog can pick the right copy.
    // Chip removals auto-commit via confirmDiscardPending before this
    // runs, so they're not counted here.
    hasPendingEdit() {
      return this.hasPendingInlineOverflow() || this.hasPendingNonOverflowEdit();
    },
    // Pending state that isn't an overlength inline editor: marker
    // drafts, typed combobox text, chip removals, Add-name dialog.
    // Split out so discardDialogText can pick the right message.
    hasPendingNonOverflowEdit() {
      for (const uid of Object.keys(this.markerDrafts)) {
        const d = this.markerDrafts[uid];
        if (!d) {
          continue;
        }
        if (this.unwrapMarkerName(d.current).trim() !== (d.original || "").trim()) {
          return true;
        }
      }
      // Pending chip removals (× icon) and typed-but-uncommitted combobox
      // text both count: Enter would instant-save, but until then the
      // characters live only in chipState.<field>.search.
      if (Object.values(this.chipState).some((s) => s.removals.length || (s.search || "").trim() !== "")) {
        return true;
      }
      // An open Add-name confirmation for an unnamed marker is also pending
      // input until the user picks Add or Cancel.
      return !!(this.addNameDialog && this.addNameDialog.visible);
    },
    // True when an inline editor is open with a value above the cap.
    // Pairs with confirmField()'s length gate so navigation sources
    // surface the discard dialog instead of dropping the invalid edit.
    hasPendingInlineOverflow() {
      if (!this.editingField || !this.photo) {
        return false;
      }
      const fieldDef = this.fieldRegistry[this.editingField];
      if (!fieldDef || !(fieldDef.maxLength > 0)) {
        return false;
      }
      const currentValue = fieldDef.read(this.photo);
      return typeof currentValue === "string" && currentValue.length > fieldDef.maxLength;
    },
    // Fire-and-forget commit of any pending chip removals. Mirrors the
    // inline-text auto-commit on blur: the user's intent (clicking ×) is
    // honored on navigation/close instead of being silently discarded.
    autoCommitChipRemovals() {
      if (this.chipState.labels.removals.length) {
        this.confirmLabels();
      }
      if (this.chipState.albums.removals.length) {
        this.confirmAlbums();
      }
    },
    // Settles dirty marker drafts as if @blur had fired — keyboard / code
    // navigation skips the input blur, so without this the discard
    // dialog would race the eventual commit.
    flushDirtyMarkerDrafts() {
      Object.keys(this.markerDrafts).forEach((uid) => {
        const draft = this.markerDrafts[uid];
        if (!draft) {
          return;
        }
        const name = this.unwrapMarkerName(draft.current).trim();
        const original = (draft.original || "").trim();
        if (!name || name === original) {
          return;
        }
        const marker = this.findMarker(uid);
        if (marker) {
          this.confirmMarkerName(marker, "blur");
        }
      });
    },
    // Async guard before close / hide / nav. Returns Promise<boolean>:
    // true = safe, false = canceled. Chip removals auto-commit first so
    // the dialog only fires for state the user might still want to keep.
    confirmDiscardPending() {
      this.autoCommitChipRemovals();
      this.flushDirtyMarkerDrafts();
      // Defer to an open Add-name dialog instead of stacking a second
      // prompt; Add commits, Cancel discards, both resolve via
      // `resolveAddNameNav`.
      if (this.addNameDialog?.visible) {
        return new Promise((resolve) => {
          this._addNameNavResolver = resolve;
        });
      }
      if (!this.hasPendingEdit()) {
        return Promise.resolve(true);
      }
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
      if (r) {
        r(true);
      }
    },
    onDiscardCancel() {
      this.discardDialog.visible = false;
      const r = this.discardDialog.resolver;
      this.discardDialog.resolver = null;
      if (r) {
        r(false);
      }
    },
    confirmField() {
      if (!this.photo || !this.canEdit) {
        this.editingField = null;
        return;
      }

      const field = this.editingField;
      const fieldDef = this.fieldRegistry[field];

      // Trim before the length gate so trailing whitespace doesn't trip the cap; Photo.update() trims again.
      if (fieldDef) {
        const currentValue = fieldDef.read(this.photo);
        if (typeof currentValue === "string") {
          const trimmed = currentValue.trim();
          if (trimmed !== currentValue) {
            fieldDef.write(this.photo, trimmed);
          }
          if (fieldDef.maxLength > 0 && trimmed.length > fieldDef.maxLength) {
            this.$notify.error(this.$gettext("%{s} is too long", { s: fieldDef.label }));
            return;
          }
        }
      }

      this.editingField = null;
      this.editOriginal = null;

      if (!this.photo.wasChanged()) {
        return;
      }

      // v-model already mutated the photo optimistically; on success
      // sync the matching Thumb fields, on failure roll back + notify.
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
    },
    // Blur handler for inline text fields: commits instead of reverting
    // so click-away / swipe / nav arrow don't lose typing. Escape still
    // cancels explicitly via cancelEditing().
    onInlineFieldBlur() {
      if (this._editStartedAt && Date.now() - this._editStartedAt < 200) {
        return;
      }
      if (!this.editingField) {
        return;
      }
      this.confirmField();
    },
    formatTime(model) {
      // Prefer Photo.getDateString() — source of truth for the Unknown /
      // year-only / month+year fallbacks. Gate on UID so limited-access
      // sessions (empty Photo placeholder) fall through to TakenAtLocal
      // instead of getDateString rendering "Unknown" against empty defaults.
      if (this.photo && this.photo.UID && typeof this.photo.getDateString === "function") {
        return this.photo.getDateString(true);
      }

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
      if (!route) {
        return;
      }
      const resolved = this.$router ? this.$router.resolve(route) : null;
      const href = resolved?.href || "";
      if (!href || typeof window === "undefined" || typeof window.open !== "function") {
        return;
      }
      window.open(href, "_blank", "noopener,noreferrer");
    },
    navigateToLabel(label) {
      if (!label) {
        return;
      }
      const slug = label.CustomSlug || label.Slug;
      if (!slug) {
        return;
      }
      this.openInNewTab({ name: "browse", query: { q: "label:" + slug } });
    },
    navigateToAlbum(album) {
      if (!album || !album.UID) {
        return;
      }
      this.openInNewTab({ name: "album", params: { album: album.UID, slug: "view" } });
    },
    navigateToPerson(marker) {
      if (!marker) {
        return;
      }
      if (marker.SubjUID) {
        this.openInNewTab({ name: "browse", query: { q: "subject:" + marker.SubjUID } });
      } else if (marker.Name) {
        this.openInNewTab({ name: "browse", query: { q: "person:" + marker.Name } });
      }
    },
    // loadChipOptions pulls typeahead suggestions from the shared cache and
    // sorts them locale-aware (backend order isn't reliably alphabetical).
    loadChipOptions(field) {
      const collator = (key) => (a, b) => (a[key] || "").localeCompare(b[key] || "", undefined, { sensitivity: "base", numeric: true });
      if (field === "labels") {
        typeaheadCache
          .getLabels()
          .then((models) => {
            this.chipState.labels.options = models.map((l) => ({ Name: l.Name, UID: l.UID })).sort(collator("Name"));
          })
          .catch(() => {});
      } else if (field === "albums") {
        typeaheadCache
          .getAlbums()
          .then((models) => {
            // Plain {Title, UID} objects: the rich Album model's
            // reactive metadata breaks v-combobox input handling.
            // Mirrors the labels mapping above.
            this.chipState.albums.options = models.map((a) => ({ Title: a.Title, UID: a.UID })).sort(collator("Title"));
          })
          .catch(() => {});
      } else if (field === "people") {
        typeaheadCache
          .getPeople()
          .then((models) => {
            // Plain {Name, UID} objects, mirroring the labels mapping.
            this.chipState.people.options = models
              .filter((p) => p && p.Name)
              .map((p) => ({ Name: p.Name, UID: p.UID }))
              .sort(collator("Name"));
          })
          .catch(() => {});
      }
    },
    // Clears typed text + selection for one combobox; the key bump
    // force-remounts the v-combobox to drop stale dropdown state.
    clearChipInput(field) {
      if (!field) {
        // Legacy callers without a field argument clear both —
        // cancelEditing() takes this path during inline-text rollback.
        Object.keys(this.chipState).forEach((f) => this.clearChipInput(f));
        return;
      }
      const state = this.chipState[field];
      if (!state) {
        return;
      }
      state.input = null;
      state.search = "";
      state.key++;
    },
    // Esc inside a chip combobox clears typed text and staged removals
    // for that field — chip parallel to the inline-text Esc revert.
    onChipEscape(field) {
      const state = this.chipState[field];
      if (!state) {
        return;
      }
      state.removals = [];
      this.clearChipInput(field);
    },
    // Generic chip-state helpers. Field is "labels" or "albums"; the key is
    // whatever uniquely identifies a chip in that field's domain (Label.ID
    // for labels, Album.UID for albums).
    isChipPendingRemoval(field, key) {
      const state = this.chipState[field];
      return Boolean(state && key != null && state.removals.includes(key));
    },
    togglePendingChipRemoval(field, key) {
      const state = this.chipState[field];
      if (!state || key == null) {
        return;
      }
      const idx = state.removals.indexOf(key);
      if (idx >= 0) {
        state.removals.splice(idx, 1);
      } else {
        state.removals.push(key);
      }
    },
    // Clears all pending removals for one chip section. Wired to the
    // section's Undo icon; restores soft-removed chips reactively.
    undoChipRemovals(field) {
      const state = this.chipState[field];
      if (!state) {
        return;
      }
      state.removals = [];
    },
    resetChipState() {
      Object.values(this.chipState).forEach((s) => {
        s.removals = [];
      });
    },
    // Click/Enter on a primary chip navigates to the related page in
    // both modes — chips act like links. Removal lives on the × icon
    // and Delete/Backspace via `onChipDelete`. Note the shape asymmetry:
    // labels are wrapped (`{ Label: { ID, ... } }`), albums aren't.
    onChipActivate(field, item) {
      if (!item) {
        return;
      }
      if (field === "labels") {
        return this.navigateToLabel(item.Label);
      } else if (field === "albums") {
        return this.navigateToAlbum(item);
      }
    },
    // Removal entry: wired to × icon and Delete/Backspace. No-op
    // outside edit mode (read-only chips stay as links); in edit mode
    // toggles pending removal until Undo or auto-commit.
    onChipDelete(field, item) {
      if (!item || !this.isEditable) {
        return;
      }
      const key = field === "labels" ? item?.Label?.ID : item.UID;
      this.togglePendingChipRemoval(field, key);
    },
    // Validates `rawName` and fires Photo.addLabel immediately; no
    // transient pending-add chip. Returns true when a save fired
    // (caller clears the input).
    addLabelImmediate(rawName) {
      if (!this.photo) {
        return false;
      }
      const name = (rawName || "").trim();
      if (!name) {
        return false;
      }
      if (name.length > LabelMaxLength.Name) {
        this.$notify.error(this.$gettext("Name too long"));
        return false;
      }
      const norm = this.$util.normalizeTitle(name);
      if (!norm) {
        return false;
      }
      // Re-typing a × clicked chip un-stages the removal locally so the
      // re-add doesn't race the deferred DELETE on auto-commit.
      const pending = this.labels.find((l) => this.isChipPendingRemoval("labels", l?.Label?.ID) && this.$util.normalizeTitle(l?.Label?.Name) === norm);
      if (pending?.Label?.ID != null) {
        this.togglePendingChipRemoval("labels", pending.Label.ID);
        return true;
      }
      // Already on the photo (and not pending removal)? Skip the API call.
      if (this.visibleLabels.some((l) => this.$util.normalizeTitle(l?.Label?.Name) === norm)) {
        return false;
      }
      // Match against the system-wide label list and send the canonical
      // name on a normalized hit, so the backend doesn't mint a
      // near-duplicate (e.g. `Hello Cat` reuses existing `hello-cat`).
      const existing = this.chipState.labels.options.find((l) => this.$util.normalizeTitle(l.Name) === norm);
      const finalName = existing ? existing.Name : name;
      this.photo.addLabel(finalName).catch(() => {
        this.$notify.error(this.$gettext("Failed to save changes"));
      });
      return true;
    },
    albumTitleConflicts(norm) {
      if (!norm) {
        return true;
      }
      return this.albums.some((a) => this.$util.normalizeTitle(a?.Title) === norm);
    },
    // Validates `album` and fires Photo.addToAlbum immediately; no
    // transient pending-add chip. onAlbumEnter wraps new albums in
    // Album.save() first so a UID exists.
    addAlbumImmediate(album) {
      if (!this.photo) {
        return false;
      }
      if (!album || typeof album !== "object" || !album.UID) {
        return false;
      }
      const title = (album.Title || "").trim();
      if (!title) {
        return false;
      }
      if (title.length > AlbumMaxLength.Title) {
        this.$notify.error(this.$gettext("Name too long"));
        return false;
      }
      // Mirrors the Labels pending-restore path (see addLabelImmediate).
      if (this.isChipPendingRemoval("albums", album.UID)) {
        this.togglePendingChipRemoval("albums", album.UID);
        return true;
      }
      if (this.albums.some((a) => a.UID === album.UID)) {
        return false;
      }
      const norm = this.$util.normalizeTitle(title);
      if (this.albumTitleConflicts(norm)) {
        return false;
      }
      this.photo.addToAlbum(album.UID).catch(() => {
        this.$notify.error(this.$gettext("Failed to save changes"));
      });
      return true;
    },
    onLabelSelected(value) {
      if (!value || typeof value !== "object" || !value.Name) {
        return;
      }
      this.addLabelImmediate(value.Name);
      this.clearChipInput("labels");
    },
    // Read the typed name from the per-field search ref. The ev.target
    // fallback guards against Vuetify clearing `search` on the same Enter
    // keystroke we handle, which would otherwise drop the pending addition.
    pendingChipName(field, ev) {
      const search = this.chipState[field]?.search;
      if (search) {
        return search;
      }
      const target = ev && ev.target ? ev.target : null;
      return target && typeof target.value === "string" ? target.value : "";
    },
    // Enter inside the Labels combobox: empty → no-op; too-long →
    // notify and leave the text; otherwise hand off to addLabelImmediate
    // and ALWAYS clear the input so the menu closes even when the label
    // was already on the photo and no API call fired.
    onLabelEnter(ev) {
      const search = this.pendingChipName("labels", ev).trim();
      if (!search) {
        return;
      }

      if (search.length > LabelMaxLength.Name) {
        this.$notify.error(this.$gettext("Name too long"));
        return;
      }

      this.addLabelImmediate(search);
      this.clearChipInput("labels");
    },
    // Confirms pending REMOVALS via Photo.removeLabel — additions take the
    // instant-save path (addLabelImmediate) and never reach this method.
    confirmLabels() {
      if (!this.photo) {
        return;
      }

      const state = this.chipState.labels;
      const removals = state.removals.slice();
      state.removals = [];

      // removeLabel's setValues(r.data) repopulates this.photo.Labels;
      // the photos.updated WS subscriber evicts the cache for the next
      // navigation. Album mutations emit albums.updated (not
      // photos.updated), which is why confirmAlbums does its own
      // evict+refind — see Photo.removeFromAlbum / addToAlbum.
      if (removals.length) {
        Promise.all(removals.map((id) => this.photo.removeLabel(id))).catch(() => {
          this.$notify.error(this.$gettext("Failed to save changes"));
        });
      }
    },
    // Confirms pending REMOVALS via Photo.removeFromAlbum — additions take
    // the instant-save path (addAlbumImmediate) and never reach this method.
    confirmAlbums() {
      if (!this.photo) {
        return;
      }

      const state = this.chipState.albums;
      const removals = state.removals.slice();
      state.removals = [];

      // Photo.removeFromAlbum owns the evict+refind per call, so the
      // sidebar's this.photo.Albums reflects the saved state without an
      // extra evictCache + find here.
      if (removals.length) {
        Promise.all(removals.map((uid) => this.photo.removeFromAlbum(uid))).catch(() => {
          this.$notify.error(this.$gettext("Failed to save changes"));
        });
      }
    },
    onAlbumSelected(value) {
      // v-combobox emits update:model-value transiently during free-text
      // typing. Bail silently on non-album values — clearing here would
      // remount the combobox and kill focus mid-keystroke. Free-text
      // Enter is committed via onAlbumEnter.
      if (!value || typeof value !== "object" || !value.UID) {
        return;
      }
      this.addAlbumImmediate(value);
      this.clearChipInput("albums");
    },
    onAlbumEnter(ev) {
      const search = this.pendingChipName("albums", ev).trim();
      if (!search) {
        return;
      }

      if (search.length > AlbumMaxLength.Title) {
        this.$notify.error(this.$gettext("Name too long"));
        return;
      }

      const norm = this.$util.normalizeTitle(search);
      if (!norm) {
        this.clearChipInput("albums");
        return;
      }

      const options = this.chipState.albums.options;

      // Normalized exact-match first: normalizeTitle is case- and
      // punctuation-insensitive (`Hello Cat`, `hello-cat`, `hello.CAT`
      // all collide). No substring fuzzy match — `test` must not silently
      // merge into `LRUTEST-ALBUM-…`; partial matches go through the
      // dropdown via onAlbumSelected.
      const exactMatch = options.find((a) => this.$util.normalizeTitle(a.Title) === norm);
      if (exactMatch) {
        this.onAlbumSelected(exactMatch);
        return;
      }

      // Skip the API round-trip if a normalized title clash already exists
      // among the photo's current albums.
      if (this.albumTitleConflicts(norm)) {
        this.clearChipInput("albums");
        return;
      }

      const album = new Album({ Title: search });

      album
        .save()
        .then(() => {
          if (album.UID && this.addAlbumImmediate(album)) {
            options.push(album);
          }
        })
        .catch(() => {})
        .finally(() => {
          this.clearChipInput("albums");
        });
    },
    confirmDateTime(data) {
      this.dateTimeDialog = false;

      const photo = this.view?.photo;
      if (!photo || !photo.UID || !this.canEdit) {
        return;
      }

      photo.Day = data.Day;
      photo.Month = data.Month;
      photo.Year = data.Year;
      photo.TimeZone = data.TimeZone;

      const localDate = photo.localDate(data.time);
      if (!localDate.isValid) {
        return;
      }

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
          if (!this.view?.model) {
            return;
          }
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
      if (!photo || !photo.UID || !this.canEdit) {
        return;
      }

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
      if (!photo || !photo.UID || !this.canEdit) {
        return;
      }

      photo.Lat = data.lat;
      photo.Lng = data.lng;
      photo.PlaceSrc = "manual";

      if (data.location?.country) {
        photo.Country = data.location.country;
      }

      photo
        .update()
        .then(() => {
          if (!this.view?.model) {
            return;
          }
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
