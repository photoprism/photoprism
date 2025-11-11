<template>
  <v-form
    ref="form"
    validate-on="invalid-input"
    autocomplete="off"
    class="p-photo-toolbar"
    accept-charset="UTF-8"
    :class="{ embedded: embedded }"
    @submit.prevent="updateQuery()"
  >
    <v-toolbar
      :density="$vuetify.display.smAndDown && !embedded ? 'compact' : 'default'"
      :height="embedded ? 45 : undefined"
      class="page-toolbar"
      color="secondary"
    >
      <template v-if="!embedded">
        <v-text-field
          :model-value="filter.q"
          :density="density"
          hide-details
          clearable
          single-line
          overflow
          rounded="pill"
          variant="solo-filled"
          color="surface-variant"
          validate-on="invalid-input"
          autocorrect="off"
          autocapitalize="none"
          autocomplete="off"
          prepend-inner-icon="mdi-tune"
          :append-inner-icon="filter.latlng ? 'mdi-map-marker-off' : ''"
          :placeholder="$gettext('Search')"
          class="input-search background-inherit elevation-0"
          :class="{ 'input-search--expanded': expanded }"
          @update:model-value="
            (v) => {
              updateFilter({ q: v });
            }
          "
          @keyup.enter="() => onUpdate()"
          @keyup.esc.exact="() => hideExpansionPanel()"
          @click:prepend-inner.stop="toggleExpansionPanel"
          @click:append-inner.stop="clearLocation"
          @click:clear="
            () => {
              onUpdate({ q: '' });
            }
          "
        ></v-text-field>

        <v-btn-toggle
          :model-value="settings.view"
          :title="$gettext('Toggle View')"
          :density="$vuetify.display.smAndDown ? 'comfortable' : 'default'"
          base-color="secondary"
          variant="flat"
          rounded="pill"
          mandatory
          border
          group
          class="ms-1"
        >
          <v-btn
            value="cards"
            icon="mdi-view-column"
            class="ps-1 action-view-cards"
            @click="setView('cards')"
          ></v-btn>
          <v-btn
            v-if="listView"
            value="list"
            icon="mdi-view-list"
            class="action-view-list"
            @click="setView('list')"
          ></v-btn>
          <v-btn
            value="mosaic"
            icon="mdi-view-comfy"
            class="pe-1 action-view-mosaic"
            @click="setView('mosaic')"
          ></v-btn>
        </v-btn-toggle>

        <v-btn
          v-if="canDelete && context === 'archive' && config.count.archived > 0"
          :title="$gettext('Delete All')"
          icon="mdi-delete-sweep"
          class="action-delete-all ms-1"
          @click.stop="deleteAll"
        >
        </v-btn>

        <p-action-menu
          v-if="$vuetify.display.mdAndUp"
          :items="menuActions"
          button-class="ms-1"
        ></p-action-menu>
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

    <div class="toolbar-expansion-panel">
      <v-expand-transition>
        <v-card v-show="expanded" flat color="secondary">
          <v-card-text class="dense">
            <v-row align="center" dense>
              <v-col cols="12" sm="6" md="3" class="p-countries-select">
                <v-select
                  :model-value="filter.country"
                  :label="$gettext('Country')"
                  :menu-props="{ maxHeight: 346 }"
                  single-line
                  hide-details
                  variant="solo-filled"
                  :density="density"
                  :items="countryOptions"
                  item-title="Name"
                  item-value="ID"
                  class="input-countries"
                  @update:model-value="
                    (v) => {
                      onUpdate({ country: v });
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
                  :density="density"
                  :items="cameraOptions"
                  item-title="Name"
                  item-value="ID"
                  @update:model-value="
                    (v) => {
                      onUpdate({ camera: v });
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
                  :density="density"
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
                  :density="density"
                  :items="sortOptions"
                  item-title="text"
                  item-value="value"
                  @update:model-value="
                    (v) => {
                      onUpdate({ order: v });
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
                  :density="density"
                  :items="yearOptions()"
                  item-title="text"
                  item-value="value"
                  @update:model-value="
                    (v) => {
                      onUpdate({ year: v });
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
                  :density="density"
                  :items="monthOptions()"
                  item-title="text"
                  item-value="value"
                  @update:model-value="
                    (v) => {
                      onUpdate({ month: v });
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
                  :density="density"
                  :items="colorOptions()"
                  item-title="Name"
                  item-value="Slug"
                  @update:model-value="
                    (v) => {
                      onUpdate({ color: v });
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
                  :density="density"
                  :items="categoryOptions"
                  item-title="Name"
                  item-value="Slug"
                  @update:model-value="
                    (v) => {
                      onUpdate({ label: v });
                    }
                  "
                >
                </v-select>
              </v-col>
            </v-row>
          </v-card-text>
        </v-card>
      </v-expand-transition>
    </div>
    <p-confirm-dialog
      :visible="dialog.delete"
      :text="$gettext(`Delete all?`)"
      :action="$gettext('Yes')"
      icon="mdi-delete-sweep-outline"
      @close="dialog.delete = false"
      @confirm="batchDelete"
    >
    </p-confirm-dialog>
  </v-form>
</template>
<script>
import * as options from "options/options";
import $api from "common/api";
import $notify from "common/notify";
import links from "common/links";

import PActionMenu from "component/action/menu.vue";
import PConfirmDialog from "component/confirm/dialog.vue";

export default {
  name: "PPhotoToolbar",
  components: {
    PActionMenu,
    PConfirmDialog,
  },
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
    const features = this.$config.getSettings().features;
    const readonly = this.$config.get("readonly");

    return {
      expanded: false,
      experimental: this.$config.get("experimental"),
      isFullScreen: !!document.fullscreenElement,
      isSuperAdmin: this.$session.isSuperAdmin(),
      config: this.$config.values,
      readonly: readonly,
      canUpload: !readonly && !this.embedded && this.$config.allow("files", "upload") && features.upload,
      canDelete: !readonly && !this.embedded && this.$config.allow("photos", "delete") && features.delete,
      canAccessLibrary: this.$config.allow("photos", "access_library"),
      featSettings: features.settings,
      listView: this.$config.getSettings()?.search?.listView,
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
    density() {
      return this.$vuetify.display.smAndDown ? "compact" : "comfortable";
    },
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
      if (this.$config.getSettings()?.search?.listView) {
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
    showExpansionPanel() {
      if (!this.expanded) {
        this.expanded = true;
      }
    },
    hideExpansionPanel() {
      if (this.expanded) {
        this.expanded = false;
      }
    },
    toggleExpansionPanel() {
      this.expanded = !this.expanded;
    },
    menuActions() {
      return [
        {
          name: "refresh",
          icon: "mdi-refresh",
          text: this.$gettext("Refresh"),
          shortcut: "Ctrl-R",
          visible: true,
          click: () => {
            this.refresh();
          },
        },
        {
          name: "upload",
          icon: "mdi-cloud-upload",
          text: this.$gettext("Upload") + "…",
          shortcut: "Ctrl-U",
          visible: this.canUpload && this.context !== "archive" && this.context !== "hidden",
          click: () => {
            this.showUpload();
          },
        },
        {
          name: "docs",
          icon: "mdi-book-open-page-variant-outline",
          text: this.$gettext("Get Started"),
          visible: this.context !== "hidden",
          href: links.firstSteps,
          target: "_blank",
        },
        {
          name: "troubleshooting",
          icon: "mdi-book-open-page-variant-outline",
          text: this.$gettext("Troubleshooting"),
          visible: this.context === "hidden",
          href: links.missingPictures,
          target: "_blank",
        },
      ];
    },
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
        this.hideExpansionPanel();
        this.refresh({ view: name });
      }
    },
    showUpload() {
      this.$event.publish("dialog.upload");
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

      if (this.$isMobile) {
        this.$router.push(route);
      } else {
        // Open in a new tab on desktop browsers.
        const routeUrl = this.$router.resolve(route).href;

        if (routeUrl) {
          this.$util.openUrl(routeUrl);
        }
      }
    },
    onUpdate(v) {
      this.updateQuery(v);
    },
    batchDelete() {
      if (!this.canDelete) {
        return;
      }

      this.dialog.delete = false;

      $api.post("batch/photos/delete", { all: true }).then(() => this.onDeleted());
    },
    onDeleted() {
      $notify.success(this.$gettext("Permanently deleted"));
      this.$clipboard.clear();
    },
  },
};
</script>
