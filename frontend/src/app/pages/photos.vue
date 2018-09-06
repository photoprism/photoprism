<template>
    <div>
        <v-form ref="form" lazy-validation @submit="formChange" dense>
    <v-toolbar>
        <v-text-field class="pt-3"
                      single-line
                      label="Search"
                      prepend-inner-icon="search"
                      clear-icon="mdi-close-circle"
                      clearable
                      v-model="query.q"
                      @keyup.enter.native="formChange"
        ></v-text-field>

        <v-spacer></v-spacer>
        <v-btn icon @click="advandedSearch = !advandedSearch">
            <v-icon>{{ advandedSearch ? 'keyboard_arrow_down' : 'keyboard_arrow_up' }}</v-icon>
        </v-btn>
    </v-toolbar><v-slide-y-transition>
        <v-card class="theme--light v-toolbar pt-0" style="box-shadow: 0 4px 4px -1px rgba(0,0,0,.2), 0 4px 5px 0 rgba(0,0,0,.14), 0 4px 10px 0 rgba(0,0,0,.12);" v-show="advandedSearch">


                    <v-card-text >
                        <v-container>
                            <v-layout row wrap>
                                <v-flex xs12 sm6 md3>
                                    <v-select @change="formChange" class="mb-2 mr-sm-2"
                                              title="Category"
                                              v-model="query.category"
                                              :options="{ 'junction': 'Junction', 'tourism': 'Tourism', 'historic': 'Historic' }"
                                              id="inlineFormCustomSelectPref">
                                        <option slot="first" :value="null"></option>
                                    </v-select>
                                </v-flex>

                                <v-flex xs12 sm6 md3>
                                    <v-select @change="formChange" class="mb-2 mr-sm-2"
                                              v-model="query.country"
                                              :options="{ '1': 'One', '2': 'Two', '3': 'Three' }"
                                              id="inlineFormCustomSelectPref">
                                        <option slot="first" :value="null">Country</option>
                                    </v-select>
                                </v-flex>
                                <v-flex xs12 sm6 md3>
                                    <v-select @change="formChange" class="mb-2 mr-sm-2"
                                              :v-model="query.camera"
                                              :options="{ '1': 'One', '2': 'Two', '3': 'Three' }"
                                              id="inlineFormCustomSelectPref">
                                        <option slot="first" :value="null">Camera Model</option>
                                    </v-select>
                                </v-flex>
                                <v-flex xs12 sm6 md3>
                                    <v-select @change="formChange" class="mb-2 mr-sm-2"
                                              v-model="dir"
                                              :options="{ 'asc': 'Ascending', 'desc': 'Descending' }"
                                              id="inlineFormCustomSelectPref">
                                        <option slot="first" :value="null">Sort Order</option>
                                    </v-select>
                                </v-flex>
                                <v-flex xs12 sm6 md3>
                                    <v-select @change="formChange" class="mb-2 mr-sm-2"
                                              v-model="view"
                                              :options="{ 'list': 'List View', 'tile': 'Tile View (small)', 'tile_large': 'Tile View (large)' }"
                                              id="inlineFormCustomSelectPref">
                                    </v-select>
                                </v-flex>

                                <v-flex xs12 sm6 md3>
                                    <v-text-field class="mb-2 mr-sm-2" title="After" type="date"></v-text-field>
                                </v-flex>

                                <v-flex xs12 sm6 md3>
                                    <v-text-field class="mb-2 mr-sm-2" title="Before" type="date"></v-text-field>
                                </v-flex>

                                <v-flex xs12 sm6 md3>
                                    <v-checkbox class="mb-2 mr-sm-2 mb-sm-0">
                                        Favorites only
                                    </v-checkbox>
                                </v-flex>
                            </v-layout>
                        </v-container>
                    </v-card-text>



        </v-card>
    </v-slide-y-transition>
        </v-form>
    <v-container fluid>



        <div class="page-container photo-grid pt-3">
            <template v-for="photo in items">

                <div class="photo hover-12">
                    <div class="info">{{ photo.TakenAt | moment("DD.MM.YYYY hh:mm:ss") }}<span class="right">{{ photo.CameraModel }}</span></div>
                    <div class="actions">
                        <span class="left">
                            <a class="action like" v-bind:class="{ favorite: photo.Favorite }" v-on:click="likePhoto(photo)">
                                <i v-if="!photo.Favorite" class="far fa-heart"></i>
                                <i v-if="photo.Favorite" class="fas fa-heart"></i>
                            </a>
                        </span>
                        <span class="center" v-if="photo.Location">
                            <v-tooltip bottom>
                                <a slot="activator" class="location" target="_blank" :href="photo.getGoogleMapsLink()">{{ photo.Location.Country }}</a>
                                <span :html="photo.Location.DisplayName"></span>
                            </v-tooltip>
                        </span>
                        <span class="right">
                            <a class="action delete" v-on:click="deletePhoto(photo)">
                                <i class="fas fa-trash"></i>
                            </a>
                        </span>
                    </div>
                <template v-for="file in photo.Files">
                    <img v-if="file.FileType === 'jpg'" :src="'/api/v1/files/' + file.ID + '/square_thumbnail?size=250'">
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
        props: {
        },
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
                this.$alert.success('Form change');
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
