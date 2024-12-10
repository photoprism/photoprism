<template>
  <v-form ref="form" validate-on="blur" autocomplete="off" class="p-photo-toolbar" accept-charset="UTF-8" :class="{ embedded: embedded }" @submit.prevent="updateQuery()">
    <v-toolbar flat :density="$vuetify.display.smAndDown ? 'compact' : 'default'" :height="embedded ? 45 : undefined" class="page-toolbar" color="secondary">
      <template v-if="!embedded">
        <v-text-field
          :model-value="filter.q"
          class="input-search background-inherit elevation-0 mb-3"
          hide-details
          clearable
          overflow
          single-line
          variant="plain"
          density="comfortable"
          validate-on="blur"
          autocorrect="off"
          autocapitalize="none"
          autocomplete="off"
          :placeholder="$gettext('Search')"
          prepend-inner-icon="mdi-magnify"
          color="surface-variant"
          @change="
            (v) => {
              updateFilter({ q: v });
            }
          "
          @keyup.enter="(e) => updateQuery({ q: e.target.value })"
          @click:clear="
            () => {
              updateQuery({ q: '' });
            }
          "
        ></v-text-field>

        <v-btn v-if="filter.latlng" icon :title="$gettext('Show more')" class="action-clear-location" @click.stop="clearLocation()">
          <v-icon>mdi-map-marker-off</v-icon>
        </v-btn>

        <v-btn icon class="hidden-xs action-reload" :title="$gettext('Reload')" @click.stop="refresh()">
          <v-icon>mdi-refresh</v-icon>
        </v-btn>

        <v-btn v-if="settings.view === 'list'" icon class="action-view-mosaic" :title="$gettext('Toggle View')" @click.stop="setView('mosaic')">
          <v-icon>mdi-view-comfy</v-icon>
        </v-btn>
        <v-btn v-else-if="settings.view === 'cards' && listView" icon class="action-view-list" :title="$gettext('Toggle View')" @click.stop="setView('list')">
          <v-icon>mdi-view-list</v-icon>
        </v-btn>
        <v-btn v-else-if="settings.view === 'cards'" icon class="action-view-mosaic" :title="$gettext('Toggle View')" @click.stop="setView('mosaic')">
          <v-icon>mdi-view-comfy</v-icon>
        </v-btn>
        <v-btn v-else icon class="action-view-cards" :title="$gettext('Toggle View')" @click.stop="setView('cards')">
          <v-icon>mdi-view-column</v-icon>
        </v-btn>
        <v-btn v-if="canDelete && context === 'archive' && config.count.archived > 0" icon class="hidden-sm-and-down action-delete-all" :title="$gettext('Delete All')" @click.stop="deleteAll()">
          <v-icon>mdi-delete-sweep</v-icon>
        </v-btn>
        <v-btn v-else-if="canUpload" icon class="hidden-sm-and-down action-upload" :title="$gettext('Upload')" @click.stop="showUpload()">
          <v-icon>mdi-cloud-upload</v-icon>
        </v-btn>

        <v-btn icon class="p-expand-search" :title="$gettext('Expand Search')" @click.stop="searchExpanded = !searchExpanded">
          <v-icon>{{ searchExpanded ? "mdi-chevron-up" : "mdi-chevron-down" }}</v-icon>
        </v-btn>
      </template>
      <template v-else>
        <v-spacer></v-spacer>
        <v-btn v-if="canAccessLibrary" icon :title="$gettext('Browse')" class="action-browse" @click.stop="onBrowse">
          <v-icon size="20">mdi-tab</v-icon>
        </v-btn>
        <v-btn v-if="onClose !== undefined" icon :title="$gettext('Close')" class="action-close" @click.stop="onClose">
          <v-icon>mdi-close</v-icon>
        </v-btn>
      </template>
    </v-toolbar>

    <v-card v-show="searchExpanded" class="pt-1 page-toolbar-expanded" flat color="secondary-lighten-1">
      <v-card-text class="dense">
        <v-row dense>
          <v-col cols="12" sm="6" md="3" class="p-countries-select">
            <v-select
              :model-value="filter.country"
              :label="$gettext('Country')"
              :menu-props="{ maxHeight: 346 }"
              single-line
              hide-details
              variant="solo-filled"
              density="comfortable"
              :items="countryOptions"
              item-title="Name"
              item-value="ID"
              class="input-countries"
              @update:model-value="
                (v) => {
                  updateQuery({ country: v });
                }
              "
            >
            </v-select>
          </v-col>
          <v-col cols="12" sm="6" md="3" class="p-camera-select">
            <v-select
              :model-value="filter.camera"
              :label="$gettext('Camera')"
              :menu-props="{ maxHeight: 346 }"
              single-line
              hide-details
              variant="solo-filled"
              density="comfortable"
              :items="cameraOptions"
              item-title="Name"
              item-value="ID"
              @update:model-value="
                (v) => {
                  updateQuery({ camera: v });
                }
              "
            >
            </v-select>
          </v-col>
          <v-col cols="12" sm="6" md="3" class="p-view-select">
            <v-select
              id="viewSelect"
              :model-value="settings.view"
              :label="$gettext('View')"
              single-line
              hide-details
              variant="solo-filled"
              density="comfortable"
              :items="viewOptions"
              item-title="text"
              item-value="value"
              @update:model-value="
                (v) => {
                  setView(v);
                }
              "
            >
            </v-select>
          </v-col>
          <v-col cols="12" sm="6" md="3" class="p-time-select">
            <v-select
              :model-value="filter.order"
              :label="$gettext('Sort Order')"
              :menu-props="{ maxHeight: 400 }"
              single-line
              variant="solo-filled"
              density="comfortable"
              :items="sortOptions"
              item-title="text"
              item-value="value"
              @update:model-value="
                (v) => {
                  updateQuery({ order: v });
                }
              "
            >
            </v-select>
          </v-col>
          <v-col cols="12" sm="6" md="3" class="p-year-select">
            <v-select
              :model-value="filter.year"
              :label="$gettext('Year')"
              :menu-props="{ maxHeight: 346 }"
              single-line
              variant="solo-filled"
              density="comfortable"
              :items="yearOptions()"
              item-title="text"
              item-value="value"
              @update:model-value="
                (v) => {
                  updateQuery({ year: v });
                }
              "
            >
            </v-select>
          </v-col>
          <v-col cols="12" sm="6" md="3" class="p-month-select">
            <v-select
              :model-value="filter.month"
              :label="$gettext('Month')"
              :menu-props="{ maxHeight: 346 }"
              single-line
              variant="solo-filled"
              density="comfortable"
              :items="monthOptions()"
              item-title="text"
              item-value="value"
              @update:model-value="
                (v) => {
                  updateQuery({ month: v });
                }
              "
            >
            </v-select>
          </v-col>
          <!-- v-col cols="12" sm="6" md="3" class="p-lens-select">
              <v-select @change="dropdownChange"
                        :label="labels.lens"
                        flat
                        variant="solo-filled"
                        hide-details
                        color="surface-variant"
                        bg-color="secondary-light"
                        item-value="ID"
                        item-title="Model"
                        v-model="filter.lens"
                        :items="lensOptions">
              </v-select>
          </v-col -->
          <v-col cols="12" sm="6" md="3" class="p-color-select">
            <v-select
              :model-value="filter.color"
              :label="$gettext('Color')"
              :menu-props="{ maxHeight: 346 }"
              single-line
              hide-details
              variant="solo-filled"
              density="comfortable"
              :items="colorOptions()"
              item-title="Name"
              item-value="Slug"
              @update:model-value="
                (v) => {
                  updateQuery({ color: v });
                }
              "
            >
            </v-select>
          </v-col>
          <v-col cols="12" sm="6" md="3" class="p-category-select">
            <v-select
              :model-value="filter.label"
              :label="$gettext('Category')"
              :menu-props="{ maxHeight: 346 }"
              single-line
              hide-details
              variant="solo-filled"
              density="comfortable"
              :items="categoryOptions"
              item-title="Name"
              item-value="Slug"
              @update:model-value="
                (v) => {
                  updateQuery({ label: v });
                }
              "
            >
            </v-select>
          </v-col>
        </v-row>
      </v-card-text>
    </v-card>
    <p-photo-delete-dialog :show="dialog.delete" :text="$gettext('Are you sure you want to delete all archived pictures?')" :action="$gettext('Delete All')" @cancel="dialog.delete = false" @confirm="batchDelete">
