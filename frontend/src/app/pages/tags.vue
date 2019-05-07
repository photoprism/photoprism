<template>
    <div v-infinite-scroll="loadMore" infinite-scroll-disabled="loadMoreDisabled" infinite-scroll-distance="10">
        <v-form ref="form" lazy-validation @submit="formChange" dense>
            <v-toolbar flat color="blue-grey lighten-4">
                <h1 class="md-display-1">Tags</h1>
                <v-spacer></v-spacer>
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
                        <v-flex xs12 sm6 md3 pa-2>
                            <v-select @change="formChange"
                                      label="Groups"
                                      flat solo hide-details
                                      color="blue-grey"
                                      v-model="query.group"
                                      :items="options.groups">
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
                        @click.stop="dialog2 = true"
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
                    <td>Label</td>
                    <td>28/11/2019</td>
                    <td>#4</td>
                    <td>55</td>
                    <td> <v-btn color="success"
                                @click.stop="dialog = true">
                        Edit
                    </v-btn></td>
                </template>
            </v-data-table>
            <v-container fluid v-if="query.view === 'cloud'">
                <v-layout justify-space-around>
                    <v-flex>
                            <v-img src="/assets/img/tagcloud.jpg" aspect-ratio="1.7"  @click.stop="dialog = true"></v-img>
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
        <photoswipe :images="results" ref="gallery"></photoswipe>
        <v-dialog v-model="dialog" dark persistent max-width="600px">
            <v-card dark>
                <v-card-title>
                    <span class="headline">Edit tag - Cat</span>
                </v-card-title>
                <v-card-text>
                    <template>
                        <form>
                            <b>Translate:</b>
                            <v-select
                                    v-model="select"
                                    :items="items"
                                    label="Language"
                            ></v-select>
                            <v-text-field
                                    v-model="translation"
                                    label="Translation"
                            ></v-text-field>
                            <v-spacer></v-spacer>
                            <v-select
                                    v-model="select"
                                    :items="items"
                                    label="Language"
                            ></v-select>
                            <v-text-field
                                    v-model="translation"
                                    label="Translation"
                            ></v-text-field>
                            <v-spacer></v-spacer>
                            <b>Add to group:</b>
                            <v-select
                                    v-model="select"
                                    :items="items2"
                                    label="Select"
                            ></v-select>
                        </form>
                    </template>
                </v-card-text>
                <v-card-actions>
                    <v-spacer></v-spacer>
                    <v-btn color="success" flat @click="dialog = false">Cancel</v-btn>
                    <v-btn color="success" flat @click="dialog = false">apply</v-btn>
                </v-card-actions>
            </v-card>
        </v-dialog>
        <v-dialog v-model="dialog2" dark persistent max-width="600px">
            <v-card dark>
                <v-card-title>
                    <span class="headline">Add tags to group</span>
                </v-card-title>
                <v-card-text>
                    <template>
                        <form>
                            13 tags selected <br>
                        <v-spacer></v-spacer>
                            <v-select
                                    v-model="select"
                                    :items="items2"
                                    label="Group"
                            ></v-select>
                        </form>
                    </template>
                </v-card-text>
                <v-card-actions>
                    <v-spacer></v-spacer>
                    <v-btn color="success" flat @click="dialog2 = false">Cancel</v-btn>
                    <v-btn color="success" flat @click="dialog2 = false">apply</v-btn>
                </v-card-actions>
            </v-card>
        </v-dialog>
    </div>
</template>

<script>
    import Photo from 'model/photo';

    export default {
        name: 'tags',
        props: {},
        data() {
            const query = this.$route.query;
            const order = query['order'] ? query['order'] : 'newest';
            const camera = query['camera'] ? parseInt(query['camera']) : 0;
            const q = query['q'] ? query['q'] : '';
            const country = query['country'] ? query['country'] : '';
            const view = query['view'] ? query['view'] : 'cloud';
            const group = query['group'] ? query['group'] : '';

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
                        {value: 'cloud', text: 'Cloud'},
                        {value: 'list', text: 'List'},
                    ],
                    'groups':  [
                        {value: 'a', text: 'Animals'},
                        {value: 'b', text: 'People'},
                        ],
                    'sorting': [
                        {value: 'newest', text: 'Mostly used'},
                        {value: 'oldest', text: 'Rarely used'},
                    ],
                },
                'listColumns': [
                    {text: 'Label', value: 'PhotoTitle'},
                    {text: 'Created  At', value: 'TakenAt'},
                    {text: 'Id', value: 'LocCity'},
                    {text: 'Nr of photos', value: 'Nr'},
                    {text: 'Actions', value: 'Edit'},
                ],
                'view': view,
                'loadMoreDisabled': true,
                'pageSize': 60,
                'offset': 0,
                'lastQuery': {},
                'submitTimeout': false,
                'selected': [],
                'dialog': false,
                'dialog2': false,
                select: null,
                items: [
                    'English',
                    'German',
                    'French',
                    'Spanish'
                ],
                items2: [
                    'Holiday',
                    'Nature',
                    'Animals',
                ],
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
                this.refreshList();
            },
            clearQuery() {
                this.query.q = '';
                this.refreshList();
            },
            openPhoto(index) {
                this.$refs.gallery.openPhoto(index)
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
                    console.log(response);
                    this.results = this.results.concat(response.models);

                    this.loadMoreDisabled = (response.models.length < this.pageSize);

                    if (this.loadMoreDisabled) {
                        this.$alert.info('All ' + this.results.length + ' photos loaded');
                    }
                });
            },
            refreshList() {
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
            },
            translation() {
                return ''
            }
        },
        beforeRouteLeave(to, from, next) {
            next()
        },
        created() {
            window.addEventListener('resize', this.handleResize);
            this.handleResize();
            this.refreshList();
        },
    };
</script>
