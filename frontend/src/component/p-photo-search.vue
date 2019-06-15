<template>
    <v-form ref="form" autocomplete="off" class="p-photo-search" lazy-validation @submit.prevent="filterChange" dense>
        <v-toolbar flat color="blue-grey lighten-4">
            <v-text-field class="pt-3 pr-3"
                          autocomplete="off"
                          single-line
                          label="Search"
                          prepend-inner-icon="search"
                          clearable
                          color="blue-grey"
                          @click:clear="clearQuery"
                          v-model="filter.q"
                          @keyup.enter.native="filterChange"
                          id="search"
            ></v-text-field>

            <v-spacer></v-spacer>

            <v-btn icon @click="searchExpanded = !searchExpanded" id="advancedMenu">
                <v-icon>{{ searchExpanded ? 'keyboard_arrow_up' : 'keyboard_arrow_down' }}</v-icon>
            </v-btn>
        </v-toolbar>

        <v-card class="pt-1"
                flat
                color="blue-grey lighten-5"
                v-show="searchExpanded">
            <v-card-text>
                <v-layout row wrap>
                    <v-flex xs12 sm6 md3 pa-2 id="countriesFlex">
                        <v-select @change="filterChange"
                                  label="Country"
                                  flat solo hide-details
                                  color="blue-grey"
                                  item-value="LocCountryCode"
                                  item-text="LocCountry"
                                  v-model="filter.country"
                                  :items="options.countries">
                        </v-select>
                    </v-flex>
                    <v-flex xs12 sm6 md3 pa-2 id="cameraFlex">
                        <v-select @change="filterChange"
                                  label="Camera"
                                  flat solo hide-details
                                  color="blue-grey"
                                  item-value="ID"
                                  item-text="CameraModel"
                                  v-model="filter.camera"
                                  :items="options.cameras">
                        </v-select>
                    </v-flex>
                    <v-flex xs12 sm6 md3 pa-2 id="viewFlex">
                        <v-select @change="settingsChange"
                                  label="View"
                                  flat solo hide-details
                                  color="blue-grey"
                                  v-model="settings.view"
                                  :items="options.views"
                                  id="viewSelect">
                        </v-select>
                    </v-flex>
                    <v-flex xs12 sm6 md3 pa-2 id="timeFlex">
                        <v-select @change="filterChange"
                                  label="Sort By"
                                  flat solo hide-details
                                  color="blue-grey"
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
            filterChange: Function,
            settingsChange: Function,
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
            clearQuery() {
                this.filter.q = '';
                this.filterChange();
            },
        }
    };
</script>
