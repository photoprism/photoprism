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
                                          v-model="query.camera"
                                          :items="options.cameras">
                                </v-select>
                            </v-flex>
                            <v-flex xs12 sm6 md3 pa-2>
                                <v-select @change="formChange"
                                          label="Sort By"
                                          flat solo
                                          color="blue-grey"
                                          v-model="dir"
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
            <div class="page-container photo-grid pt-3">
                <template v-for="photo in items">

                    <div class="photo hover-12">
                        <div class="info">{{ photo.TakenAt | moment("DD.MM.YYYY hh:mm:ss") }}<span class="right">{{ photo.CameraModel }}</span>
                        </div>
                        <div class="actions">
                        <span class="left">
                            <a class="action like" v-bind:class="{ favorite: photo.Favorite }"
                               v-on:click="likePhoto(photo)">
                                <i v-if="!photo.Favorite" class="far fa-heart"></i>
                                <i v-if="photo.Favorite" class="fas fa-heart"></i>
                            </a>
                        </span>
                            <span class="center" v-if="photo.Location">
                            <v-tooltip bottom>
                                <a slot="activator" class="location" target="_blank" :href="photo.getGoogleMapsLink()">{{ photo.Location.Country }}</a>
                                <span>{{ photo.Location.DisplayName }}</span>
                            </v-tooltip>
                        </span>
                            <span class="right">
                            <a class="action delete" v-on:click="deletePhoto(photo)">
                                <i class="fas fa-trash"></i>
                            </a>
                        </span>
                        </div>
                        <template v-for="file in photo.Files">
                            <img v-if="file.FileType === 'jpg'"
                                 :src="'/api/v1/files/' + file.ID + '/square_thumbnail?size=250'">
                        </template>
                    </div>
                </template>
            </div>
            <div style="clear: both"></div>
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
            const resultCount = query.hasOwnProperty('count') ? parseInt(query['count']) : 70;
            const resultPage = query.hasOwnProperty('page') ? parseInt(query['page']) : 1;
            const resultOffset = resultCount * (resultPage - 1);
            const order = query.hasOwnProperty('order') ? query['order'] : 'taken_at';
            const dir = query.hasOwnProperty('dir') ? query['dir'] : '';
            const q = query.hasOwnProperty('q') ? query['q'] : '';
            const view = query.hasOwnProperty('view') ? query['view'] : 'tile';

            return {
                'advandedSearch': false,
                'items': [],
                'query': {
                    category: '',
                    country: '',
                    camera: '',
                    after: '',
                    before: '',
                    favorites_only: '',
                    q: q,
                },
                'options': {
                    'categories': [ { value: '', text: 'All Categories' }, { value: 'junction', text: 'Junction' }, { value: 'tourism', text: 'Tourism'}, { value: 'historic', text: 'Historic'} ],
                    'countries': [{ value: '', text: 'All Countries' }, { value: 'de', text: 'Germany' }, { value: 'ca', text: 'Canada'}, { value: 'us', text: 'United States'}],
                    'cameras': [{ value: '', text: 'All Cameras' }, { value: 'iPhone SE', text: 'iPhone SE' }, { value: 'EOS 6D', text: 'Canon EOS 6D'}],
                    'sorting': [{ value: '', text: 'Sort by date taken' }, { value: 'imported', text: 'Sort by date imported'}, { value: 'score', text: 'Sort by relevance' }],
                },
                'page': resultPage,
                'order': order,
                'dir': dir,
                'view': view,
                'resultCount': resultCount,
                'resultOffset': resultOffset,
                'resultTotal': 'Many',
                'lastQuery': {},
                'submitTimeout': false,
            };
        },
        methods: {
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
                    order: this.order !== '' ? this.order + ' ' + this.dir : '',
                };

                Object.assign(params, this.query);

                // Don't query the same data more than once
                if (_.isEqual(this.lastQuery, params)) return;

                this.lastQuery = params;

                // Set URL hash
                const urlParams = {
                    count: this.resultCount,
                    page: this.page,
                    order: this.order,
                    dir: this.dir,
                };

                Object.assign(urlParams, this.query);

                this.$router.replace({query: urlParams});

                Photo.search(params).then(response => {
                    console.log(response);
                    this.resultTotal = parseInt(response.headers['x-result-total']);
                    this.resultCount = parseInt(response.headers['x-result-count']);
                    this.resultOffset = parseInt(response.headers['x-result-offset']);
                    this.items = response.models;
                    this.$alert.info(this.items.length + ' photos found');
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
