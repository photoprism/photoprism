<template>
    <div>
        <v-toolbar flat color="blue-grey lighten-4">
            <h1>Import</h1>
        </v-toolbar>
    <v-container fluid>
        <v-form>
        <p class="md-subheading">
            You have two possibilities to get your photos into photoprism.</p>
            <h2>Import & Index</h2>
            <p>Importing means the photos you upload are renamed (the naming schema you can define in settings), and moved to the originals folder sorted by year and month.
                Additionally duplicates are removed and images get tagged and metadata (like location, camera model etc.) will be extracted. In case you have not supported file types
                (e.g. videos) within the folder you import --> those are ignored. </p>
            <v-btn color="success" @click="$refs.inputUpload.click()" type="file" class="importbtn" disabled>Import & Index</v-btn>
            <input v-show="false" ref="inputUpload" type="file" @change="inputFileExcel" >
                <v-flex xs12 sm6 offset-sm3>
                    <v-card class="card">
                        <v-card-title primary-title>
                            <div>
                                <div>598 JPEG and 432 RAW files found</div>
                            </div>
                        </v-card-title>

                        <v-card-actions>
                            <v-btn flat color="success">Start</v-btn>
                            <v-btn flat color="success">Cancel</v-btn>
                        </v-card-actions>
                    </v-card>
                </v-flex>
        </v-form>
    </v-container>
        <v-container fluid>
            <v-form>
            <h2>Index</h2>
            <p>In case you already have a nice folder structure you can only index the photos. Therefore in settings you need to set the base directory to the directory your photos
            are in. The index functionality will then just tag the images and extract the metadata.
        </p>
             <v-btn color="success" @click="$refs.inputUpload.click()" type="file" class="importbtn" disabled>Index</v-btn>
          </v-form>
    </v-container>
</div>
</template>

<script>
    import Photo from 'model/photo';

    export default {
        name: 'import',
        props: {},
        data() {
            const query = this.$route.query;
            const order = query['order'] ? query['order'] : 'newest';
            const camera = query['camera'] ? parseInt(query['camera']) : 0;
            const q = query['q'] ? query['q'] : '';
            const country = query['country'] ? query['country'] : '';
            const view = query['view'] ? query['view'] : 'details';
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
                        {value: 'details', text: 'Details'},
                        {value: 'list', text: 'List'},
                        {value: 'tiles', text: 'Tiles'},
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
                    {text: 'Description', value: 'PhotoFavorite'},
                    {text: 'Taken At', value: 'TakenAt'},
                    {text: 'City', value: 'LocCity'},
                    {text: 'Country', value: 'LocCountry'},
                    {text: 'Camera', value: 'CameraModel'},
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

<style scoped>
    h1, h2 {
        font-weight: normal;
    }

    ul {
        list-style-type: none;
        padding: 0;
    }

    li {
        display: inline-block;
        margin: 0 10px;
    }
</style>
