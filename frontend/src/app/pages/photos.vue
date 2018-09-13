<template>
    <div>
        <v-form ref="form" lazy-validation @submit="formChange" dense>
            <v-toolbar flat color="blue-grey lighten-4">
                <v-text-field class="pt-3 pr-3"
                              single-line
                              label="Search"
                              prepend-inner-icon="search"
                              clearable
                              color="blue-grey"
                              @click:clear="clearQuery"
                              v-model="query.q"
                              @keyup.enter.native="formChange"
                ></v-text-field>
                <!-- v-btn @click="formChange" color="secondary">Create Filter</v-btn -->
                <v-spacer></v-spacer>

                <v-btn icon @click="advandedSearch = !advandedSearch">
                    <v-icon>{{ advandedSearch ? 'keyboard_arrow_down' : 'keyboard_arrow_up' }}</v-icon>
                </v-btn>
            </v-toolbar>
            <v-slide-y-transition>
                <v-card class="pt-0"
                        flat
                        color="blue-grey lighten-4"
                        v-show="advandedSearch">
                    <v-card-text>
                        <v-layout row wrap>
                            <v-flex xs12 sm6 md3 pa-2>
                                <v-select @change="formChange"
                                          label="Category"
                                          flat solo
                                          color="blue-grey"
                                          v-model="query.category"
                                          :items="options.categories">
                                </v-select>
                            </v-flex>
                            <v-flex xs12 sm6 md3 pa-2>
                                <v-select @change="formChange"
                                          label="Country"
                                          flat solo
                                          color="blue-grey"
                                          v-model="query.country"
                                          :items="options.countries">
                                </v-select>
                            </v-flex>
                            <v-flex xs12 sm6 md3 pa-2>
                                <v-select @change="formChange"
                                          label="Camera"
                                          flat solo
                                          color="blue-grey"
                                          v-model="query.camera_id"
                                          :items="options.cameras">
                                </v-select>
                            </v-flex>
                            <v-flex xs12 sm6 md3 pa-2>
                                <v-select @change="formChange"
                                          label="Sort By"
                                          flat solo
                                          color="blue-grey"
                                          v-model="query.order"
                                          :items="options.sorting">
                                </v-select>
                            </v-flex>
                        </v-layout>
                    </v-card-text>
                </v-card>
            </v-slide-y-transition>
        </v-form>
        <v-container fluid>
            <v-speed-dial
                    fixed
                    bottom
                    right
                    direction="top"
                    open-on-hover
                    transition="slide-y-reverse-transition"
            >
                <v-btn
                        slot="activator"
                        color="blue-grey darken-1"
                        dark
                        fab
                >
                    <v-icon>menu</v-icon>
                </v-btn>
                <v-btn
                        fab
                        dark
                        small
                        color="deep-purple lighten-2"
                >
                    <v-icon>favorite</v-icon>
                </v-btn>
                <v-btn
                        fab
                        dark
                        small
                        color="cyan accent-4"
                >
                    <v-icon>youtube_searched_for</v-icon>
                </v-btn>
                <v-btn
                        fab
                        dark
                        small
                        color="teal accent-4"
                >
                    <v-icon>save</v-icon>
                </v-btn>
                <v-btn
                        fab
                        dark
                        small
                        color="yellow accent-4"
                >
                    <v-icon>create_new_folder</v-icon>
                </v-btn>

                <v-btn
                        fab
                        dark
                        small
                        color="red"
                >
                    <v-icon>delete</v-icon>
                </v-btn>
            </v-speed-dial>
            <v-container grid-list-xs fluid class="pa-0">
                <v-layout row wrap>
                    <v-flex
                            v-for="photo in results"
                            :key="photo.ID"
                            xs12 sm6 md3 lg2 d-flex
                            v-bind:class="{ selected: photo.selected }"
                            class="photo-tile"
                    >
                        <v-tooltip bottom>
                            <v-card-actions flat tile class="d-flex" slot="activator" @click="selectPhoto(photo)"
                                            @mouseover="overPhoto(photo)" @mouseleave="leavePhoto(photo)">
                                <v-img :src="'/api/v1/files/' + photo.FileID + '/square_thumbnail?size=500'"
                                       aspect-ratio="1"
                                       class="grey lighten-2"
                                >
                                    <v-layout
                                            slot="placeholder"
                                            fill-height
                                            align-center
                                            justify-center
                                            ma-0
                                    >
                                        <v-progress-circular indeterminate color="grey lighten-5"></v-progress-circular>
                                    </v-layout>
                                </v-img>
                            </v-card-actions>
                            <span>{{ photo.PhotoTitle }}<br/>{{ photo.TakenAt | moment('DD/MM/YYYY') }} / {{ photo.CameraModel }}</span>
                        </v-tooltip>
                    </v-flex>
                </v-layout>
            </v-container>
            <v-snackbar
                    v-model="snackbarVisible"
                    bottom
                    :timeout="0"
            >
                {{ snackbarText }}
                <v-btn
                        class="pr-0"
                        color="primary"
                        icon
                        flat
                        @click="clearSelection()"
                >
                    <v-icon>close</v-icon>
                </v-btn>
            </v-snackbar>
        </v-container>
    </div>
