<template>
    <div v-infinite-scroll="loadMore" infinite-scroll-disabled="loadMoreDisabled" infinite-scroll-distance="10">
        <v-container fluid>
            <v-btn @click.stop="dialog= true">Dialog</v-btn>
            <v-dialog v-model="dialog" dark fullscreen transition="dialog-bottom-transition">
                <v-card dark>
                <v-layout row wrap justify-center class="px-4 py-5">
                    <v-flex md8 xs12>
                        <v-card dark flat>
                            <v-img src="/static/img/tagcloud.jpg" aspect-ratio="1" class="mb-5 mx-5"></v-img>
                        </v-card>
                    </v-flex>
                    <v-flex md4 xs12>
                        <v-card dark flat>
                            <v-card-text>
                                <form>
                                <v-text-field
                                        v-model="Title"
                                        label="Title"
                                        placeholder="Tagcloud"
                                ></v-text-field>
                                <v-spacer></v-spacer>
                                <v-text-field
                                        v-model="artist"
                                        label="Artist"
                                        placeholder="Unknown"
                                ></v-text-field>
                                <v-spacer></v-spacer>
                                <v-text-field
                                        v-model="taken"
                                        label="Taken at"
                                        placeholder="02/02/19 00:02"
                                ></v-text-field>
                                <v-spacer></v-spacer>
                                <v-text-field
                                        v-model="location"
                                        label="Location"
                                        placeholder="Berlin"
                                ></v-text-field>
                                <v-spacer></v-spacer>
                                <v-text-field
                                        v-model="camera"
                                        label="Camera"
                                        placeholder="Iphone 5S"
                                ></v-text-field>
                                <v-spacer></v-spacer>
                                <v-text-field
                                        v-model="lense"
                                        label="Lense"
                                        placeholder="xxx"
                                ></v-text-field>
                                <v-spacer></v-spacer>
                                <v-text-field
                                        v-model="aperture"
                                        label="Aperture"
                                        placeholder=""
                                ></v-text-field>
                                <v-spacer></v-spacer>
                                <v-text-field
                                        v-model="focal"
                                        label="Focal Length"
                                        placeholder=""
                                ></v-text-field>
                                <v-spacer></v-spacer>
                                <v-text-field
                                        v-model="color"
                                        label="Color"
                                        placeholder="unknown"
                                ></v-text-field>
                                <v-spacer></v-spacer>

                                </form>
                            <v-combobox
                                    v-model="model2"
                                    :filter="filter"
                                    :hide-no-data="!search"
                                    :items="items"
                                    :search-input.sync="search"
                                    hide-selected
                                    label="Tags"
                                    multiple
                                    small-chips
                            >
                                <template v-slot:no-data>
                                    <v-list-tile>
                                        <span class="subheading">Create</span>
                                        <v-chip
                                                color= "blue"
                                                label
                                                small
                                        >
                                            search
                                        </v-chip>
                                    </v-list-tile>
                                </template>
                                <template v-slot:selection="{ item, parent, selected }">
                                    <v-chip
                                            v-if="item === Object(item)"
                                            color= "primary"
                                            :selected="selected"
                                            label
                                            small
                                    >
        <span class="pr-2">
          item text
        </span>
                                        <v-icon
                                                small
                                                @click="parent.selectItem(item)"
                                        >close</v-icon>
                                    </v-chip>
                                </template>
                                <template v-slot:item="{ index, item }">
                                    <v-list-tile-content>
                                        <v-text-field
                                                v-if="editing === item"
                                                v-model="editing.text"
                                                autofocus
                                                flat
                                                background-color="transparent"
                                                hide-details
                                                solo
                                                @keyup.enter="edit(index, item)"
                                        ></v-text-field>
                                        <v-chip
                                                v-else
                                                color="red"
                                                dark
                                                label
                                                small
                                        >
                                            item text
                                        </v-chip>
                                    </v-list-tile-content>
                                </template>
                            </v-combobox>
                            </v-card-text>
                        </v-card>
                    </v-flex>
                </v-layout>
                </v-card>
            </v-dialog>
        </v-container>
    </div>
</template>

<script>
    import Photo from 'model/photo';

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
                'dialog' : true,
                'dialog2': false,
                'search': null,
                'activator': null,
                'attach': null,
                'colors': ['green', 'purple', 'indigo', 'primary', 'success', 'orange'],
                'color': '',
                'editing': null,
                'index': -1,
                'items': [
                    { header: 'Select a tag or create one' },
                    {text: 'Cat', color: 'primary'},
                    {text: 'Sun', color: 'red'},
                    {text: 'Dog', color: 'primary'},
                    {text: 'Holiday', color: 'primary'},
                    {text: 'Tiger', color: 'primary'},
                    {text: 'Soup', color: 'primary'},
                    {text: 'Night', color: 'primary'},
                    {text: 'Table', color: 'primary'},
                    {text: 'Apple', color: 'primary'},
                    {text: 'Frog', color: 'primary'},
                ],
                'nonce': 1,
                'menu': false,
                'model': [

                ],
                'model2': [
                    {text: 'Cat', color: 'primary'},
                    {text: 'Dog', color: 'primary'},
                    {text: 'Holiday', color: 'primary'},
                    {text: 'Tiger', color: 'primary'},
                    {text: 'Soup', color: 'primary'},
                    {text: 'Night', color: 'primary'},
                    {text: 'Table', color: 'primary'},
                    {text: 'Apple', color: 'primary'},
                    {text: 'Frog', color: 'primary'},
                ],
                'x': 0,
                'y': 0,
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
