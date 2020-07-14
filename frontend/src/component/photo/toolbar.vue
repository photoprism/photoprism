<template>
  <v-form lazy-validation dense
          ref="form" autocomplete="off" class="p-photo-toolbar" accept-charset="UTF-8"
          @submit.prevent="filterChange">
    <v-toolbar flat color="secondary">
      <v-text-field class="pt-3 pr-3 input-search"
                    browser-autocomplete="off"
                    single-line
                    :label="$gettext('Search')"
                    prepend-inner-icon="search"
                    clearable
                    color="secondary-dark"
                    @click:clear="clearQuery"
                    v-model="filter.q"
                    @keyup.enter.native="filterChange"
      ></v-text-field>

      <v-spacer></v-spacer>

      <v-btn icon @click.stop="refresh" class="hidden-xs-only action-reload">
        <v-icon>refresh</v-icon>
      </v-btn>

      <v-btn icon v-if="settings.view === 'cards'" @click.stop="setView('list')">
        <v-icon>view_list</v-icon>
      </v-btn>
      <v-btn icon v-else-if="settings.view === 'list'" @click.stop="setView('mosaic')">
        <v-icon>view_comfy</v-icon>
      </v-btn>
      <v-btn icon v-else @click.stop="setView('cards')">
        <v-icon>view_column</v-icon>
      </v-btn>

      <v-btn icon @click.stop="showUpload()" v-if="!$config.values.readonly && $config.feature('upload')"
             class="hidden-sm-and-down action-upload">
        <v-icon>cloud_upload</v-icon>
      </v-btn>

      <v-btn icon @click.stop="searchExpanded = !searchExpanded" class="p-expand-search">
        <v-icon>{{ searchExpanded ? 'keyboard_arrow_up' : 'keyboard_arrow_down' }}</v-icon>
      </v-btn>
    </v-toolbar>

    <v-card class="pt-1"
            flat
            color="secondary-light"
            v-show="searchExpanded">
      <v-card-text>
        <v-layout row wrap>
          <v-flex xs12 sm6 md3 pa-2 class="p-countries-select">
            <v-select @change="dropdownChange"
                      :label="labels.country"
                      flat solo hide-details
                      color="secondary-dark"
                      item-value="ID"
                      item-text="Name"
                      v-model="filter.country"
                      :items="countryOptions"
                      class="input-countries"
            >
            </v-select>
          </v-flex>
          <v-flex xs12 sm6 md3 pa-2 class="p-camera-select">
            <v-select @change="dropdownChange"
                      :label="labels.camera"
                      flat solo hide-details
                      color="secondary-dark"
                      item-value="ID"
                      item-text="Name"
                      v-model="filter.camera"
                      :items="cameraOptions">
            </v-select>
          </v-flex>
          <v-flex xs12 sm6 md3 pa-2 class="p-view-select">
            <v-select @change="dropdownChange"
                      :label="labels.view"
                      flat solo hide-details
                      color="secondary-dark"
                      v-model="settings.view"
                      :items="options.views"
                      id="viewSelect">
            </v-select>
          </v-flex>
          <v-flex xs12 sm6 md3 pa-2 class="p-time-select">
            <v-select @change="dropdownChange"
                      :label="labels.sort"
                      flat solo hide-details
                      color="secondary-dark"
                      v-model="filter.order"
                      :items="options.sorting">
            </v-select>
          </v-flex>
          <v-flex xs12 sm6 md3 pa-2 class="p-year-select">
            <v-select @change="dropdownChange"
                      :label="labels.year"
                      flat solo hide-details
                      color="secondary-dark"
                      item-value="value"
                      item-text="text"
                      v-model="filter.year"
                      :items="yearOptions()">
            </v-select>
          </v-flex>
          <v-flex xs12 sm6 md3 pa-2 class="p-month-select">
            <v-select @change="dropdownChange"
                      :label="labels.month"
                      flat solo hide-details
                      color="secondary-dark"
                      item-value="value"
                      item-text="text"
                      v-model="filter.month"
                      :items="monthOptions()">
            </v-select>
          </v-flex>
          <!-- v-flex xs12 sm6 md3 pa-2 class="p-lens-select">
              <v-select @change="dropdownChange"
                        :label="labels.lens"
                        flat solo hide-details
                        color="secondary-dark"
                        item-value="ID"
                        item-text="Model"
                        v-model="filter.lens"
                        :items="lensOptions">
              </v-select>
          </v-flex -->
          <v-flex xs12 sm6 md3 pa-2 class="p-color-select">
            <v-select @change="dropdownChange"
                      :label="labels.color"
                      flat solo hide-details
                      color="secondary-dark"
                      item-value="Slug"
                      item-text="Name"
                      v-model="filter.color"
                      :items="colorOptions()">
            </v-select>
          </v-flex>
          <v-flex xs12 sm6 md3 pa-2 class="p-category-select">
            <v-select @change="dropdownChange"
                      :label="labels.category"
                      flat solo hide-details
                      color="secondary-dark"
                      item-value="Slug"
                      item-text="Name"
                      v-model="filter.label"
                      :items="categoryOptions">
            </v-select>
          </v-flex>
        </v-layout>
      </v-card-text>
    </v-card>
  </v-form>
</template>
<script>
    import Event from "pubsub-js";
    import * as options from "options/options";

    export default {
        name: 'p-photo-toolbar',
        props: {
            dirty: Boolean,
            filter: Object,
            settings: Object,
            refresh: Function,
            filterChange: Function,
        },
        data() {
            return {
                experimental: this.$config.get("experimental"),
                isFullScreen: !!document.fullscreenElement,
                config: this.$config.values,
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
                        {value: 'added', text: this.$gettext('Recently added')},
                        {value: 'edited', text: this.$gettext('Recently edited')},
                        {value: 'newest', text: this.$gettext('Newest first')},
                        {value: 'oldest', text: this.$gettext('Oldest first')},
                        {value: 'name', text: this.$gettext('Sort by file name')},
                        {value: 'similar', text: this.$gettext('Group by similarity')},
                        {value: 'relevance', text: this.$gettext('Most relevant')},
                    ],
                },
                labels: {
                    search: this.$gettext("Search"),
                    view: this.$gettext("View"),
                    country: this.$gettext("Country"),
                    camera: this.$gettext("Camera"),
                    lens: this.$gettext("Lens"),
                    year: this.$gettext("Year"),
                    month: this.$gettext("Month"),
                    color: this.$gettext("Color"),
                    category: this.$gettext("Category"),
                    sort: this.$gettext("Sort Order"),
                    before: this.$gettext("Taken before"),
                    after: this.$gettext("Taken after"),
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
            dropdownChange() {
                this.filterChange();

                if (window.innerWidth < 600) {
                    this.searchExpanded = false;
                }
            },
            setView(name) {
                this.settings.view = name;
                this.filterChange();
            },
            clearQuery() {
                this.filter.q = '';
                this.filterChange();
            },
            showUpload() {
                Event.publish("dialog.upload");
            }
        },
    };
</script>
