<template>
    <div class="p-page p-page-album-photos" v-infinite-scroll="loadMore" :infinite-scroll-disabled="scrollDisabled"
         :infinite-scroll-distance="10" :infinite-scroll-listen-for-event="'scrollRefresh'">

        <p-album-photo-search :settings="settings" :filter="filter" :filter-change="updateQuery"
                        :refresh="refresh"></p-album-photo-search>

        <v-container fluid class="pa-0">
            <p-scroll-top></p-scroll-top>

            <p-photo-clipboard :refresh="refresh" :selection="selection" :album="model"></p-photo-clipboard>

            <p-album-photo-mosaic v-if="settings.view === 'mosaic'" :photos="results" :selection="selection"
                            :open-photo="openPhoto"></p-album-photo-mosaic>
            <p-album-photo-list v-else-if="settings.view === 'list'" :photos="results" :selection="selection"
                          :open-photo="openPhoto" :open-location="openLocation"></p-album-photo-list>
            <p-album-photo-details v-else-if="settings.view === 'details'" :photos="results" :selection="selection"
                             :open-photo="openPhoto" :open-location="openLocation"></p-album-photo-details>
            <p-album-photo-tiles v-else :photos="results" :selection="selection" :open-photo="openPhoto"></p-album-photo-tiles>
        </v-container>
    </div>
</template>

<script>
    import Photo from "model/photo";
    import Album from "model/album";

    export default {
        name: 'p-page-album-photos',
        props: {
            staticFilter: Object
        },
        watch: {
            '$route'() {
                const query = this.$route.query;

                this.uuid = this.$route.params.uuid;
                this.filter.q = query['q'];
                this.filter.camera = query['camera'] ? parseInt(query['camera']) : 0;
                this.filter.country = query['country'] ? query['country'] : '';
                this.lastFilter = {};
                this.routeName = this.$route.name;
                this.findAlbum();
                this.search();
            }
        },
        data() {
            const uuid = this.$route.params.uuid;
            const query = this.$route.query;
            const routeName = this.$route.name;
            const order = query['order'] ? query['order'] : 'newest';
            const camera = query['camera'] ? parseInt(query['camera']) : 0;
            const q = query['q'] ? query['q'] : '';
            const country = query['country'] ? query['country'] : '';
            const view = this.viewType();
            const filter = {country: country, camera: camera, order: order, q: q};
            const settings = {view: view};

            return {
                model: new Album(),
                uuid: uuid,
                results: [],
                scrollDisabled: true,
                pageSize: 60,
                offset: 0,
                selection: this.$clipboard.selection,
                settings: settings,
                filter: filter,
                lastFilter: {},
                routeName: routeName,
            };
        },
        methods: {
            viewType() {
                let queryParam = this.$route.query['view'];
                let storedType = window.localStorage.getItem("photo_view_type");

                if (queryParam) {
                    window.localStorage.setItem("photo_view_type", queryParam);
                    return queryParam;
                } else if (storedType) {
                    return storedType;
                } else if (window.innerWidth < 960) {
                    return 'mosaic';
                } else if (window.innerWidth > 1600) {
                    return 'details';
                }

                return 'tiles';
            },
            openLocation(index) {
                const photo = this.results[index];

                if (photo.PhotoLat && photo.PhotoLong) {
                    this.$router.push({name: "places", query: {lat: photo.PhotoLat, long: photo.PhotoLong}});
                } else if (photo.LocName) {
                    this.$router.push({name: "places", query: {q: photo.LocName}});
                } else if (photo.LocCity) {
                    this.$router.push({name: "places", query: {q: photo.LocCity}});
                } else if (photo.LocCountry) {
                    this.$router.push({name: "places", query: {q: photo.LocCountry}});
                } else {
                    this.$router.push({name: "places", query: {q: photo.CountryName}});
                }
            },
            openPhoto(index) {
                this.$viewer.show(this.results, index)
            },
            loadMore() {
                if (this.scrollDisabled) return;

                this.scrollDisabled = true;

                this.offset += this.pageSize;

                const params = {
                    count: this.pageSize,
                    offset: this.offset,
                };

                Object.assign(params, this.lastFilter);

                Photo.search(params).then(response => {
                    this.results = this.results.concat(response.models);

                    this.scrollDisabled = (response.models.length < this.pageSize);

                    if (this.scrollDisabled) {
                        this.$notify.info('All ' + this.results.length + ' photos loaded');
                    }
                });
            },
            updateQuery() {
                const query = {
                    view: this.settings.view
                };

                Object.assign(query, this.filter);

                for (let key in query) {
                    if (query[key] === undefined || !query[key]) {
                        delete query[key];
                    }
                }

                this.$router.replace({query: query});
            },
            searchParams() {
                const params = {
                    count: this.pageSize,
                    offset: this.offset,
                    album: this.uuid,
                };

                Object.assign(params, this.filter);

                if (this.staticFilter) {
                    Object.assign(params, this.staticFilter);
                }

                return params;
            },
            refresh() {
                this.lastFilter = {};
                const pageSize = this.pageSize;
                this.pageSize = this.offset + pageSize;
                this.search();
                this.offset = this.pageSize;
                this.pageSize = pageSize;
            },
            search() {
                this.scrollDisabled = true;

                // Don't query the same data more than once
                if (JSON.stringify(this.lastFilter) === JSON.stringify(this.filter)) {
                    this.$nextTick(() => this.$emit("scrollRefresh"));
                    return;
                }

                Object.assign(this.lastFilter, this.filter);

                this.offset = 0;

                const params = this.searchParams();

                Photo.search(params).then(response => {
                    this.results = response.models;

                    this.scrollDisabled = (response.models.length < this.pageSize);

                    if (this.scrollDisabled) {
                        if (!this.results.length) {
                            this.$notify.warning("No photos found");
                        } else if (this.results.length === 1) {
                            this.$notify.info("One photo found");
                        } else {
                            this.$notify.info(this.results.length + " photos found");
                        }
                    } else {
                        this.$notify.info('More than 50 photos found');

                        this.$nextTick(() => this.$emit("scrollRefresh"));
                    }
                });
            },
            findAlbum() {
                this.model.find(this.uuid).then(m => {
                    this.model = m;
                    this.$config.page.title = this.model.AlbumName;
                    window.document.title = this.model.AlbumName;
                });
            },
        },
        created() {
            this.findAlbum();
            this.search();
        },
    };
</script>
