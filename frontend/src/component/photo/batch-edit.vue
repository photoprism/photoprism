<template>
  <v-dialog
    ref="dialog"
    tabindex="-1"
    :model-value="visible"
    :fullscreen="$vuetify.display.mdAndDown"
    scrim
    scrollable
    persistent
    class="p-dialog p-photo-batch-edit v-dialog--sidepanel v-dialog--batch-edit"
    @click.stop="onClick"
    @keydown.esc.exact="onClose"
    @after-enter="afterEnter"
    @after-leave="afterLeave"
  >
    <v-card ref="content" class="edit-batch__card" :tile="$vuetify.display.mdAndDown" tabindex="-1">
      <v-toolbar flat color="navigation" :density="$vuetify.display.mdAndDown ? 'compact' : 'comfortable'">
        <v-btn icon class="action-close" @click.stop="onClose">
          <v-icon>mdi-close</v-icon>
        </v-btn>

        <v-toolbar-title>
          {{ formTitle }}
        </v-toolbar-title>
      </v-toolbar>

      <v-progress-linear v-if="saving" :indeterminate="true" color="surface-variant"></v-progress-linear>

      <v-row v-if="formData" dense :class="!$vuetify.display.mdAndDown ? 'overflow-hidden' : ''">
        <!-- Desktop view -->
        <v-col v-if="!$vuetify.display.mdAndDown" cols="12" lg="4" class="scroll-col">
          <div v-if="model.models" class="edit-batch photo-results list-view">
            <v-data-table
              :headers="tableHeaders"
              :items="model.models"
              :show-select="false"
              hide-default-footer
              item-key="ID"
              density="comfortable"
              class="bg-transparent v-table--batch-edit"
            >
              <template #header.select>
                <v-checkbox
                  :model-value="isAllSelected"
                  hide-details
                  density="compact"
                  @update:model-value="toggleAll"
                />
              </template>

              <template #item.select="{ item }">
                <v-checkbox
                  :model-value="isSelected(item)"
                  hide-details
                  density="compact"
                  @update:model-value="toggle(item)"
                />
              </template>

              <template #item.preview="{ item, index }">
                <div class="media result col-preview">
                  <div
                    :style="`background-image: url(${item.thumbnailUrl('tile_50')})`"
                    class="preview"
                    @touchstart.passive="onMouseDown($event, index)"
                    @touchend.stop="onSelectClick($event, index, false)"
                    @mousedown="onMouseDown($event, index)"
                    @click.stop.prevent="openPhoto(index)"
                  >
                    <div class="preview__overlay"></div>
                    <button
                      v-if="item.Type === 'video' || item.Type === 'live' || item.Type === 'animated'"
                      class="input-open"
                      @click.stop.prevent="openPhoto(index)"
                    >
                      <i v-if="item.Type === 'live'" class="action-live" :title="$gettext('Live')">
                        <icon-live-photo />
                      </i>
                      <i
                        v-else-if="item.Type === 'animated'"
                        class="mdi mdi-file-gif-box"
                        :title="$gettext('Animated')"
                      />
                      <i v-else-if="item.Type === 'video'" class="mdi mdi-play" :title="$gettext('Video')" />
                    </button>
                  </div>
                </div>
              </template>

              <template #item.name="{ item, index }">
                <span
                  class="meta-data meta-title col-auto break-word text-start clickable"
                  :title="item.FileName"
                  @click.exact="openPhoto(index)"
                >
                  {{ item.getOriginalName() }}
                </span>
              </template>
            </v-data-table>
          </div>
        </v-col>

        <!-- Mobile view -->
        <v-col v-else cols="12" class="px-0">
          <div v-if="model.models" class="edit-batch photo-results list-view">
            <v-expansion-panels
              v-model="expanded"
              variant="accordion"
              density="compact"
              class="elevation-0 ra-4 bg-transparent"
            >
              <v-expansion-panel class="pa-0 ra-4 elevation-0">
                <v-expansion-panel-title class="px-4">{{ $gettext("Selection") }}</v-expansion-panel-title>
                <v-expansion-panel-text class="px-2">
                  <v-data-table
                    :headers="mobileTableHeaders"
                    :items="model.models"
                    :show-select="false"
                    hide-default-footer
                    item-key="ID"
                    density="compact"
                    class="elevation-0 ra-4 v-table--batch-edit"
                  >
                    <template #header.select>
                      <v-checkbox
                        :model-value="isAllSelected"
                        hide-details
                        density="compact"
                        @update:model-value="toggleAll"
                      />
                    </template>

                    <template #item.select="{ item }">
                      <v-checkbox
                        :model-value="isSelected(item)"
                        hide-details
                        density="compact"
                        @update:model-value="toggle(item)"
                      />
                    </template>

                    <template #item.preview="{ item, index }">
                      <div class="media result col-preview">
                        <div
                          :style="`background-image: url(${item.thumbnailUrl('tile_50')})`"
                          class="preview"
                          @touchstart.passive="onMouseDown($event, index)"
                          @touchend.stop="onSelectClick($event, index, false)"
                          @mousedown="onMouseDown($event, index)"
                          @click.stop.prevent="openPhoto(index)"
                        >
                          <div class="preview__overlay"></div>
                          <button
                            v-if="item.Type === 'video' || item.Type === 'live' || item.Type === 'animated'"
                            class="input-open"
                            @click.stop.prevent="openPhoto(index)"
                          >
                            <i v-if="item.Type === 'live'" class="action-live" :title="$gettext('Live')">
                              <icon-live-photo />
                            </i>
                            <i
                              v-else-if="item.Type === 'animated'"
                              class="mdi mdi-file-gif-box"
                              :title="$gettext('Animated')"
                            />
                            <i v-else-if="item.Type === 'video'" class="mdi mdi-play" :title="$gettext('Video')" />
                          </button>
                        </div>
                      </div>
                    </template>

                    <template #item.name="{ item, index }">
                      <span
                        class="meta-data meta-title col-auto text-start clickable edit-batch__file-name"
                        :title="item.FileName"
                        @click.exact="openPhoto(index)"
                      >
                        {{ item.getOriginalName() }}
                      </span>
                    </template>
                  </v-data-table>
                </v-expansion-panel-text>
              </v-expansion-panel>
            </v-expansion-panels>
          </div>
        </v-col>

        <v-col v-if="model.values" cols="12" lg="8" class="scroll-col">
          <v-form
            ref="form"
            validate-on="invalid-input"
            class="p-form p-form-photo-details-meta pa-0"
            accept-charset="UTF-8"
            @submit.prevent="save"
          >
            <div class="form-body">
              <div class="form-controls">
                <v-row dense>
                  <v-col cols="12" class="text-subtitle-2">{{ $gettext(`Description`) }}</v-col>
                  <v-col cols="12">
                    <v-text-field
                      hide-details
                      :label="$pgettext(`Photo`, `Title`)"
                      :model-value="formData.Title.value"
                      :placeholder="getFieldData('text-field', 'Title').placeholder"
                      :persistent-placeholder="getFieldData('text-field', 'Title').persistent"
                      :append-inner-icon="getIcon('text-field', 'Title')"
                      autocomplete="off"
                      density="comfortable"
                      class="input-title"
                      @click:append-inner="toggleField('Title', $event)"
                      @update:model-value="(val) => changeValue(val, 'text-field', 'Title')"
                    ></v-text-field>
                  </v-col>
                  <v-col cols="12">
                    <v-textarea
                      hide-details
                      autocomplete="off"
                      auto-grow
                      :label="$gettext('Caption')"
                      :model-value="formData.Caption.value"
                      :placeholder="getFieldData('text-field', 'Caption').placeholder"
                      :persistent-placeholder="getFieldData('text-field', 'Caption').persistent"
                      :append-inner-icon="getIcon('text-field', 'Caption')"
                      :rows="1"
                      density="comfortable"
                      class="input-caption"
                      @click:append-inner="toggleField('Caption', $event)"
                      @update:model-value="(val) => changeValue(val, 'text-field', 'Caption')"
                    ></v-textarea>
                  </v-col>
                </v-row>
                <v-row dense>
                  <v-col cols="12" class="text-subtitle-2">{{ $gettext(`Date & Time`) }}</v-col>
                  <v-col cols="6" md="3">
                    <v-autocomplete
                      v-model="formData.Day.value"
                      :label="$gettext('Day')"
                      autocomplete="off"
                      hide-details
                      hide-no-data
                      :items="getFieldData('select-field', 'Day').items"
                      :placeholder="getFieldData('select-field', 'Day').placeholder"
                      :persistent-placeholder="getFieldData('select-field', 'Day').persistent"
                      item-title="text"
                      item-value="value"
                      density="comfortable"
                      validate-on="input"
                      class="input-day"
                      @update:model-value="(val) => changeSelectValue(val, 'select-field', 'Day')"
                    >
                    </v-autocomplete>
                  </v-col>
                  <v-col cols="6" md="3">
                    <v-autocomplete
                      v-model="formData.Month.value"
                      :label="$gettext('Month')"
                      autocomplete="off"
                      hide-details
                      hide-no-data
                      :items="getFieldData('select-field', 'Month').items"
                      :placeholder="getFieldData('select-field', 'Month').placeholder"
                      :persistent-placeholder="getFieldData('select-field', 'Month').persistent"
                      item-title="text"
                      item-value="value"
                      density="comfortable"
                      validate-on="input"
                      class="input-month"
                      @update:model-value="(val) => changeSelectValue(val, 'select-field', 'Month')"
                    >
                    </v-autocomplete>
                  </v-col>
                  <v-col cols="12" sm="6" md="3">
                    <v-autocomplete
                      v-model="formData.Year.value"
                      :label="$gettext('Year')"
                      autocomplete="off"
                      hide-details
                      hide-no-data
                      :items="getFieldData('select-field', 'Year').items"
                      :placeholder="getFieldData('select-field', 'Year').placeholder"
                      :persistent-placeholder="getFieldData('select-field', 'Year').persistent"
                      item-title="text"
                      item-value="value"
                      density="comfortable"
                      validate-on="input"
                      class="input-year"
                      @update:model-value="(val) => changeSelectValue(val, 'select-field', 'Year')"
                    >
                    </v-autocomplete>
                  </v-col>
                  <v-col cols="12" sm="6" md="3">
                    <v-autocomplete
                      v-model="formData.TimeZone.value"
                      :label="$gettext('Time Zone')"
                      hide-no-data
                      :items="getFieldData('select-field', 'TimeZone').items"
                      :placeholder="getFieldData('select-field', 'TimeZone').placeholder"
                      :persistent-placeholder="getFieldData('select-field', 'TimeZone').persistent"
                      item-value="ID"
                      item-title="Name"
                      density="comfortable"
                      class="input-timezone"
                      @update:model-value="(val) => changeSelectValue(val, 'select-field', 'TimeZone')"
                    ></v-autocomplete>
                  </v-col>
                  <v-col cols="12" class="text-subtitle-2">{{ $gettext(`Location`) }}</v-col>
                  <v-col cols="12" md="6">
                    <p-location-input
                      :latlng="currentCoordinates"
                      :placeholder="locationPlaceholder"
                      :persistent-placeholder="true"
                      hide-details
                      :label="locationLabel"
                      density="comfortable"
                      validate-on="input"
                      :show-map-button="!placesDisabled"
                      :map-button-title="$gettext('Adjust Location')"
                      :map-button-disabled="placesDisabled"
                      :is-mixed="isLocationMixed"
                      :is-deleted="isLocationDeleted"
                      class="input-coordinates"
                      @update:latlng="updateLatLng"
                      @changed="onLocationChanged"
                      @open-map="adjustLocation"
                      @delete="onLocationDelete"
                      @undo="onLocationUndo"
                    ></p-location-input>
                  </v-col>
                  <v-col cols="12" sm="6" md="3">
                    <v-autocomplete
                      v-model="formData.Country.value"
                      :label="$gettext('Country')"
                      hide-details
                      hide-no-data
                      autocomplete="off"
                      item-value="Code"
                      item-title="Name"
                      :items="getFieldData('select-field', 'Country').items"
                      :placeholder="getFieldData('select-field', 'Country').placeholder"
                      :persistent-placeholder="getFieldData('select-field', 'Country').persistent"
                      :readonly="isCountryReadOnly"
                      density="comfortable"
                      validate-on="input"
                      class="input-country"
                      @update:model-value="(val) => changeSelectValue(val, 'select-field', 'Country')"
                    >
                    </v-autocomplete>
                  </v-col>
                  <v-col cols="12" sm="6" md="3">
                    <v-text-field
                      hide-details
                      flat
                      autocomplete="off"
                      autocorrect="off"
                      autocapitalize="none"
                      :label="$gettext('Altitude (m)')"
                      :model-value="formData.Altitude.value"
                      :placeholder="getFieldData('input-field', 'Altitude').placeholder"
                      :persistent-placeholder="getFieldData('input-field', 'Altitude').persistent"
                      :append-inner-icon="getIcon('input-field', 'Altitude')"
                      color="surface-variant"
                      density="comfortable"
                      validate-on="input"
                      class="input-altitude"
                      @click:append-inner="toggleField('Altitude', $event)"
                      @update:model-value="(val) => changeValue(val, 'input-field', 'Altitude')"
                    ></v-text-field>
                  </v-col>
                </v-row>
                <v-row dense>
                  <v-col cols="12" class="text-subtitle-2">{{ $pgettext(`Edit`, `Content`) }}</v-col>
                  <v-col cols="12" sm="8">
                    <v-textarea
                      hide-details
                      autocomplete="off"
                      auto-grow
                      :label="$gettext('Subject')"
                      :model-value="formData.DetailsSubject.value"
                      :placeholder="getFieldData('text-field', 'DetailsSubject').placeholder"
                      :persistent-placeholder="getFieldData('text-field', 'DetailsSubject').persistent"
                      :append-inner-icon="getIcon('text-field', 'DetailsSubject')"
                      :rows="1"
                      density="comfortable"
                      class="input-subject"
                      @click:append-inner="toggleField('DetailsSubject', $event)"
                      @update:model-value="(val) => changeValue(val, 'text-field', 'DetailsSubject')"
                    ></v-textarea>
                  </v-col>
                  <v-col cols="12" sm="4">
                    <v-autocomplete
                      v-model="formData.Type.value"
                      :label="$gettext('Type')"
                      autocomplete="off"
                      hide-details
                      hide-no-data
                      :items="getFieldData('select-field', 'Type').items"
                      :placeholder="getFieldData('select-field', 'Type').placeholder"
                      :persistent-placeholder="getFieldData('select-field', 'Type').persistent"
                      item-title="text"
                      item-value="value"
                      density="comfortable"
                      validate-on="input"
                      class="input-type"
                      @update:model-value="(val) => changeSelectValue(val, 'select-field', 'Type')"
                    >
                    </v-autocomplete>
                  </v-col>
                  <v-col cols="12" sm="6">
                    <v-text-field
                      hide-details
                      autocomplete="off"
                      :label="$gettext('Copyright')"
                      :model-value="formData.DetailsCopyright.value"
                      :placeholder="getFieldData('text-field', 'DetailsCopyright').placeholder"
                      :persistent-placeholder="getFieldData('text-field', 'DetailsCopyright').persistent"
                      :append-inner-icon="getIcon('text-field', 'DetailsCopyright')"
                      density="comfortable"
                      class="input-copyright"
                      @click:append-inner="toggleField('DetailsCopyright', $event)"
                      @update:model-value="(val) => changeValue(val, 'text-field', 'DetailsCopyright')"
                    ></v-text-field>
                  </v-col>
                  <v-col cols="12" sm="6">
                    <v-text-field
                      hide-details
                      autocomplete="off"
                      :label="$gettext('Artist')"
                      :model-value="formData.DetailsArtist.value"
                      :placeholder="getFieldData('text-field', 'DetailsArtist').placeholder"
                      :persistent-placeholder="getFieldData('text-field', 'DetailsArtist').persistent"
                      :append-inner-icon="getIcon('text-field', 'DetailsArtist')"
                      density="comfortable"
                      class="input-artist"
                      @click:append-inner="toggleField('DetailsArtist', $event)"
                      @update:model-value="(val) => changeValue(val, 'text-field', 'DetailsArtist')"
                    ></v-text-field>
                  </v-col>
                  <v-col cols="12">
                    <v-textarea
                      hide-details
                      autocomplete="off"
                      auto-grow
                      :label="$gettext('License')"
                      :model-value="formData.DetailsLicense.value"
                      :placeholder="getFieldData('text-field', 'DetailsLicense').placeholder"
                      :persistent-placeholder="getFieldData('text-field', 'DetailsLicense').persistent"
                      :append-inner-icon="getIcon('text-field', 'DetailsLicense')"
                      :rows="1"
                      density="comfortable"
                      class="input-license"
                      @click:append-inner="toggleField('DetailsLicense', $event)"
                      @update:model-value="(val) => changeValue(val, 'text-field', 'DetailsLicense')"
                    ></v-textarea>
                  </v-col>
                </v-row>
                <v-row dense>
                  <v-col cols="12" class="text-subtitle-2">{{ $gettext(`Labels`) }}</v-col>
                  <v-col cols="12">
                    <p-input-chip-selector
                      v-model:items="labelItems"
                      :available-items="availableLabelOptions"
                      :resolve-item-from-text="resolveLabelFromText"
                      :normalize-title-for-compare="normalizeLabelTitleForCompare"
                      :input-placeholder="$gettext('Select or create labels')"
                      :empty-text="$gettext('No labels assigned')"
                      :loading="loading"
                      :disabled="false"
                      @update:items="onLabelsUpdate"
                    />
                  </v-col>
                </v-row>
                <v-row dense>
                  <v-col cols="12" class="text-subtitle-2">{{ $gettext(`Albums`) }}</v-col>
                  <v-col cols="12">
                    <p-input-chip-selector
                      v-model:items="albumItems"
                      :available-items="availableAlbumOptions"
                      :input-placeholder="$gettext('Select or create albums')"
                      :empty-text="$gettext('No albums assigned')"
                      :loading="loading"
                      :disabled="false"
                      @update:items="onAlbumsUpdate"
                    />
                  </v-col>
                </v-row>
                <v-row dense>
                  <v-col v-for="fieldName in toggleFieldsArray" :key="fieldName" cols="12" sm="12" md="6" lg="6" xl="3">
                    <div class="d-flex flex-column">
                      <label class="form-label mb-3 text-subtitle-2">{{ getFieldDisplayName(fieldName) }}</label>
                      <v-btn-toggle v-model="formData[fieldName].value" mandatory color="primary" density="compact">
                        <v-btn
                          v-for="option in toggleOptions(fieldName)"
                          :key="option.value"
                          :value="option.value"
                          size="small"
                          density="compact"
                          @click="changeToggleValue(option.value, fieldName)"
                        >
                          {{ option.text }}
                        </v-btn>
                      </v-btn-toggle>
                    </div>
                  </v-col>
                </v-row>
              </div>
            </div>

            <div class="form-actions form-actions--sticky">
              <div class="action-buttons">
                <v-btn color="button" variant="flat" class="action-close" @click.stop="close">
                  <span v-if="hasUnsavedChanges()">{{ $gettext(`Discard`) }}</span>
                  <span v-else>{{ $gettext(`Close`) }}</span>
                </v-btn>
                <v-btn
                  color="highlight"
                  variant="flat"
                  class="action-apply action-approve"
                  :loading="saving"
                  :disabled="saving || !hasUnsavedChanges()"
                  @click.stop="save(false)"
                >
                  <span>{{ $gettext(`Save`) }}</span>
                </v-btn>
              </div>
            </div>
          </v-form>
        </v-col>
      </v-row>
    </v-card>
    <p-location-dialog
      :visible="locationDialog"
      :latlng="currentCoordinates"
      @close="locationDialog = false"
      @confirm="confirmLocation"
    ></p-location-dialog>
  </v-dialog>
