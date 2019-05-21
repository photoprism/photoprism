<template>
    <div class="p-page p-page-photos" v-infinite-scroll="loadMore" infinite-scroll-disabled="loadMoreDisabled"
         infinite-scroll-distance="10" infinite-scroll-listen-for-event="infiniteScrollRefresh">
        <v-form ref="form" lazy-validation @submit="formChange" dense>
            <v-toolbar flat color="blue-grey lighten-4">
                <v-text-field class="pt-3 pr-3"
                              single-line
                              label="Search"
                              prepend-inner-icon="search"
                              clearable
                              color="blue-grey"
                              @click:clear="clearQuery"
                              v-model="filter.q"
                              @keyup.enter.native="formChange"
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
                            <v-select @change="formChange"
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
                            <v-select @change="formChange"
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
                            <v-select @change="viewChange"
                                      label="View"
                                      flat solo hide-details
                                      color="blue-grey"
                                      v-model="view"
                                      :items="options.views"
                                      id="viewSelect">
                            </v-select>
                        </v-flex>
                        <v-flex xs12 sm6 md3 pa-2 id="timeFlex">
                            <v-select @change="formChange"
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
        <v-container fluid>
            <v-speed-dial
                    fixed
                    bottom
                    right
                    direction="top"
                    v-model="menuVisible"
                    transition="slide-y-reverse-transition"
                    class="p-photo-menu"
            >
                <v-btn
                        slot="activator"
                        color="grey darken-2"
                        dark
                        fab
                >
                    <v-icon v-if="selected.length === 0">menu</v-icon>
                    <span v-else>{{ selected.length }}</span>
                </v-btn>

                <v-btn
                        fab
                        dark
                        small
                        color="deep-purple lighten-2"
                        @click.stop="batchLike()"
                        :disabled="!selected.length"
                >
                    <v-icon>favorite</v-icon>
                </v-btn>
                <v-btn
                        fab
                        dark
                        small
                        color="cyan accent-4"
                        @click.stop="batchTag()"
                >
                    <v-icon>label</v-icon>
                </v-btn>
                <v-btn
                        fab
                        dark
                        small
                        color="teal accent-4"
                        @click.stop="batchDownload()"
                >
                    <v-icon>save</v-icon>
                </v-btn>
                <v-btn
                        fab
                        dark
                        small
                        color="yellow accent-4"
                        @click.stop="batchAlbum()"
                >
                    <v-icon>create_new_folder</v-icon>
                </v-btn>

                <v-btn
                        fab
                        dark
                        small
                        color="delete"
                        @click.stop="batchDelete()"
                        :disabled="!selected.length"
                >
                    <v-icon>delete</v-icon>
                </v-btn>
                <v-btn
                        fab
                        dark
                        small
                        color="grey"
                        @click.stop="$clipboard.clear()"
                        :disabled="!selected.length"
                >
                    <v-icon>clear</v-icon>
                </v-btn>
            </v-speed-dial>

            <p-photo-tiles v-if="view === 'tiles'" :photos="results" :selection="selected" :select="selectPhoto"
                           :open="openPhoto" :like="likePhoto"></p-photo-tiles>
            <p-photo-mosaic v-if="view === 'mosaic'" :photos="results" :selection="selected" :select="selectPhoto"
                            :open="openPhoto" :like="likePhoto"></p-photo-mosaic>
            <p-photo-details v-if="view === 'details'" :photos="results" :selection="selected"
                             :select="selectPhoto" :open="openPhoto" :like="likePhoto"></p-photo-details>
            <p-photo-list v-if="view === 'list'" :photos="results" :selection="selected" :select="selectPhoto"
                          :open="openPhoto" :like="likePhoto"></p-photo-list>
        </v-container>
    </div>
</template>

<script>
    import Photo from "model/photo";

    export default {
        name: 'photos',
        props: {},
        data() {
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
                'searchExpanded': false,
                'loadMoreDisabled': true,
                'menuVisible': false,
                'results': [],
                'selected': this.$clipboard.selection,
                'view': view,
                'pageSize': 60,
                'offset': 0,
                'filter': {
                    country: country,
                    camera: camera,
                    order: order,
                    q: q,
                },
                'lastFilter': {},
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
            batchLike() {
                this.$alert.warning("Not implemented yet");
                this.menuVisible = false;
            },
            batchDelete() {
                this.$alert.warning("Not implemented yet");
                this.menuVisible = false;
            },
            batchTag() {
                this.$alert.warning("Not implemented yet");
                this.menuVisible = false;
            },
            batchAlbum() {
                this.$alert.warning("Not implemented yet");
                this.menuVisible = false;
            },
            batchDownload() {
                this.$alert.warning("Not implemented yet");
                this.menuVisible = false;
            },
            selectPhoto(photo) {
                this.$clipboard.toggle(photo);
            },
            likePhoto(photo) {
                photo.PhotoFavorite = !photo.PhotoFavorite;
                photo.like(photo.PhotoFavorite);
            },
            deletePhoto(photo) {
                this.$alert.success('Photo deleted');
            },
            openLocation(photo) {
                if (photo.PhotoLat && photo.PhotoLong) {
                    this.$router.push({name: 'Places', query: {lat: photo.PhotoLat, long: photo.PhotoLong}});
                } else if (photo.LocName) {
                    this.$router.push({name: 'Places', query: {q: photo.LocName}});
                } else if (photo.LocCity) {
                    this.$router.push({name: 'Places', query: {q: photo.LocCity}});
                } else if (photo.LocCountry) {
                    this.$router.push({name: 'Places', query: {q: photo.LocCountry}});
                } else {
                    this.$router.push({name: 'Places', query: {q: photo.CountryName}});
                }
            },
            viewChange() {
                this.updateQuery();
            },
            formChange() {
                this.refreshList();
            },
            clearQuery() {
                this.query.q = '';
                this.refreshList();
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

                Object.assign(params, this.lastFilter);

                Photo.search(params).then(response => {
                    this.results = this.results.concat(response.models);

                    this.loadMoreDisabled = (response.models.length < this.pageSize);

                    if (this.loadMoreDisabled) {
                        this.$alert.info('All ' + this.results.length + ' photos loaded');
                    }
                });
            },
            updateQuery() {
                const query = {
                    view: this.view
                };

                Object.assign(query, this.filter);

                this.$router.replace({query: query});

                this.$nextTick(() => this.$emit("infiniteScrollRefresh"));
            },
            refreshList() {
                this.loadMoreDisabled = true;

                // Don't query the same data more than once
                if (JSON.stringify(this.lastFilter) === JSON.stringify(this.filter)) return;

                Object.assign(this.lastFilter, this.filter);

                this.offset = 0;

                this.updateQuery();

                const params = {
                    count: this.pageSize,
                    offset: this.offset,
                };

                Object.assign(params, this.filter);

                Photo.search(params).then(response => {
                    this.results = response.models;

                    this.loadMoreDisabled = (response.models.length < this.pageSize);

                    if (this.loadMoreDisabled) {
                        this.$alert.info(this.results.length + ' photos found');
                    } else {
                        this.$alert.info('More than 50 photos found');
                    }
                });
            },
        },
        beforeRouteLeave(to, from, next) {
            next();
        },
        created() {
            this.refreshList();
        },
    };
</script>
