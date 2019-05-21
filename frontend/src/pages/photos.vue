<template>
    <div class="p-page p-page-photos" v-infinite-scroll="loadMore" infinite-scroll-disabled="loadMoreDisabled"
         infinite-scroll-distance="10" infinite-scroll-listen-for-event="infiniteScrollRefresh">

        <p-photo-search :settings="settings" :filter="filter" :filter-change="refreshList" :settings-change="updateQuery"></p-photo-search>

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

            <p-photo-mosaic v-if="settings.view === 'mosaic'" :photos="results" :selection="selected" :select="selectPhoto"
                            :open="openPhoto" :like="likePhoto"></p-photo-mosaic>
            <p-photo-list v-else-if="settings.view === 'list'" :photos="results" :selection="selected" :select="selectPhoto"
                          :open="openPhoto" :like="likePhoto"></p-photo-list>
            <p-photo-details v-else-if="settings.view === 'details'" :photos="results" :selection="selected"
                             :select="selectPhoto" :open="openPhoto" :like="likePhoto"></p-photo-details>
            <p-photo-tiles v-else :photos="results" :selection="selected" :select="selectPhoto"
                           :open="openPhoto" :like="likePhoto"></p-photo-tiles>
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

            return {
                'searchExpanded': false,
                'loadMoreDisabled': true,
                'menuVisible': false,
                'results': [],
                'pageSize': 60,
                'offset': 0,
                'selected': this.$clipboard.selection,
                'settings': {
                    'view': view,
                },
                'filter': {
                    country: country,
                    camera: camera,
                    order: order,
                    q: q,
                },
                'lastFilter': {},
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
                    view: this.settings.view
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
