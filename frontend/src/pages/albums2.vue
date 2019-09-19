<template>
    <div v-infinite-scroll="loadMore" infinite-scroll-disabled="loadMoreDisabled" infinite-scroll-distance="10">
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
                    <v-icon>{{ advandedSearch ? 'keyboard_arrow_up' : 'keyboard_arrow_down' }}</v-icon>
                </v-btn>
            </v-toolbar>

            <v-card class="pt-1"
                    flat
                    color="blue-grey lighten-5"
                    v-show="advandedSearch">
                <v-card-text>
                    <v-layout row wrap>
                        <v-flex xs12 sm6 md3 pa-2>
                            <v-select @change="formChange"
                                      label="Country"
                                      flat solo hide-details
                                      color="blue-grey"
                                      item-value="LocCountryCode"
                                      item-text="LocCountry"
                                      v-model="query.country"
                                      :items="options.countries">
                            </v-select>
                        </v-flex>
                        <v-flex xs12 sm6 md3 pa-2>
                            <v-select @change="formChange"
                                      label="Camera"
                                      flat solo hide-details
                                      color="blue-grey"
                                      item-value="ID"
                                      item-text="CameraModel"
                                      v-model="query.camera"
                                      :items="options.cameras">
                            </v-select>
                        </v-flex>
                        <v-flex xs12 sm6 md3 pa-2>
                            <v-select @change="formChange"
                                      label="View"
                                      flat solo hide-details
                                      color="blue-grey"
                                      v-model="query.view"
                                      :items="options.views">
                            </v-select>
                        </v-flex>
                        <v-flex xs12 sm6 md3 pa-2>
                            <v-select @change="formChange"
                                      label="Sort By"
                                      flat solo hide-details
                                      color="blue-grey"
                                      v-model="query.order"
                                      :items="options.sorting">
                            </v-select>
                        </v-flex>
                    </v-layout>
                </v-card-text>
            </v-card>
        </v-form>
        <v-container fluid>
            <v-layout wrap>
                <v-flex xs12>
                    <h1>
                        <v-text-field label="Album name*" required value="South Africa 2018" solo></v-text-field>
                    </h1>
                </v-flex>
                <v-flex xs12>
                    <v-textarea label="Description" value="Nice photo collection of the last vacation"
                                solo></v-textarea>
                </v-flex>
            </v-layout>
            <p> In case you want to add photos to this album
                --> go to the photos view --> select all photos you want to add
                --> click add to album in the bottom right menu and select this album.</p>
            <p> In case you want to remove photos from this album
                --> select all photos you want to remove
                --> click remove in the bottom right menu.</p>
        </v-container>

        <v-container fluid>
            <v-speed-dial
                    fixed
                    bottom
                    right
                    direction="top"
                    open-on-hover
                    transition="slide-y-reverse-transition"
                    style="right: 8px; bottom: 8px;"
            >
                <v-btn
                        slot="activator"
                        color="grey darken-2"
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
                        color="delete"
                >
                    <v-icon>delete</v-icon>
                </v-btn>
            </v-speed-dial>
            <v-data-table
                    :headers="listColumns"
                    :items="results"
                    hide-actions
                    class="elevation-1"
                    v-if="query.view === 'list'"
                    select-all
                    disable-initial-sort
                    item-key="ID"
                    v-model="selected"
                    :no-data-text="'No photos matched your search'"
            >
                <template slot="items" slot-scope="props">
                    <td>
                        <v-checkbox
                                v-model="props.selected"
                                primary
                                hide-details
                        ></v-checkbox>
                    </td>
                    <td>{{ props.item.PhotoTitle }}</td>
                    <td>{{ props.item.TakenAt | luxon:format('dd/MM/yyyy hh:mm:ss') }}</td>
                    <td>{{ props.item.LocCity }}</td>
                    <td>{{ props.item.LocCountry }}</td>
                    <td>{{ props.item.CameraModel }}</td>
                    <td>{{ props.item.PhotoFavorite ? 'Yes' : 'No' }}</td>
                </template>
            </v-data-table>

            <v-container grid-list-xs fluid class="pa-0" v-if="query.view === 'details'">
                <v-card v-if="results.length === 0">
                    <v-card-title primary-title>
                        <div>
                            <h3 class="title mb-3">No photos matched your search</h3>
                            <div>Try using other terms and search options such as category, country and camera.</div>
                        </div>
                    </v-card-title>
                </v-card>
                <v-layout row wrap>
                    <v-flex
                            v-for="(photo, index) in results"
                            :key="photo.ID"
                            xs12 sm6 md4 lg3 d-flex
                    >
                        <v-hover>
                            <v-card tile slot-scope="{ hover }"
                                    :dark="photo.selected"
                                    :class="photo.selected ? 'elevation-14 ma-1' : 'elevation-2 ma-2'">
                                <v-img
                                        :src="photo.getThumbnailUrl('tile_500')"
                                        aspect-ratio="1"
                                        v-bind:class="{ selected: photo.selected }"
                                        style="cursor: pointer"
                                        class="grey lighten-2"
                                        @click="openPhoto(index)"

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

                                    <v-btn v-if="hover || photo.selected" :flat="!hover" icon large absolute
                                           :ripple="false" style="right: 4px; bottom: 4px;"
                                           @click.stop.prevent="selectPhoto(photo)">
                                        <v-icon v-if="photo.selected" color="white">check_box</v-icon>
                                        <v-icon v-else color="white">check_box_outline_blank</v-icon>
                                    </v-btn>

                                    <v-btn v-if="hover || photo.PhotoFavorite" :flat="!hover" icon large absolute
                                           :ripple="false" style="bottom: 4px; left: 4px"
                                           @click.stop.prevent="likePhoto(photo)">
                                        <v-icon v-if="photo.PhotoFavorite" color="white">favorite
                                        </v-icon>
                                        <v-icon v-else color="white">favorite_border</v-icon>
                                    </v-btn>
                                </v-img>


                                <v-card-title primary-title class="pa-3">
                                    <div>
                                        <h3 class="subheading mb-2" :title="photo.PhotoTitle">{{ photo.PhotoTitle |
                                            truncate(80) }}</h3>
                                        <div class="caption">
                                            <v-icon size="14">date_range</v-icon>
                                            {{ photo.TakenAt | luxon:format('dd/MM/yyyy hh:mm:ss') }}
                                            <br/>
                                            <v-icon size="14">photo_camera</v-icon>
                                            {{ photo.getCamera() }}
                                            <br/>
                                            <v-icon size="14">location_on</v-icon>
                                            <span :title="photo.getFullLocation()">{{ photo.getLocation() }}</span>
                                        </div>
                                    </div>
                                </v-card-title>
                            </v-card>
                        </v-hover>
                    </v-flex>
                </v-layout>
            </v-container>

            <v-container grid-list-xs fluid class="pa-0" v-if="query.view === 'tiles'">
                <v-card v-if="results.length === 0">
                    <v-card-title primary-title>
                        <div>
                            <h3 class="headline mb-3">No photos matched your search</h3>
                            <div>Try using other terms and search options such as category, country and camera.</div>
                        </div>
                    </v-card-title>
                </v-card>
                <v-layout row wrap>
                    <v-flex
                            v-for="(photo, index) in results"
                            :key="photo.ID"
                            xs12 sm6 md3 lg2 d-flex
                            v-bind:class="{ selected: photo.selected }"
                    >
                        <v-hover>
                            <v-card tile slot-scope="{ hover }"
                                    :dark="photo.selected"
                                    :class="photo.selected ? 'elevation-14 ma-1' : hover ? 'elevation-6 ma-2' : 'elevation-2 ma-2'">
                                <v-img :src="photo.getThumbnailUrl('tile_500')"
                                       aspect-ratio="1"
                                       class="grey lighten-2"
                                       style="cursor: pointer"
                                       @click="openPhoto(index)"
                                >
                                    <v-layout
                                            slot="placeholder"
                                            fill-height
                                            align-center
                                            justify-center
                                            ma-0
                                    >
                                        <v-progress-circular indeterminate
                                                             color="grey lighten-5"></v-progress-circular>
                                    </v-layout>

                                    <v-btn v-if="hover || photo.selected" :flat="!hover" icon large absolute
                                           :ripple="false" style="right: 4px; bottom: 4px;"
                                           @click.stop.prevent="selectPhoto(photo)">
                                        <v-icon v-if="photo.selected" color="white">check_box</v-icon>
                                        <v-icon v-else color="white">check_box_outline_blank</v-icon>
                                    </v-btn>

                                    <v-btn v-if="hover || photo.PhotoFavorite" :flat="!hover" icon large absolute
                                           :ripple="false" style="bottom: 4px; left: 4px"
                                           @click.stop.prevent="likePhoto(photo)">
                                        <v-icon v-if="photo.PhotoFavorite" color="white">favorite</v-icon>
                                        <v-icon v-else color="white">favorite_border</v-icon>
                                    </v-btn>
                                </v-img>

                            </v-card>
                        </v-hover>
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

    export default {
        name: 'browse',
        props: {},
        data() {
            dialog: false;
            const query = this.$route.query;
            const order = query['order'] ? query['order'] : 'newest';
            const camera = query['camera'] ? parseInt(query['camera']) : 0;
            const q = query['q'] ? query['q'] : '';
            const country = query['country'] ? query['country'] : '';
            const view = query['view'] ? query['view'] : 'tiles';
            const cameras = [{ID: 0, CameraModel: 'All Cameras'}].concat(this.$config.getValue('cameras'));
            const countries = [{
                LocCountryCode: '',
                LocCountry: 'All Countries'
            }].concat(this.$config.getValue('countries'));

            return {
                'snackbarVisible': false,
                'snackbarText': '',
                'advandedSearch': false,
                'window': {
                    width: 0,
                    height: 0
                },
                'results': [],
                'query': {
                    view: view,
                    country: country,
                    camera: camera,
                    order: order,
                    q: q,
                },
                'options': {
                    'categories': [
                        {value: '', text: 'All Categories'},
                        {value: 'airport', text: 'Airport'},
                        {value: 'amenity', text: 'Amenity'},
                        {value: 'building', text: 'Building'},
                        {value: 'historic', text: 'Historic'},
                        {value: 'shop', text: 'Shop'},
                        {value: 'tourism', text: 'Tourism'},
                    ],
                    'views': [
                        {value: 'tiles', text: 'Tiles'},
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
                'listColumns': [
                    {text: 'Title', value: 'PhotoTitle'},
                    {text: 'Taken At', value: 'TakenAt'},
                    {text: 'City', value: 'LocCity'},
                    {text: 'Country', value: 'LocCountry'},
                    {text: 'Camera', value: 'CameraModel'},
                    {text: 'Favorite', value: 'PhotoFavorite'},
                ],
                'view': view,
                'loadMoreDisabled': true,
                'pageSize': 60,
                'offset': 0,
                'lastQuery': {},
                'submitTimeout': false,
                'selected': [],
                'dialog': false,
            };
        },
        destroyed() {
            window.removeEventListener('resize', this.handleResize)
        },
        methods: {
            handleResize() {
                this.window.width = window.innerWidth;
                this.window.height = window.innerHeight;
            },
            clearSelection() {
                for (let i = 0; i < this.selected.length; i++) {
                    this.selected[i].selected = false;
                }
                this.selected = [];
                this.updateSnackbar();
            },
            updateSnackbar(text) {
                if (!text) text = "";

                this.snackbarText = text;

                this.snackbarVisible = this.snackbarText !== "";
            },
            showSnackbar() {
                this.snackbarVisible = this.snackbarText !== "";
            },
            hideSnackbar() {
                this.snackbarVisible = false;
            },
            selectPhoto(photo, ev) {
                if (photo.selected) {
                    for (let i = 0; i < this.selected.length; i++) {
                        if (this.selected[i].id === photo.id) {
                            this.selected.splice(i, 1);
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
                photo.PhotoFavorite = !photo.PhotoFavorite;
                photo.like(photo.PhotoFavorite);
            },
            deletePhoto(photo) {
                this.$alert.success('Photo deleted');
            },
            formChange(event) {
                this.search();
            },
            clearQuery() {
                this.query.q = '';
                this.search();
            },
            openPhoto(index) {
                this.$viewer.show(this.results, index)
            },
            loadMore() {
                if (this.loadMoreDisabled) return;

                this.loadMoreDisabled = true;

                this.offset += this.pageSize;

                const params = {
                    count: this.pageSize,
                    offset: this.offset,
                };

                Object.assign(params, this.lastQuery);

                Photo.search(params).then(response => {
                    this.results = this.results.concat(response.models);

                    this.loadMoreDisabled = (response.models.length < this.pageSize);

                    if (this.loadMoreDisabled) {
                        this.$alert.info('All ' + this.results.length + ' photos loaded');
                    }
                });
            },
            search() {
                this.loadMoreDisabled = true;

                // Don't query the same data more than once:197
                if (JSON.stringify(this.lastQuery) === JSON.stringify(this.query)) return;

                Object.assign(this.lastQuery, this.query);

                this.offset = 0;

                this.$router.replace({query: this.query});

                const params = {
                    count: this.pageSize,
                    offset: this.offset,
                };

                Object.assign(params, this.query);

                Photo.search(params).then(response => {
                    this.results = response.models;

                    this.loadMoreDisabled = (response.models.length < this.pageSize);

                    if (this.loadMoreDisabled) {
                        this.$alert.info(this.results.length + ' photos found');
                    } else {
                        this.$alert.info('More than 50 photos found');
                    }
                });
            }
        },
        beforeRouteLeave(to, from, next) {
            next()
        },
        created() {
            window.addEventListener('resize', this.handleResize);
            this.handleResize();
            this.search();
        },
    };
</script>
