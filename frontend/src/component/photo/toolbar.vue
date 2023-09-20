<template>
  <v-form ref="form" lazy-validation
          dense autocomplete="off" class="p-photo-toolbar" accept-charset="UTF-8"
          :class="{'embedded': embedded}"
          @submit.prevent="updateQuery()">
    <v-toolbar flat :dense="$vuetify.breakpoint.smAndDown" :height="embedded ? 45 : undefined"
               class="page-toolbar" color="secondary">
      <template v-if="!embedded">
        <v-text-field :value="filter.q"
                      class="input-search background-inherit elevation-0"
                      solo hide-details clearable overflow single-line validate-on-blur
                      autocorrect="off"
                      autocapitalize="none"
                      browser-autocomplete="off"
                      :label="$gettext('Search')"
                      prepend-inner-icon="search"
                      color="secondary-dark"
                      @change="(v) => {updateFilter({'q': v})}"
                      @keyup.enter.native="(e) => updateQuery({'q': e.target.value})"
                      @click:clear="() => {updateQuery({'q': ''})}"
        ></v-text-field>

        <v-btn v-if="filter.latlng" icon :title="$gettext('Show more')" class="action-clear-location" @click.stop="clearLocation()">
          <v-icon>location_off</v-icon>
        </v-btn>

        <v-btn icon class="hidden-xs-only action-reload" :title="$gettext('Reload')" @click.stop="refresh()">
          <v-icon>refresh</v-icon>
        </v-btn>

        <v-btn v-if="settings.view === 'cards'" icon :title="$gettext('Toggle View')" @click.stop="setView('list')">
          <v-icon>view_list</v-icon>
        </v-btn>
        <v-btn v-else-if="settings.view === 'list'" icon :title="$gettext('Toggle View')"
               @click.stop="setView('mosaic')">
          <v-icon>view_comfy</v-icon>
        </v-btn>
        <v-btn v-else icon :title="$gettext('Toggle View')" @click.stop="setView('cards')">
          <v-icon>view_column</v-icon>
        </v-btn>

        <v-btn v-if="canDelete && context === 'archive' && config.count.archived > 0" icon
               class="hidden-sm-and-down action-delete-all"
               :title="$gettext('Delete All')" @click.stop="deleteAll()">
          <v-icon>delete_sweep</v-icon>
        </v-btn>
        <v-btn v-else-if="canUpload" icon class="hidden-sm-and-down action-upload"
               :title="$gettext('Upload')" @click.stop="showUpload()">
          <v-icon>cloud_upload</v-icon>
        </v-btn>

        <v-btn icon class="p-expand-search" :title="$gettext('Expand Search')"
               @click.stop="searchExpanded = !searchExpanded">
          <v-icon>{{ searchExpanded ? 'keyboard_arrow_up' : 'keyboard_arrow_down' }}</v-icon>
        </v-btn>
      </template>
      <template v-else>
        <v-spacer></v-spacer>
        <v-btn v-if="canAccessLibrary" icon :title="$gettext('Open')" class="action-open-tab" @click.stop="openInTab">
          <v-icon size="20">tab</v-icon>
        </v-btn>
        <v-btn v-if="onClose !== undefined" icon :title="$gettext('Close')" class="action-close" @click.stop="onClose">
          <v-icon>close</v-icon>
        </v-btn>
      </template>
    </v-toolbar>

    <v-card v-show="searchExpanded"
            class="pt-1 page-toolbar-expanded"
            flat
            color="secondary-light">
      <v-card-text>
        <v-layout row wrap>
          <v-flex xs12 sm6 md3 pa-2 class="p-countries-select">
            <v-select :value="filter.country"
                      :label="$gettext('Country')"
                      :menu-props="{'maxHeight':346}"
                      flat solo hide-details
                      color="secondary-dark"
                      background-color="secondary"
                      item-value="ID"
                      item-text="Name"
                      :items="countryOptions"
                      class="input-countries"
                      @change="(v) => {updateQuery({'country': v})}"
            >
            </v-select>
          </v-flex>
          <v-flex xs12 sm6 md3 pa-2 class="p-camera-select">
            <v-select :value="filter.camera"
                      :label="$gettext('Camera')"
                      :menu-props="{'maxHeight':346}"
                      flat solo hide-details
                      color="secondary-dark"
                      background-color="secondary"
                      item-value="ID"
                      item-text="Name"
                      :items="cameraOptions"
                      @change="(v) => {updateQuery({'camera': v})}">
            </v-select>
          </v-flex>
          <v-flex xs12 sm6 md3 pa-2 class="p-view-select">
            <v-select id="viewSelect"
                      :value="settings.view"
                      :label="$gettext('View')"
                      flat solo hide-details
                      color="secondary-dark"
                      background-color="secondary"
                      :items="options.views"
                      @change="(v) => {setView(v)}">
            </v-select>
          </v-flex>
          <v-flex xs12 sm6 md3 pa-2 class="p-time-select">
            <v-select :value="filter.order"
                      :label="$gettext('Sort Order')"
                      :menu-props="{'maxHeight':400}"
                      flat solo hide-details
                      color="secondary-dark"
                      background-color="secondary"
                      :items="options.sorting"
                      @change="(v) => {updateQuery({'order': v})}">
            </v-select>
          </v-flex>
          <v-flex xs12 sm6 md3 pa-2 class="p-year-select">
            <v-select :value="filter.year"
                      :label="$gettext('Year')"
                      :menu-props="{'maxHeight':346}"
                      flat solo hide-details
                      color="secondary-dark"
                      background-color="secondary"
                      item-value="value"
                      item-text="text"
                      :items="yearOptions()"
                      @change="(v) => {updateQuery({'year': v})}">
            </v-select>
          </v-flex>
          <v-flex xs12 sm6 md3 pa-2 class="p-month-select">
            <v-select :value="filter.month"
                      :label="$gettext('Month')"
                      :menu-props="{'maxHeight':346}"
                      flat solo hide-details
                      color="secondary-dark"
                      background-color="secondary"
                      item-value="value"
                      item-text="text"
                      :items="monthOptions()"
                      @change="(v) => {updateQuery({'month': v})}">
            </v-select>
          </v-flex>
          <!-- v-flex xs12 sm6 md3 pa-2 class="p-lens-select">
              <v-select @change="dropdownChange"
                        :label="labels.lens"
                        flat solo hide-details
                        color="secondary-dark"
                        background-color="secondary-light"
                        item-value="ID"
                        item-text="Model"
                        v-model="filter.lens"
                        :items="lensOptions">
              </v-select>
          </v-flex -->
          <v-flex xs12 sm6 md3 pa-2 class="p-color-select">
            <v-select :value="filter.color"
                      :label="$gettext('Color')"
                      :menu-props="{'maxHeight':346}"
                      flat solo hide-details
                      color="secondary-dark"
                      background-color="secondary"
                      item-value="Slug"
                      item-text="Name"
                      :items="colorOptions()"
                      @change="(v) => {updateQuery({'color': v})}">
            </v-select>
          </v-flex>
          <v-flex xs12 sm6 md3 pa-2 class="p-category-select">
            <v-select :value="filter.label"
                      :label="$gettext('Category')"
                      :menu-props="{'maxHeight':346}"
                      flat solo hide-details
                      color="secondary-dark"
                      background-color="secondary"
                      item-value="Slug"
                      item-text="Name"
                      :items="categoryOptions"
                      @change="(v) => {updateQuery({'label': v})}">
            </v-select>
          </v-flex>
        </v-layout>
      </v-card-text>
    </v-card>
    <p-photo-delete-dialog
      :show="dialog.delete"
      :text="$gettext('Are you sure you want to delete all archived pictures?')"
      :action="$gettext('Delete All')"
      @cancel="dialog.delete = false" @confirm="batchDelete">
    </p-photo-delete-dialog>
  </v-form>