</template>
<script>
import * as options from "options/options";
import IconLivePhoto from "../icon/live-photo.vue";
import { Batch } from "model/batch";
import Album from "model/album";
import Label from "model/label";
import Thumb from "model/thumb";
import PLocationDialog from "component/location/dialog.vue";
import PLocationInput from "component/location/input.vue";
import PInputChipSelector from "component/input/chip-selector.vue";
import $util from "common/util";

// TODO: Handle cases where users have more than 10000 albums and/or labels.
const MaxResults = 10000;
const iconClear = "mdi-close-circle";
const iconUndo = "mdi-undo";

export default {
  name: "PPhotoBatchEdit",
  components: {
    IconLivePhoto,
    PLocationDialog,
    PLocationInput,
    PInputChipSelector,
  },
  props: {
    visible: {
      type: Boolean,
      default: false,
    },
    selection: {
      type: Array,
      default: () => [],
    },
    openDate: {
      type: Function,
      default: () => {},
    },
    openLocation: {
      type: Function,
      default: () => {},
    },
    editPhoto: {
      type: Function,
      default: () => {},
    },
  },
  emits: ["close"],
  data() {
    return {
      model: new Batch(),
      uid: "",
      loading: false,
      saving: false,
      subscriptions: [],
      selectionsFullInfo: [],
      selectedPhotosLength: 0,
      expanded: [0],
      isAllSelected: true,
      allSelectedLength: 0,
      options,
      firstVisibleElementIndex: 0,
      lastVisibleElementIndex: 0,
      mouseDown: {
        index: -1,
        scrollY: window.scrollY,
        timeStamp: -1,
      },
      values: {},
      formData: null,
      previousFormData: {},
      deletedFields: {},
      toggleFieldsArray: ["Scan", "Favorite", "Private", "Panorama"],
      actions: { none: "none", update: "update", add: "add", remove: "remove" },
      locationDialog: false,
      placesDisabled: !this.$config.feature("places"),
      locationLabel: this.$gettext(`Location`),
      availableAlbumOptions: [],
      availableLabelOptions: [],
      albumItems: [],
      labelItems: [],
      tableHeaders: [
        {
          key: "select",
          title: "",
          sortable: false,
          headerProps: { class: "col-select" },
        },
        {
          key: "preview",
          title: this.$gettext(`Preview`),
          sortable: false,
          headerProps: { class: "col-preview" },
        },
        {
          key: "name",
          title: this.$gettext(`File Name`),
          sortable: false,
          headerProps: { class: "col-name break-word" },
        },
      ],
      mobileTableHeaders: [
        {
          key: "select",
          title: "",
          sortable: false,
          headerProps: { class: "col-select" },
        },
        {
          key: "preview",
          title: this.$gettext(`Preview`),
          sortable: false,
          headerProps: { class: "col-preview" },
        },
        {
          key: "name",
          title: this.$gettext(`File Name`),
          sortable: false,
          headerProps: { class: "col-name break-word" },
        },
      ],
    };
  },
  computed: {
    formTitle() {
      return this.$gettext(`Edit Photos (%{n})`, { n: this.allSelectedLength });
    },
    currentCoordinates() {
      if (this.isLocationMixed || this.deletedFields.Lat || this.deletedFields.Lng) {
        return [0, 0];
      }
      const latData = this.values?.Lat;
      const lngData = this.values?.Lng;

      // If no form data yet, return default
      if (!this.formData || !latData || !lngData) {
        return [0, 0];
      }

      const lat = this.formData.Lat.value;
      const lng = this.formData.Lng.value;

      // If form data has been updated, use the form data values
      if (this.formData.Lat.action === this.actions.update || this.formData.Lng.action === this.actions.update) {
        return [parseFloat(lat) || 0, parseFloat(lng) || 0];
      }

      // Handle mixed values or empty values
      if (
        latData.mixed ||
        lngData.mixed ||
        lat === options.Mixed.String ||
        lng === options.Mixed.String ||
        lat === "" ||
        lng === "" ||
        lat === null ||
        lng === null
      ) {
        return [0, 0];
      }

      return [parseFloat(lat) || 0, parseFloat(lng) || 0];
    },
    locationPlaceholder() {
      if (this.deletedFields.Lat || this.deletedFields.Lng) {
        return options.Deleted.Placeholder();
      } else if (this.isLocationMixed) {
        return options.Mixed.Placeholder();
      }

      const lat = this.formData?.Lat?.value;
      const lng = this.formData?.Lng?.value;

      if ((lat === null || lat === 0) && (lng === null || lng === 0)) {
        return "37.75267, -122.543";
      }

      return "";
    },
    isLocationDeleted() {
      // Check if explicitly deleted
      if (this.deletedFields.Lat || this.deletedFields.Lng) {
        return true;
      }

      // Check if coordinates have been changed (should show undo)
      if (this.formData && this.previousFormData) {
        const currentLat = this.formData.Lat?.value || 0;
        const currentLng = this.formData.Lng?.value || 0;
        const previousLat = this.previousFormData.Lat?.value || 0;
        const previousLng = this.previousFormData.Lng?.value || 0;

        if (currentLat !== previousLat || currentLng !== previousLng) {
          return true; // This will show the undo button
        }
      }

      return false;
    },
    isLocationMixed() {
      if (this.isLocationDeleted) {
        return false;
      }
      const latData = this.values?.Lat;
      const lngData = this.values?.Lng;

      // If form data has been updated, not mixed anymore
      if (this.formData?.Lat?.action === this.actions.update || this.formData?.Lng?.action === this.actions.update) {
        return false;
      }

      return latData?.mixed || lngData?.mixed;
    },
    isCountryReadOnly() {
      if (!this.formData || !this.formData.Lat || !this.formData.Lng) {
        return false;
      }
      return !!(this.formData.Lat.value || this.formData.Lng.value);
    },
  },
  watch: {
    selection: {
      deep: true,
      handler(newSelection) {
        if (!this.visible || !Array.isArray(newSelection) || !this.model || !Array.isArray(this.model.selection)) {
          return;
        }

        const selectedIds = new Set(newSelection);

        this.model.selection.forEach((entry) => {
          const shouldBeSelected = selectedIds.has(entry.id);
          if (entry.selected !== shouldBeSelected) {
            entry.selected = shouldBeSelected;
          }
        });

        this.updateToggleAll();
        this.allSelectedLength = this.model.getLengthOfAllSelected();
      },
    },
  },
  created() {
    // this.subscriptions.push(this.$event.subscribe("photos.updated", (ev, data) => this.onUpdate(ev, data)));
    // this.fetchAvailableOptions();
  },
  beforeUnmount() {
    for (let i = 0; i < this.subscriptions.length; i++) {
      this.$event.unsubscribe(this.subscriptions[i]);
    }
  },
  methods: {
    afterEnter() {
      this.$view.enter(this);

      // Load data when dialog opens.
      this.placesDisabled = !this.$config.feature("places");
      this.expanded = [];
      return this.fetchAvailableOptions()
        .then(() => {
          return this.model.load(this.selection);
        })
        .then(() => {
          this.values = this.model.values;
          this.setFormData();
          this.allSelectedLength = this.model.getLengthOfAllSelected();
        })
        .catch(() => {
          this.values = {};
          this.setFormData();
          this.allSelectedLength = 0;
        });
    },
    afterLeave() {
      // Reset when dialog is closed.
      this.model = new Batch();
      this.values = {};
      this.deletedFields = {};
      this.allSelectedLength = 0;
      this.$view.leave(this);
    },
    getModelCount() {
      return this.model?.models?.length ? this.model.models.length : 0;
    },
    normalizeLabelTitleForCompare(s) {
      return $util.normalizeLabelTitle(s);
    },
    resolveLabelFromText(inputTitle) {
      if (!inputTitle || !Array.isArray(this.availableLabelOptions)) {
        return null;
      }

      const t = String(inputTitle).trim();
      if (!t) return null;

      const nt = $util.normalizeLabelTitle(t);
      const st = $util.slugifyLabelTitle(t);

      let found = this.availableLabelOptions.find((o) => o.title.toLowerCase() === t.toLowerCase());
      if (found) return { value: found.value, title: found.title };

      found = this.availableLabelOptions.find((o) => o.slug === st || o.customSlug === st);
      if (found) return { value: found.value, title: found.title };

      found = this.availableLabelOptions.find((o) => $util.normalizeLabelTitle(o.title) === nt);
      if (found) return { value: found.value, title: found.title };

      return { value: "", title: t };
    },
    changeValue(newValue, fieldType, fieldName) {
      if (!fieldName) return;

      const previousValue = this.previousFormData[fieldName].value;
      this.formData[fieldName].action = this.actions.update;

      // Convert numeric fields to proper types
      let processedValue = newValue;
      if (fieldType === "input-field") {
        if (fieldName === "Lat" || fieldName === "Lng") {
          processedValue = parseFloat(newValue) || 0;
        } else if (
          ["Altitude", "Day", "Month", "Year", "Iso", "FocalLength", "CameraID", "LensID"].includes(fieldName)
        ) {
          processedValue = parseInt(newValue) || 0;
        } else if (fieldName === "FNumber") {
          processedValue = parseFloat(newValue) || 0;
        }
      }

      this.formData[fieldName].value = processedValue;

      if (processedValue === previousValue) {
        this.formData[fieldName].action = this.actions.none;
      }

      this.getIcon(fieldType, fieldName);
    },
    changeSelectValue(newValue, fieldType, fieldName) {
      if (!fieldName) return;

      const previousValue = this.previousFormData[fieldName].value;
      this.formData[fieldName].action = this.actions.update;

      if (fieldName === "Day" || fieldName === "Month" || fieldName === "Year") {
        // If the incoming value is an object (i.e., a selection was made from the list), use the .value property.
        // If the user manually entered a number, it will be a string or number, so parseInt it.
        let processedValue;
        if (
          typeof newValue === "object" &&
          newValue !== null &&
          Object.prototype.hasOwnProperty.call(newValue, "value")
        ) {
          processedValue = newValue.value;
        } else {
          processedValue = parseInt(newValue, 10) || 0;
        }

        this.formData[fieldName].value = processedValue;

        if (processedValue === previousValue) {
          this.formData[fieldName].action = this.actions.none;
        }
        // Auto-clamp only when date is fully known (no unknown/mixed) per spec.
        this.clampBatchDayIfResolvable();
      } else if (fieldName === "Type") {
        // Special logic for the Type field
        const processedValue = newValue.value !== undefined ? newValue.value : newValue;
        this.formData[fieldName].value = processedValue;
        if (processedValue === previousValue) {
          this.formData[fieldName].action = this.actions.none;
        }
      } else {
        // General logic for all other select fields (Country, TimeZone, etc.)
        this.formData[fieldName].value = newValue;

        const newVal = newValue !== options.Mixed.ID ? newValue : options.Mixed.String;
        if (newVal === previousValue) {
          this.formData[fieldName].action = this.actions.none;
        }
      }
    },
    // Helpers for batch clamping when Month/Year are fully known and not mixed
    batchYear() {
      const y = parseInt(this.formData?.Year?.value, 10);
      return isNaN(y) ? 0 : y;
    },
    batchMonth() {
      const m = parseInt(this.formData?.Month?.value, 10);
      return isNaN(m) ? 0 : m;
    },
    isBatchDateResolvable() {
      // Only if both Month and Year are > 0 and their initial state was not mixed
      const y = this.batchYear();
      const m = this.batchMonth();
      const yearMixed = !!this.formData?.Year?.mixed;
      const monthMixed = !!this.formData?.Month?.mixed;
      return y > 0 && m > 0 && !yearMixed && !monthMixed;
    },
    clampBatchDayIfResolvable() {
      if (!this.isBatchDateResolvable()) return;
      const y = this.batchYear();
      const m = this.batchMonth();
      const d = parseInt(this.formData?.Day?.value, 10);
      if (isNaN(d) || d <= 0) return; // Unknown or empty: do nothing in batch UI
      const maxDay = new Date(Date.UTC(y, m, 0)).getUTCDate();
      if (d > maxDay) {
        this.formData.Day.value = maxDay;
        this.formData.Day.action = this.actions.update;
      }
    },
    changeToggleValue(newValue, fieldName) {
      if (!fieldName) return;

      const previousValue = this.previousFormData[fieldName].value;
      this.formData[fieldName].action = this.actions.update;
      this.formData[fieldName].value = newValue;

      if (newValue === previousValue) {
        this.formData[fieldName].action = this.actions.none;
      }
    },
    setFormData() {
      this.deletedFields = {};

      this.formData = this.model.getDefaultFormData();

      const fieldConfigs = [
        { type: "text-field", name: "Title" },
        { type: "text-field", name: "DetailsSubject" },
        { type: "text-field", name: "Caption" },
        { type: "select-field", name: "Day" },
        { type: "select-field", name: "Month" },
        { type: "select-field", name: "Year" },
        { type: "select-field", name: "TimeZone" },
        { type: "select-field", name: "Country" },
        { type: "input-field", name: "Altitude" },
        { type: "input-field", name: "Lat" },
        { type: "input-field", name: "Lng" },
        { type: "text-field", name: "DetailsArtist" },
        { type: "text-field", name: "DetailsCopyright" },
        { type: "text-field", name: "DetailsLicense" },
        // DetailsKeywords and the following EXIF fields are wired through the form so
        // their mixed state survives round-trips. UI inputs (and server-side saves) will
        // be added in a follow-up, so leave them here until that work lands.
        { type: "text-field", name: "DetailsKeywords" },
        { type: "select-field", name: "Type" },
        { type: "input-field", name: "Iso" },
        { type: "input-field", name: "FocalLength" },
        { type: "input-field", name: "FNumber" },
        { type: "text-field", name: "Exposure" },
        { type: "input-field", name: "CameraID" },
        { type: "input-field", name: "LensID" },
      ];

      fieldConfigs.forEach(({ type, name, key }) => {
        const formKey = key || name;
        const fieldData = this.values[formKey];

        if (!fieldData) return;

        const { value, placeholder } = this.getFieldData(type, name);
        this.formData[formKey] = {
          action: this.actions.none,
          mixed: fieldData.mixed || false,
          value: value !== undefined ? value : "",
        };

        if (type === "text-field" || type === "input-field") {
          this.previousFormData[formKey] = {
            value,
            placeholder,
            action: fieldData.action || this.actions.none,
            mixed: fieldData.mixed || false,
          };
        } else {
          this.previousFormData[formKey] = {
            value,
            action: fieldData.action || this.actions.none,
            mixed: fieldData.mixed || false,
          };
        }
      });

      // Set values for toggle fields (boolean fields)
      this.toggleFieldsArray.forEach((fieldName) => {
        const fieldData = this.values[fieldName];
        if (!fieldData) return;

        const toggleValue = this.getToggleValue(fieldName);

        this.formData[fieldName] = {
          action: this.actions.none,
          mixed: fieldData.mixed || false,
          value: toggleValue,
        };
        this.previousFormData[fieldName] = {
          action: fieldData.action || this.actions.none,
          mixed: fieldData.mixed || false,
          value: toggleValue,
        };
      });

      // Initialize Albums and Labels from backend data
      const albumsData = this.values.Albums || {
        items: [],
        mixed: false,
        action: this.actions.none,
      };
      const labelsData = this.values.Labels || {
        items: [],
        mixed: false,
        action: this.actions.none,
      };

      this.formData.Albums = {
        action: albumsData.action || this.actions.none,
        mixed: albumsData.mixed || false,
        items: albumsData.items || [],
      };

      this.formData.Labels = {
        action: labelsData.action || this.actions.none,
        mixed: labelsData.mixed || false,
        items: labelsData.items || [],
      };

      this.previousFormData.Albums = {
        action: albumsData.action || this.actions.none,
        mixed: albumsData.mixed || false,
        items: [...(albumsData.items || [])],
      };

      this.previousFormData.Labels = {
        action: labelsData.action || this.actions.none,
        mixed: labelsData.mixed || false,
        items: [...(labelsData.items || [])],
      };

      // Update data properties for v-model
      this.albumItems = [...(albumsData.items || [])];
      this.labelItems = [...(labelsData.items || [])];
    },
    toggleOptions(fieldName) {
      const fieldData = this.values[fieldName];
      if (!fieldData) return [];

      const result = [
        { text: this.$gettext("Yes"), value: true },
        { text: this.$gettext("No"), value: false },
      ];

      if (fieldData.mixed) {
        result.splice(1, 0, { value: options.Mixed.String, text: options.Mixed.Label() });
      }

      return result;
    },
    getToggleValue(fieldName) {
      const fieldData = this.values[fieldName];

      if (!fieldData) {
        return false;
      }

      if (fieldData.mixed) {
        return options.Mixed.String;
      } else {
        return fieldData.value;
      }
    },
    toggleField(fieldName, event) {
      const classList = event.target.classList;

      if (classList.contains(iconUndo)) {
        this.deletedFields[fieldName] = false;
        this.formData[fieldName].action = this.actions.none;
        this.formData[fieldName].value = this.previousFormData[fieldName]?.value || "";

        // Restore the original mixed state when undoing changes
        if (this.previousFormData[fieldName]?.mixed !== undefined) {
          this.formData[fieldName].mixed = this.previousFormData[fieldName].mixed;
        }
      } else if (classList.contains(iconClear)) {
        this.deletedFields[fieldName] = true;

        if (fieldName === "Altitude") {
          this.formData[fieldName].action = this.actions.update;
          this.formData[fieldName].value = 0;
        } else {
          this.formData[fieldName].action = this.actions.remove;
          this.formData[fieldName].value = "";
        }
      }
    },
    getIcon(fieldType, fieldName) {
      const fieldData = this.values[fieldName];
      const isDeleted = this.deletedFields?.[fieldName];

      if (!fieldData) return;
      const previousField = this.previousFormData[fieldName];

      if (this.formData[fieldName].value !== previousField?.value || isDeleted) {
        return iconUndo;
      } else if (fieldData.mixed) {
        return iconClear;
      } else if (fieldType === "text-field" && fieldData.value !== null && fieldData.value !== "") {
        return iconClear;
      } else if (
        fieldType === "input-field" &&
        fieldData.value !== 0 &&
        fieldData.value !== null &&
        fieldData.value !== ""
      ) {
        return iconClear;
      }
    },
    getFieldData(fieldType, fieldName) {
      const fieldData = this.values[fieldName];
      const isDeleted = this.deletedFields?.[fieldName];

      if (!fieldData) return { value: "", placeholder: "", persistent: false };

      // Helper function to format numeric values
      const formatNumericValue = (value) => {
        if (["Lat", "Lng", "FNumber"].includes(fieldName)) {
          return parseFloat(value) || 0;
        } else if (
          ["Altitude", "Day", "Month", "Year", "Iso", "FocalLength", "CameraID", "LensID"].includes(fieldName)
        ) {
          return parseInt(value) || 0;
        }
        return value;
      };

      // Handle common states
      if (isDeleted) {
        return {
          value: fieldType === "input-field" ? 0 : "",
          placeholder: fieldType === "text-field" ? options.Deleted.Placeholder() : "",
          persistent: fieldType === "text-field",
        };
      }

      if (fieldData.mixed) {
        if (fieldType === "select-field") {
          const items = this.getItemsArray(fieldName, true);
          return {
            value: this.getValue(fieldName, items),
            placeholder: options.Mixed.Placeholder(),
            persistent: true,
            items,
          };
        }
        return {
          value: fieldType === "input-field" ? "" : "",
          placeholder: options.Mixed.Placeholder(),
          persistent: true,
        };
      }

      // Handle non-mixed values
      if (fieldType === "text-field") {
        return {
          value: fieldData.value || "",
          placeholder: "",
          persistent: false,
        };
      }

      if (fieldType === "input-field") {
        return {
          value: formatNumericValue(fieldData.value) || 0,
          placeholder: "",
          persistent: false,
        };
      }

      if (fieldType === "select-field") {
        const items = this.getItemsArray(fieldName, fieldData.mixed);

        if (fieldName === "Type" && fieldData.value) {
          const matchingOption = items.find((item) => item.value === fieldData.value);
          return {
            value: matchingOption ? matchingOption.text : fieldData.value,
            placeholder: "",
            persistent: false,
            items,
          };
        }

        return {
          value: formatNumericValue(fieldData.value),
          placeholder: "",
          persistent: false,
          items,
        };
      }
    },
    getValue(fieldName, items) {
      if (fieldName === "Day" || fieldName === "Month" || fieldName === "Year") {
        return items.find((item) => item.value === options.Mixed.ID)?.text;
      }
      if (fieldName === "Country") {
        return items.find((item) => item.Code === options.Mixed.ID)?.Name;
      }
      if (fieldName === "TimeZone") {
        return items.find((item) => item.ID === options.Mixed.ID)?.Name;
      }
      if (fieldName === "Type") {
        return items.find((item) => item.value === options.Mixed.String)?.text;
      }
    },
    getItemsArray(fieldName, isMixed) {
      if (fieldName === "Day") {
        return options.Batch(options.Days(), isMixed);
      }
      if (fieldName === "Month") {
        return options.Batch(options.MonthsShort(), isMixed);
      }
      if (fieldName === "Year") {
        return options.Batch(options.Years(1900), isMixed);
      }
      if (fieldName === "Country") {
        return options.Batch(options.Countries(), isMixed);
      }
      if (fieldName === "TimeZone") {
        return options.Batch(options.TimeZones(), isMixed);
      }
      if (fieldName === "Type") {
        return options.Batch(options.PhotoTypes(), isMixed);
      }
    },
    openPhoto(index) {
      const targetIndex = typeof index === "number" ? index : 0;
      this.$lightbox.openView(this, targetIndex);
    },
    getLightboxContext(index = 0) {
      const photos = Array.isArray(this.model?.models) ? this.model.models : [];

      if (!photos.length) {
        return { models: [], index: 0, allowEdit: false };
      }

      const thumbs = Thumb.fromPhotos(photos);
      let targetIndex = typeof index === "number" ? index : 0;

      if (targetIndex < 0 || targetIndex >= thumbs.length) {
        targetIndex = 0;
      }

      return {
        models: thumbs,
        index: targetIndex,
        context: "batch-edit",
        allowEdit: false,
        allowSelect: false,
      };
    },
    isSelected(model) {
      return this.model.isSelected(model.UID);
    },
    onClick(ev) {
      // Closes dialog when user clicks on background and model data is unchanged.
      if (!ev || !ev?.target?.classList?.contains("v-overlay__scrim")) {
        return;
      }
      ev.preventDefault();
      this.onClose();
    },
    onSelectClick(ev, index, select) {
      // Handle v-checkbox update:model-value event (ev will be boolean)
      if (select !== false) {
        this.toggle(this.model.models[index]);
      }
    },
    toggle(photo) {
      this.model.toggle(photo.UID);
      this.updateToggleAll();
      this.allSelectedLength = this.model.getLengthOfAllSelected();
    },
    updateToggleAll() {
      this.isAllSelected = this.model.selection.every((photo) => photo.selected);
    },
    toggleAll(value) {
      // Handle v-checkbox update:model-value event (value will be boolean)
      this.isAllSelected = value;
      this.model.toggleAll(this.isAllSelected);
      this.allSelectedLength = this.model.getLengthOfAllSelected();
    },
    onMouseDown(ev, index) {
      this.mouseDown.index = index;
      this.mouseDown.scrollY = window.scrollY;
      this.mouseDown.timeStamp = ev.timeStamp;
    },
    onClose() {
      // Closes the dialog only if there are no unsaved changes.
      if (this.hasUnsavedChanges()) {
        this.$refs?.dialog?.animateClick();
      } else {
        this.close();
      }
    },
    close() {
      // Close the dialog.
      this.$emit("close");
    },
    updateLatLng(latlng) {
      const newLat = parseFloat(latlng[0]) || 0;
      const newLng = parseFloat(latlng[1]) || 0;

      const previousLat = this.previousFormData.Lat?.value || 0;
      const previousLng = this.previousFormData.Lng?.value || 0;

      this.formData.Lat.value = newLat;
      this.formData.Lng.value = newLng;

      // Only set action to update if values actually changed
      this.formData.Lat.action = newLat !== previousLat ? this.actions.update : this.actions.none;
      this.formData.Lng.action = newLng !== previousLng ? this.actions.update : this.actions.none;

      this.deletedFields.Lat = false;
      this.deletedFields.Lng = false;
    },
    onLocationChanged(data) {
      if (data && data.lat !== undefined && data.lng !== undefined) {
        this.updateLatLng([data.lat, data.lng]);
      }
      // Update country when location is changed and country data is available
      if (data?.location?.country) {
        this.formData.Country.value = data.location.country;
        this.formData.Country.action = this.actions.update;
      }
      this.deletedFields.Lat = false;
      this.deletedFields.Lng = false;
    },
    onLocationDelete() {
      this.deletedFields.Lat = true;
      this.deletedFields.Lng = true;
      this.formData.Lat.action = this.actions.update;
      this.formData.Lng.action = this.actions.update;
      this.formData.Lat.value = 0;
      this.formData.Lng.value = 0;
    },
    onLocationUndo() {
      // Reset deleted state
      this.deletedFields.Lat = false;
      this.deletedFields.Lng = false;

      // Reset form actions and values to original
      this.formData.Lat.action = this.actions.none;
      this.formData.Lng.action = this.actions.none;
      this.formData.Lat.value = this.previousFormData.Lat?.value || 0;
      this.formData.Lng.value = this.previousFormData.Lng?.value || 0;
    },
    adjustLocation() {
      this.locationDialog = true;
    },
    confirmLocation(data) {
      if (data && data.lat !== undefined && data.lng !== undefined) {
        this.updateLatLng([data.lat, data.lng]);
        this.onLocationChanged(data);
      }

      this.locationDialog = false;
    },
    save(close) {
      this.saving = true;

      // Filter form data to only include fields with changes
      const filteredFormData = this.getFilteredFormData();

      if (!filteredFormData || Object.keys(filteredFormData).length === 0) {
        this.saving = false;
        if (close) {
          this.$emit("close");
        }
        return Promise.resolve();
      }

      // Get currently selected photo UIDs from the model
      const currentlySelectedUIDs = this.model.selection.filter((photo) => photo.selected).map((photo) => photo.id);

      return this.model
        .save(currentlySelectedUIDs, filteredFormData)
        .then(() => {
          // Save response already includes updated values, so reuse them to avoid a second POST.
          this.values = this.model.values;
          this.setFormData();

          this.$notify.success(this.$gettext("Changes successfully saved"));

          if (close) {
            this.$emit("close");
          }
        })
        .catch(() => {
          this.$notify.error(this.$gettext("Failed to save changes"));
        })
        .finally(() => {
          this.saving = false;
        });
    },
    getFilteredFormData() {
      const filtered = {};

      for (const [key, field] of Object.entries(this.formData)) {
        if (field) {
          // For Albums and Labels (Items type)
          if (key === "Albums" || key === "Labels") {
            // Ensure action is always "none" instead of null when there are no changes
            const action = field.action || this.actions.none;

            // Only include in filtered data if there's an actual change
            if (action !== this.actions.none) {
              const processedItems = field.items.map((item) => {
                const itemCopy = { ...item };
                delete itemCopy.isNew;
                return itemCopy;
              });

              filtered[key] = {
                ...field,
                action,
                items: processedItems,
              };
            }
          } else if (field.action && field.action !== this.actions.none) {
            // For regular fields with changes
            const isMixed = field.action !== this.actions.none ? false : field.mixed || false;
            filtered[key] = {
              action: field.action,
              mixed: isMixed,
              value: field.value,
            };
          }
        }
      }

      return filtered;
    },
    hasUnsavedChanges() {
      // Returns true if there are unsaved changes in the form.
      if (!this.formData) {
        return false;
      }
      const filtered = this.getFilteredFormData();
      return Object.keys(filtered).length > 0;
    },
    getFieldDisplayName(fieldName) {
      const displayNames = {
        Scan: this.$gettext("Scan"),
        Favorite: this.$gettext("Favorite"),
        Private: this.$gettext("Private"),
        Panorama: this.$gettext("Panorama"),
      };
      return displayNames[fieldName] || fieldName;
    },
    // PInputChipSelector methods
    onAlbumsUpdate(updatedItems) {
      this.albumItems = updatedItems;
      this.updateCollectionItems("Albums", updatedItems);
    },
    onLabelsUpdate(updatedItems) {
      this.labelItems = updatedItems;
      this.updateCollectionItems("Labels", updatedItems);
    },
    updateCollectionItems(collectionType, items) {
      // Check if there are any actual changes (add or remove actions)
      const hasChanges = items.some((item) => item.action === "add" || item.action === "remove");

      // Update the form data
      this.formData[collectionType].items = items;
      this.formData[collectionType].action = hasChanges ? this.actions.update : this.actions.none;
    },
    // Fetch available options for dropdowns
    async fetchAvailableOptions() {
      try {
        this.loading = true;

        const [albumsResponse, labelsResponse] = await Promise.all([
          Album.search({ count: MaxResults, type: "album", order: "name" }),
          Label.search({ count: MaxResults, order: "name", all: true }),
        ]);

        this.availableAlbumOptions = (albumsResponse.models || []).map((album) => ({
          value: album.UID,
          title: album.Title,
        }));

        this.availableLabelOptions = (labelsResponse.models || []).map((label) => ({
          value: label.UID,
          title: label.Name,
          slug: (label.Slug || "").toLowerCase(),
          customSlug: (label.CustomSlug || "").toLowerCase(),
        }));
      } catch (error) {
        console.error("Error fetching available options:", error);
      } finally {
        this.loading = false;
      }
    },
  },
};
</script>
