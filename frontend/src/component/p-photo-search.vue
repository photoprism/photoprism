<template>
    <v-form ref="form" autocomplete="off" class="p-photo-search" lazy-validation @submit.prevent="filterChange" dense>
        <v-toolbar flat color="secondary">
            <v-text-field class="pt-3 pr-3 p-search-field"
                          autocomplete="off"
                          single-line
                          label="Search"
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
                                  label="Country"
                                  flat solo hide-details
                                  color="secondary-dark"
                                  item-value="LocCountryCode"
                                  item-text="LocCountry"
                                  v-model="filter.country"
                                  :items="options.countries">
                        </v-select>
                    </v-flex>
                    <v-flex xs12 sm6 md3 pa-2 class="p-camera-select">
                        <v-select @change="dropdownChange"
                                  label="Camera"
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
                                  label="View"
                                  flat solo hide-details
                                  color="secondary-dark"
                                  v-model="settings.view"
                                  :items="options.views"
                                  id="viewSelect">
                        </v-select>
                    </v-flex>
                    <v-flex xs12 sm6 md3 pa-2 class="p-time-select">
                        <v-select @change="dropdownChange"
                                  label="Sort By"
                                  flat solo hide-details
                                  color="secondary-dark"
                                  v-model="filter.order"
                                  :items="options.sorting">
                        </v-select>
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
            const cameras = [{ID: 0, CameraModel: 'All Cameras'}].concat(this.$config.getValue('cameras'));
            const countries = [{
                LocCountryCode: '',
                LocCountry: 'All Countries'
            }].concat(this.$config.getValue('countries'));

            return {
                searchExpanded: false,
                options: {
                    'views': [
                        {value: 'tiles', text: 'Tiles'},
                        {value: 'mosaic', text: 'Mosaic'},
                        {value: 'details', text: 'Details'},
                        {value: 'list', text: 'List'},
                    ],
                    'countries': countries,
                    'cameras': cameras,
                    'sorting': [
                        {value: 'newest', text: 'Newest first'},
                        {value: 'oldest', text: 'Oldest first'},
                        {value: 'imported', text: 'Recently imported'},
                    ],
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
            setView(name) {
                this.settings.view = name;
                this.filterChange();
            },
            clearQuery() {
                this.filter.q = '';
                this.filterChange();
            },
        }
    };
</script>