</template>
<script>
import Event from "pubsub-js";
import * as options from "options/options";
import Api from "common/api";
import Notify from "common/notify";

export default {
  name: 'PPhotoToolbar',
  props: {
    context: {
      type: String,
      default: 'photos',
    },
    filter: {
      type: Object,
      default: () => {
      },
    },
    staticFilter: {
      type: Object,
      default: () => {
      },
    },
    updateFilter: {
      type: Function,
      default: () => {
      },
    },
    updateQuery: {
      type: Function,
      default: () => {
      },
    },
    settings: {
      type: Object,
      default: () => {
      },
    },
    refresh: {
      type: Function,
      default: () => {
      },
    },
    onClose: {
      type: Function,
      default: undefined,
    },
    embedded: {
      type: Boolean,
      default: false
    },
  },
  data() {
    const features = this.$config.settings().features;
    const readonly = this.$config.get("readonly");
    return {
      experimental: this.$config.get("experimental"),
      isFullScreen: !!document.fullscreenElement,
      config: this.$config.values,
      readonly: readonly,
      canUpload: !readonly && !this.embedded && this.$config.allow("files", "upload") && features.upload,
      canDelete: !readonly && !this.embedded && this.$config.allow("photos", "delete") && features.delete,
      canAccessLibrary: this.$config.allow("photos", "access_library"),
      searchExpanded: false,
      all: {
        countries: [{ID: "", Name: this.$gettext("All Countries")}],
        cameras: [{ID: 0, Name: this.$gettext("All Cameras")}],
        lenses: [{ID: 0, Name: this.$gettext("All Lenses")}],
        colors: [{Slug: "", Name: this.$gettext("All Colors")}],
        categories: [{Slug: "", Name: this.$gettext("All Categories")}],
        months: [{value: 0, text: this.$gettext("All Months")}],
        years: [{value: 0, text: this.$gettext("All Years")}],
      },
      options: {
        'views': [
          {value: 'mosaic', text: this.$gettext('Mosaic')},
          {value: 'cards', text: this.$gettext('Cards')},
          {value: 'list', text: this.$gettext('List')},
        ],
        'sorting': [
          {value: 'newest', text: this.$gettext('Newest First')},
          {value: 'oldest', text: this.$gettext('Oldest First')},
          {value: 'added', text: this.$gettext('Recently Added')},
          {value: 'edited', text: this.$gettext('Recently Edited')},
          {value: 'name', text: this.$gettext('File Name')},
          {value: 'size', text: this.$gettext('File Size')},
          {value: 'duration', text: this.$gettext('Video Duration')},
          {value: 'relevance', text: this.$gettext('Most Relevant')},
          {value: 'similar', text: this.$gettext('Visual Similarity')},
        ],
      },
      dialog: {
        delete: false,
      },
    };
  },
  computed: {
    countryOptions() {
      return this.all.countries.concat(this.config.countries);
    },
    cameraOptions() {
      return this.all.cameras.concat(this.config.cameras);
    },
    lensOptions() {
      return this.all.lenses.concat(this.config.lenses);
    },
    categoryOptions() {
      return this.all.categories.concat(this.config.categories);
    },
  },
  methods: {
    colorOptions() {
      return this.all.colors.concat(options.Colors());
    },
    monthOptions() {
      return this.all.months.concat(options.Months());
    },
    yearOptions() {
      return this.all.years.concat(options.IndexedYears());
    },
    setView(name) {
      if (name) {
        this.refresh({'view': name});
      }
    },
    showUpload() {
      Event.publish("dialog.upload");
    },
    deleteAll() {
      if (!this.canDelete) {
        return;
      }

      this.dialog.delete = true;
    },
    clearLocation() {
      this.$router.push({ name: "browse" });
    },
    openInTab() {
      const url = this.$router.resolve({name: 'places_location', query: this.staticFilter}).href;
      if (url) {
        window.open(url, '_blank');
      }
    },
    batchDelete() {
      if (!this.canDelete) {
        return;
      }

      this.dialog.delete = false;

      Api.post("batch/photos/delete", {"all": true}).then(() => this.onDeleted());
    },
    onDeleted() {
      Notify.success(this.$gettext("Permanently deleted"));
      this.$clipboard.clear();
    },
  },
};
</script>