</template>

<script>
    import Photo from 'model/photo';
    import _ from 'lodash/lang';

    export default {
        name: 'photos',
        props: {},
        data() {
            const query = this.$route.query;
            const resultCount = query.hasOwnProperty('count') ? parseInt(query['count']) : 60;
            const resultPage = query.hasOwnProperty('page') ? parseInt(query['page']) : 1;
            const resultOffset = resultCount * (resultPage - 1);
            const order = query.hasOwnProperty('order') && query['order'] != "" ? query['order'] : 'taken_at DESC';
            const camera_id = query.hasOwnProperty('camera_id') ? parseInt(query['camera_id']) : '';
            const q = query.hasOwnProperty('q') ? query['q'] : '';
            const view = query.hasOwnProperty('view') ? query['view'] : 'tile';
            const cameras = [{value: '', text: 'All Cameras'}];

            console.log(this.$config.getValue('cameras'));
            this.$config.getValue('cameras').forEach(function (camera) {
                cameras.push({value: camera.ID, text: camera.CameraModel});
            });

            return {
                'snackbarVisible': false,
                'snackbarText': '',
                'advandedSearch': false,
                'results': [],
                'query': {
                    category: '',
                    country: '',
                    camera_id: camera_id,
                    order: order,
                    q: q,
                },
                'options': {
                    'categories': [
                        {value: '', text: 'All Categories'},
                        {value: 'junction', text: 'Junction'},
                        {value: 'tourism', text: 'Tourism'},
                        {value: 'historic', text: 'Historic'},
                    ],
                    'countries': [
                        {value: '', text: 'All Countries'},
                        {value: 'de', text: 'Germany'},
                        {value: 'ca', text: 'Canada'},
                        {value: 'us', text: 'United States'}
                    ],
                    'cameras': cameras,
                    'sorting': [
                        {value: 'taken_at DESC', text: 'Newest first'},
                        {value: 'taken_at', text: 'Oldest first'},
                        {value: 'created_at DESC', text: 'Recently imported'},
                    ],
                },
                'page': resultPage,
                'view': view,
                'resultCount': resultCount,
                'resultOffset': resultOffset,
                'resultTotal': 'Many',
                'lastQuery': {},
                'submitTimeout': false,
                'selected': []
            };
        },
        methods: {
            overPhoto(photo) {

            },
            leavePhoto(photo) {

            },
            clearSelection() {
                for (let i = 0; i < this.selected.length; i++) {
                    this.selected[i].selected = false;
                }
                this.selected = [];
                this.snackbarText = '';
                this.snackbarVisible = false;
            },
            selectPhoto(photo) {
                console.log(photo)
                if (photo.selected) {
                    for (let i = 0; i < this.selected.length; i++) {
                        if (this.selected[i].id === photo.id) {
                            this.selected.splice(i, 1)
                            break;
                        }
                    }

                    photo.selected = false;
                } else {
                    this.selected.push(photo);
                    photo.selected = true;
                }

                if (this.selected.length > 0) {
                    if (this.selected.length === 1) {
                        this.snackbarText = 'One photo selected';
                    } else {
                        this.snackbarText = this.selected.length + ' photos selected';
                    }
                    this.snackbarVisible = true;
                } else {
                    this.snackbarText = '';
                    this.snackbarVisible = false;
                }
            },
            likePhoto(photo) {
                photo.Favorite = !photo.Favorite;
            },
            deletePhoto(photo) {
                this.$alert.success('Photo deleted');
            },
            formChange(event) {
                this.refreshList();
            },
            clearQuery() {
                this.query.q = '';
                this.refreshList();
            },
            refreshList() {
                // Compose query parameters
                const params = {
                    count: this.resultCount,
                    offset: this.resultCount * (this.page - 1),
                };

                Object.assign(params, this.query);

                // Don't query the same data more than once
                if (_.isEqual(this.lastQuery, params)) return;

                this.lastQuery = params;

                // Set URL hash
                const urlParams = {
                    count: this.resultCount,
                    page: this.page,
                };

                Object.assign(urlParams, this.query);

                this.$router.replace({query: urlParams});

                Photo.search(params).then(response => {
                    console.log(response);
                    this.resultTotal = parseInt(response.headers['x-result-total']);
                    this.resultCount = parseInt(response.headers['x-result-count']);
                    this.resultOffset = parseInt(response.headers['x-result-offset']);
                    this.results = response.models;
                    this.$alert.info(this.results.length + ' photos found');
                });
            }
        },
        created() {
            this.refreshList();
        },
    };
</script>

<style scoped>
</style>
