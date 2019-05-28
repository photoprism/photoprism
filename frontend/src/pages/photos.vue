<template>
    <div class="p-page p-page-photos" v-infinite-scroll="loadMore" :infinite-scroll-disabled="scrollDisabled"
         :infinite-scroll-distance="10" :infinite-scroll-listen-for-event="'scrollRefresh'">

        <p-photo-search :settings="settings" :filter="filter" :filter-change="search"
                        :settings-change="updateQuery"></p-photo-search>

        <v-container fluid>
            <p-photo-clipboard :selection="selection"></p-photo-clipboard>

            <p-photo-mosaic v-if="settings.view === 'mosaic'" :photos="results" :selection="selection"
                            :open-photo="openPhoto"></p-photo-mosaic>
            <p-photo-list v-else-if="settings.view === 'list'" :photos="results" :selection="selection"
                          :open-photo="openPhoto" :open-location="openLocation"></p-photo-list>
            <p-photo-details v-else-if="settings.view === 'details'" :photos="results" :selection="selection"
                             :open-photo="openPhoto" :open-location="openLocation"></p-photo-details>
            <p-photo-tiles v-else :photos="results" :selection="selection" :open-photo="openPhoto"></p-photo-tiles>
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
            '$route' () {
                this.lastFilter = {};
                this.search();
            }
        },
        data() {
            const query = this.$route.query;
            const order = query['order'] ? query['order'] : 'newest';
            const camera = query['camera'] ? parseInt(query['camera']) : 0;
            const q = query['q'] ? query['q'] : '';
            const country = query['country'] ? query['country'] : '';
            const view = query['view'] ? query['view'] : 'tiles';
            const filter = {country: country, camera: camera, order: order, q: q};
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
            };
        },
        methods: {
            openLocation(index) {
                const photo = this.results[index];

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
                        this.$alert.info('All ' + this.results.length + ' photos loaded');
                    }
                });
            },
            updateQuery() {
                const query = {
                    view: this.settings.view
                };

                Object.assign(query, this.filter);

                this.$router.replace({query: query});

                this.$nextTick(() => this.$emit("scrollRefresh"));
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
            search() {
                this.scrollDisabled = true;

                // Don't query the same data more than once
                if (JSON.stringify(this.lastFilter) === JSON.stringify(this.filter)) return;

                Object.assign(this.lastFilter, this.filter);

                this.offset = 0;

                this.updateQuery();

                const params = this.searchParams();

                Photo.search(params).then(response => {
                    this.results = response.models;

                    this.scrollDisabled = (response.models.length < this.pageSize);

                    if (this.scrollDisabled) {
                        this.$alert.info(this.results.length + ' photos found');
                    } else {
                        this.$alert.info('More than 50 photos found');
                    }
                });
            },
        },
        created() {
            this.search();
        },
    };
</script>
