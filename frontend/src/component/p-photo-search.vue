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

            <v-btn icon v-if="settings.view === 'tiles'" @click.stop="setView('details')">
                <v-icon>view_column</v-icon>
            </v-btn>

            <v-btn icon v-if="settings.view === 'details'" @click.stop="setView('list')">
                <v-icon>view_list</v-icon>
            </v-btn>

            <v-btn icon v-if="settings.view === 'list'" @click.stop="setView('mosaic')">
                <v-icon>view_comfy</v-icon>
            </v-btn>

            <v-btn icon v-if="settings.view === 'mosaic'" @click.stop="setView('tiles')">
                <v-icon>view_module</v-icon>
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
                                  :items="options.countries">
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
                                  :items="options.cameras">
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
                    <v-flex xs6 pa-2 class="p-time-after">
                        <v-menu v-model="showAfterPicker"
                                :close-on-content-click="false"
                                :nudge-right="40"
                                transition="scale-transition"
                                offset-y
                                min-width="290px"
                        >
                            <template v-slot:activator="{ on }">
                                <v-text-field v-model="filter.after"
                                              :label="labels.after"
                                              prepend-inner-icon="date_range"
                                              clearable
                                              flat solo hide-details
                                              @change="datepickerChange"
                                              @click:clear="clearAfter"
                                              color="secondary-dark"
                                              v-on="on"
                                ></v-text-field>
                            </template>
                            <v-date-picker v-model="filter.after" color="secondary-dark"
                                           @input="datepickerChange">
                            </v-date-picker>
                        </v-menu>
                    </v-flex>
                    <v-flex xs6 pa-2 class="p-time-before">
                        <v-menu v-model="showBeforePicker"
                                :close-on-content-click="false"
                                :nudge-right="40"
                                transition="scale-transition"
                                offset-y
                                min-width="290px"
                        >
                            <template v-slot:activator="{ on }">
                                <v-text-field v-model="filter.before"
                                              :label="labels.before"
                                              prepend-inner-icon="date_range"
                                              flat solo hide-details
                                              clearable
                                              color="secondary-dark"
                                              @change="datepickerChange"
                                              @click:clear="clearBefore"
                                              v-on="on"
                                ></v-text-field>
                            </template>
                            <v-date-picker v-model="filter.before" color="secondary-dark"
                                           @input="datepickerChange">
                            </v-date-picker>
                        </v-menu>
                    </v-flex>
                </v-layout>
            </v-card-text>
        </v-card>
    </v-form>
</template>
<script>
    export default {
        name: 'p-photo-search',
        props: {
            filter: Object,
            settings: Object,
            refresh: Function,
            filterChange: Function,
        },
        data() {
            const cameras = [{ID: 0, CameraModel: this.$gettext('All Cameras')}].concat(this.$config.getValue('cameras'));
            const countries = [{
                code: '',
                name: this.$gettext('All Countries')
            }].concat(this.$config.getValue('countries'));

            return {
                searchExpanded: false,
                options: {
                    'views': [
                        {value: 'tiles', text: this.$gettext('Tiles')},
                        {value: 'mosaic', text: this.$gettext('Mosaic')},
                        {value: 'details', text: this.$gettext('Details')},
                        {value: 'list', text: this.$gettext('List')},
                    ],
                    'countries': countries,
                    'cameras': cameras,
                    'sorting': [
                        {value: 'imported', text: this.$gettext('Recently imported')},
                        {value: 'newest', text: this.$gettext('Newest first')},
                        {value: 'oldest', text: this.$gettext('Oldest first')},
                    ],
                },
                showAfterPicker: false,
                showBeforePicker: false,
                labels: {
                    search: this.$gettext("Search"),
                    view: this.$gettext("View"),
                    country: this.$gettext("Country"),
                    camera: this.$gettext("Camera"),
                    sort: this.$gettext("Sort By"),
                    before: this.$gettext("Taken before"),
                    after: this.$gettext("Taken after"),
                },
            };
        },
        methods: {
            dropdownChange() {
                this.filterChange();

                if (window.innerWidth < 600) {
                    this.searchExpanded = false;
                }
            },
            datepickerChange() {
                this.showAfterPicker = false;
                this.showBeforePicker = false;

                this.dropdownChange();
            },
            setView(name) {
                this.settings.view = name;
                this.filterChange();
            },
            clearBefore() {
                this.filter.before = '';
                this.datepickerChange();
            },
            clearAfter() {
                this.filter.after = '';
                this.datepickerChange();
            },
            clearQuery() {
                this.filter.q = '';
                this.filterChange();
            },
        }
    };
</script>
