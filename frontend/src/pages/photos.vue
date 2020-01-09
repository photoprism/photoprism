<template>
    <div class="p-page p-page-photos" v-infinite-scroll="loadMore" :infinite-scroll-disabled="scrollDisabled"
         :infinite-scroll-distance="10" :infinite-scroll-listen-for-event="'scrollRefresh'">

        <p-photo-search :settings="settings" :filter="filter" :filter-change="updateQuery"
                        :refresh="refresh"></p-photo-search>

        <v-container fluid class="pa-4" v-if="loading">
            <v-progress-linear color="secondary-dark"  :indeterminate="true"></v-progress-linear>
        </v-container>
        <v-container fluid class="pa-0" v-else>
            <p-scroll-top></p-scroll-top>

            <p-photo-clipboard :refresh="refresh" :selection="selection" :context="context"></p-photo-clipboard>

            <p-photo-mosaic v-if="settings.view === 'mosaic'"
                            :photos="results"
                            :selection="selection"
                            :open-photo="openPhoto"></p-photo-mosaic>
            <p-photo-list v-else-if="settings.view === 'list'"
                          :photos="results"
                          :selection="selection"
                          :open-photo="openPhoto"
                          :open-location="openLocation"></p-photo-list>
            <p-photo-details v-else
                             :photos="results"
                             :selection="selection"
                             :open-photo="openPhoto"
                             :open-location="openLocation"></p-photo-details>
        </v-container>
    </div>
</template>

<script>
    import Photo from "model/photo";

    export default {
        name: 'p-page-photos',
        props: {
            staticFilter: Object
        },
        watch: {
            '$route'() {
                const query = this.$route.query;

                this.filter.q = query['q'] ? query['q'] : '';
                this.filter.camera = query['camera'] ? parseInt(query['camera']) : 0;
                this.filter.country = query['country'] ? query['country'] : '';
                this.filter.lens = query['lens'] ? parseInt(query['lens']) : 0;
                this.filter.year = query['year'] ? parseInt(query['year']) : 0;
                this.filter.color = query['color'] ? query['color'] : '';
                this.filter.label = query['label'] ? query['label'] : '';
                this.settings.view = this.viewType();
                this.lastFilter = {};
                this.routeName = this.$route.name;
                this.search();
            }
        },
        data() {
            const query = this.$route.query;
            const routeName = this.$route.name;
            const order = query['order'] ? query['order'] : 'newest';
            const camera = query['camera'] ? parseInt(query['camera']) : 0;
            const q = query['q'] ? query['q'] : '';
            const country = query['country'] ? query['country'] : '';
            const lens = query['lens'] ? parseInt(query['lens']) : 0;
            const year = query['year'] ? parseInt(query['year']) : 0;
            const color = query['color'] ? query['color'] : '';
            const label = query['label'] ? query['label'] : '';
            const view = this.viewType();
            const filter = {
                country: country,
                camera: camera,
                lens: lens,
                label: label,
                year: year,
                color: color,
                order: order,
                q: q,
                /* before: before,
                after: after, */
            };
            const settings = {view: view};

            return {
                results: [],
                scrollDisabled: true,
                pageSize: 60,
                offset: 0,
                selection: this.$clipboard.selection,
                settings: settings,
                filter: filter,
                lastFilter: {},
                routeName: routeName,
                loading: true
            };
        },
        computed: {
            context: function () {
                if(!this.staticFilter) {
                    return "photos"
                }

                if(this.staticFilter.archived) {
                    return "archive"
                } else if (this.staticFilter.favorites) {
                    return "favorites"
                }

                return ""
            }
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
                }

                return 'details';
            },
            openLocation(index) {
                const photo = this.results[index];

                if (photo.PhotoLat && photo.PhotoLng) {
                    this.$router.push({name: "places", query: {lat: String(photo.PhotoLat), lng: String(photo.PhotoLng)}});
                } else if (photo.LocCity) {
                    this.$router.push({name: "places", query: {q: photo.LocCity}});
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
                        this.$notify.info(this.$gettext('All ') + this.results.length + this.$gettext(' photos loaded'));
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

                if (JSON.stringify(this.$route.query) === JSON.stringify(query)) {
                    return
                }

                this.$router.replace({query: query});
            },
            searchParams() {
                const params = {
                    count: this.pageSize,
                    offset: this.offset,
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
                this.loading = true;

                const params = this.searchParams();

                Photo.search(params).then(response => {
                    this.loading = false;
                    this.results = response.models;

                    this.scrollDisabled = (response.models.length < this.pageSize);

                    if (this.scrollDisabled) {
                        if (!this.results.length) {
                            this.$notify.warning(this.$gettext("No photos found"));
                        } else if (this.results.length === 1) {
                            this.$notify.info(this.$gettext("One photo found"));
                        } else {
                            this.$notify.info(this.results.length + this.$gettext(" photos found"));
                        }
                    } else {
                        this.$notify.info(this.$gettext('More than 50 photos found'));

                        this.$nextTick(() => this.$emit("scrollRefresh"));
                    }
                }).catch(() => this.loading = false);
            },
        },
        created() {
            this.search();
        },
        destroyed() {
        }
    };
</script>
