<template>
    <v-form lazy-validation dense
            ref="form" autocomplete="off" class="p-photo-search" accept-charset="UTF-8"
            @submit.prevent="filterChange">
        <v-toolbar flat color="secondary">
            <v-text-field class="pt-3 pr-3 p-search-field"
                          browser-autocomplete="off"
                          single-line
                          :label="labels.search"
                          prepend-inner-icon="search"
                          clearable
                          color="secondary-dark"
                          @click:clear="clearQuery"
                          v-model="filter.q"
                          @keyup.enter.native="filterChange"
            ></v-text-field>

            <v-spacer></v-spacer>

            <v-btn icon @click.stop="refresh" class="hidden-xs-only">
                <v-icon>refresh</v-icon>
            </v-btn>

            <v-btn icon v-if="settings.view === 'details'" @click.stop="setView('list')">
                <v-icon>view_list</v-icon>
            </v-btn>
            <v-btn icon v-else-if="settings.view === 'list'" @click.stop="setView('mosaic')">
                <v-icon>view_comfy</v-icon>
            </v-btn>
            <v-btn icon v-else @click.stop="setView('details')">
                <v-icon>view_column</v-icon>
            </v-btn>

            <v-btn icon @click.stop="showUpload()" v-if="!this.$config.values.readonly" class="hidden-md-and-down">
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
                                  item-value="code"
                                  item-text="name"
                                  v-model="filter.country"
                                  :items="countryOptions">
                        </v-select>
                    </v-flex>
                    <v-flex xs12 sm6 md3 pa-2 class="p-year-select">
                        <v-select @change="dropdownChange"
                                  :label="labels.year"
                                  flat solo hide-details
                                  color="secondary-dark"
                                  item-value="year"
                                  item-text="label"
                                  v-model="filter.year"
                                  :items="yearOptions">
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
                    <v-flex xs12 sm6 md3 pa-2 class="p-camera-select">
                        <v-select @change="dropdownChange"
                                  :label="labels.camera"
                                  flat solo hide-details
                                  color="secondary-dark"
                                  item-value="ID"
                                  item-text="CameraModel"
                                  v-model="filter.camera"
                                  :items="cameraOptions">
                        </v-select>
                    </v-flex>
                    <v-flex xs12 sm6 md3 pa-2 class="p-lens-select">
                        <v-select @change="dropdownChange"
                                  :label="labels.lens"
                                  flat solo hide-details
                                  color="secondary-dark"
                                  item-value="ID"
                                  item-text="LensModel"
                                  v-model="filter.lens"
                                  :items="lensOptions">
                        </v-select>
                    </v-flex>
                    <v-flex xs12 sm6 md3 pa-2 class="p-color-select">
                        <v-select @change="dropdownChange"
                                  :label="labels.color"
                                  flat solo hide-details
                                  color="secondary-dark"
                                  item-value="name"
                                  item-text="label"
                                  v-model="filter.color"
                                  :items="colorOptions">
                        </v-select>
                    </v-flex>
                    <v-flex xs12 sm6 md3 pa-2 class="p-category-select">
                        <v-select @change="dropdownChange"
                                  :label="labels.category"
                                  flat solo hide-details
                                  color="secondary-dark"
                                  item-value="LabelName"
                                  item-text="Title"
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

    export default {
        name: 'p-photo-search',
        props: {
            dirty: Boolean,
            filter: Object,
            settings: Object,
            refresh: Function,
            filterChange: Function,
        },
        data() {
            return {
                config: this.$config.values,
                searchExpanded: false,
                all: {
                    countries: [{ code: "", name: this.$gettext("All Countries")}],
                    cameras: [{ID: 0, CameraModel: this.$gettext("All Cameras")}],
                    lenses: [{ID: 0, LensModel: this.$gettext("All Lenses")}],
                    colors: [{label: this.$gettext("All Colors"), name: ""}],
                    categories: [{LabelName: "", Title: this.$gettext("All Categories")}],
                },
                options: {
                    'views': [
                        {value: 'mosaic', text: this.$gettext('Mosaic')},
                        {value: 'details', text: this.$gettext('Details')},
                        {value: 'list', text: this.$gettext('List')},
                    ],
                    'sorting': [
                        {value: 'imported', text: this.$gettext('Recently imported')},
                        {value: 'newest', text: this.$gettext('Newest first')},
                        {value: 'oldest', text: this.$gettext('Oldest first')},
                    ],
                },
                labels: {
                    search: this.$gettext("Search"),
                    view: this.$gettext("View"),
                    country: this.$gettext("Country"),
                    camera: this.$gettext("Camera"),
                    lens: this.$gettext("Lens"),
                    year: this.$gettext("Year"),
                    color: this.$gettext("Color"),
                    category: this.$gettext("Category"),
                    sort: this.$gettext("Sort By"),
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
            colorOptions() {
                return this.all.colors.concat(this.config.colors);
            },
            categoryOptions() {
                return this.all.categories.concat(this.config.categories);
            },
            yearOptions() {
                let result = [{"year": 0, "label": this.$gettext("All Years")}];

                if (this.config.years) {
                    for (let i = 0; i < this.config.years.length; i++) {
                        result.push({"year": this.config.years[i], "label": this.config.years[i].toString()});
                    }
                }

                return result;
            },
        },
        methods: {
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
                Event.publish("upload.show");
            }
        },
    };
</script>