</p-photo-delete-dialog>
  </v-form>
</template>
<script>
import Event from "pubsub-js";
import * as options from "options/options";
import Api from "common/api";
import Notify from "common/notify";

export default {
  name: "PPhotoToolbar",
  props: {
    context: {
      type: String,
      default: "photos",
    },
    filter: {
      type: Object,
      default: () => {},
    },
    staticFilter: {
      type: Object,
      default: () => {},
    },
    updateFilter: {
      type: Function,
      default: () => {},
    },
    updateQuery: {
      type: Function,
      default: () => {},
    },
    settings: {
      type: Object,
      default: () => {},
    },
    refresh: {
      type: Function,
      default: () => {},
    },
    onClose: {
      type: Function,
      default: undefined,
    },
    embedded: {
      type: Boolean,
      default: false,
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
      listView: this.$config.settings()?.search?.listView,
      all: {
        countries: [{ ID: "", Name: this.$gettext("All Countries") }],
        cameras: [{ ID: 0, Name: this.$gettext("All Cameras") }],
        lenses: [{ ID: 0, Name: this.$gettext("All Lenses") }],
        colors: [{ Slug: "", Name: this.$gettext("All Colors") }],
        categories: [{ Slug: "", Name: this.$gettext("All Categories") }],
        months: [{ value: 0, text: this.$gettext("All Months") }],
        years: [{ value: 0, text: this.$gettext("All Years") }],
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
    categoryOptions() {
      return this.all.categories.concat(this.config.categories);
    },
    viewOptions() {
      if (this.$config.settings()?.search?.listView) {
        return [
          { value: "mosaic", text: this.$gettext("Mosaic") },
          { value: "cards", text: this.$gettext("Cards") },
          { value: "list", text: this.$gettext("List") },
        ];
      } else {
        return [
          { value: "mosaic", text: this.$gettext("Mosaic") },
          { value: "cards", text: this.$gettext("Cards") },
        ];
      }
    },
    sortOptions() {
      switch (this.context) {
        case "archive":
          return [
            { value: "newest", text: this.$gettext("Newest First") },
            { value: "oldest", text: this.$gettext("Oldest First") },
            { value: "added", text: this.$gettext("Recently Added") },
            { value: "archived", text: this.$gettext("Recently Archived") },
            { value: "title", text: this.$gettext("Picture Title") },
            { value: "name", text: this.$gettext("File Name") },
            { value: "size", text: this.$gettext("File Size") },
            { value: "duration", text: this.$gettext("Video Duration") },
          ];
        case "hidden":
        case "review":
          return [
            { value: "newest", text: this.$gettext("Newest First") },
            { value: "oldest", text: this.$gettext("Oldest First") },
            { value: "added", text: this.$gettext("Recently Added") },
            { value: "title", text: this.$gettext("Picture Title") },
            { value: "name", text: this.$gettext("File Name") },
            { value: "size", text: this.$gettext("File Size") },
            { value: "duration", text: this.$gettext("Video Duration") },
          ];
        default:
          return [
            { value: "newest", text: this.$gettext("Newest First") },
            { value: "oldest", text: this.$gettext("Oldest First") },
            { value: "added", text: this.$gettext("Recently Added") },
            { value: "edited", text: this.$gettext("Recently Edited") },
            { value: "title", text: this.$gettext("Picture Title") },
            { value: "name", text: this.$gettext("File Name") },
            { value: "size", text: this.$gettext("File Size") },
            { value: "duration", text: this.$gettext("Video Duration") },
            { value: "similar", text: this.$gettext("Visual Similarity") },
            { value: "relevance", text: this.$gettext("Most Relevant") },
          ];
      }
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
        if (name === "list" && !this.listView) {
          name = "mosaic";
        }

        this.refresh({ view: name });
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
    onBrowse() {
      const route = { name: "places_browse", query: this.staticFilter };
      const routeUrl = this.$router.resolve(route).href;
      if (routeUrl) {
        window.open(routeUrl, "_blank");
      }
    },
    batchDelete() {
      if (!this.canDelete) {
        return;
      }

      this.dialog.delete = false;

      Api.post("batch/photos/delete", { all: true }).then(() => this.onDeleted());
    },
    onDeleted() {
      Notify.success(this.$gettext("Permanently deleted"));
      this.$clipboard.clear();
    },
  },
};
</script>
