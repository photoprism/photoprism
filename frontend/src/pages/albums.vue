<template>
    <div class="p-page p-page-albums" v-infinite-scroll="loadMore" :infinite-scroll-disabled="scrollDisabled"
         :infinite-scroll-distance="10" :infinite-scroll-listen-for-event="'scrollRefresh'">

        <v-form ref="form" class="p-albums-search" lazy-validation @submit.prevent="updateQuery" dense>
            <v-toolbar flat color="secondary">
                <v-text-field class="pt-3 pr-3"
                              single-line
                              :label="labels.search"
                              browser-autocomplete="off"
                              prepend-inner-icon="search"
                              clearable
                              color="secondary-dark"
                              @click:clear="clearQuery"
                              v-model="filter.q"
                              @keyup.enter.native="updateQuery"
                              id="search"
                ></v-text-field>

                <v-spacer></v-spacer>

                <v-btn icon @click.stop="refresh" :class="dirty ? 'secondary-light': ''">
                    <v-icon>refresh</v-icon>
                </v-btn>

                <v-btn icon @click.prevent="create">
                    <v-icon>add</v-icon>
                </v-btn>
            </v-toolbar>
        </v-form>

        <v-container fluid class="pa-4" v-if="loading">
            <v-progress-linear color="secondary-dark" :indeterminate="true"></v-progress-linear>
        </v-container>
        <v-container fluid class="pa-0" v-else>
            <p-scroll-top></p-scroll-top>

            <p-album-clipboard :refresh="refresh" :selection="selection"></p-album-clipboard>

            <v-container grid-list-xs fluid class="pa-2 p-albums p-albums-details">
                <v-card v-if="results.length === 0" class="p-albums-empty secondary-light lighten-1" flat>
                    <v-card-title primary-title>
                        <div>
                            <h3 class="title mb-3">
                                <translate>No albums matched your search</translate>
                            </h3>
                            <div>
                                <translate>Try again using a different term or create a new album</translate>
                            </div>
                        </div>
                    </v-card-title>
                </v-card>
                <v-layout row wrap class="p-results">
                    <v-flex
                            v-for="(album, index) in results"
                            :key="index"
                            class="p-album"
                            xs6 sm4 md3 lg2 d-flex
                    >
                        <v-hover>
                            <v-card tile class="accent lighten-3"
                                    slot-scope="{ hover }"
                                    :class="selection.includes(album.AlbumUUID) ? 'elevation-10 ma-0' : 'elevation-0 ma-1'"
                                    :to="{name: 'album', params: {uuid: album.AlbumUUID, slug: album.AlbumSlug}}"
                            >
                                <v-img
                                        :src="album.getThumbnailUrl('tile_500')"
                                        aspect-ratio="1"
                                        class="accent lighten-2"
                                >
                                    <v-layout
                                            slot="placeholder"
                                            fill-height
                                            align-center
                                            justify-center
                                            ma-0
                                    >
                                        <v-progress-circular indeterminate
                                                             color="accent lighten-5"></v-progress-circular>
                                    </v-layout>

                                    <v-btn v-if="hover || selection.length > 0" :flat="!hover" :ripple="false"
                                           icon small absolute
                                           class="p-album-select"
                                           @click.stop.prevent="toggleSelection(album.AlbumUUID)">
                                        <v-icon v-if="selection.includes(album.AlbumUUID)" color="white">check_circle
                                        </v-icon>
                                        <v-icon v-else color="accent lighten-3">radio_button_off</v-icon>
                                    </v-btn>
                                </v-img>

                                <v-card-actions @click.stop.prevent="">
                                    <v-edit-dialog
                                            :return-value.sync="album.AlbumName"
                                            lazy
                                            @save="onSave(album)"
                                            class="p-inline-edit"
                                    >
                                        <span v-if="album.AlbumName">
                                            {{ album.AlbumName }}
                                        </span>
                                        <span v-else>
                                            <v-icon>edit</v-icon>
                                        </span>
                                        <template v-slot:input>
                                            <v-text-field
                                                    v-model="album.AlbumName"
                                                    :rules="[titleRule]"
                                                    :label="labels.name"
                                                    color="secondary-dark"
                                                    single-line
                                                    autofocus
                                            ></v-text-field>
                                        </template>
                                    </v-edit-dialog>

                                    <v-spacer></v-spacer>
                                    <v-btn icon @click.stop.prevent="album.toggleLike()">
                                        <v-icon v-if="album.AlbumFavorite" color="#FFD600">star
                                        </v-icon>
                                        <v-icon v-else color="accent lighten-2">star</v-icon>
                                    </v-btn>
                                </v-card-actions>
                            </v-card>
                        </v-hover>
                    </v-flex>
                </v-layout>
            </v-container>
        </v-container>
    </div>
</template>

<script>
    import Album from "model/album";
    import {DateTime} from "luxon";
    import Event from "pubsub-js";

    export default {
        name: 'p-page-albums',
        props: {
            staticFilter: Object
        },
        watch: {
            '$route'() {
                const query = this.$route.query;

                this.filter.q = query['q'] ? query['q'] : '';
                this.lastFilter = {};
                this.routeName = this.$route.name;
                this.search();
            }
        },
        data() {
            const query = this.$route.query;
            const routeName = this.$route.name;
            const q = query['q'] ? query['q'] : '';
            const filter = {q: q};
            const settings = {};

            return {
                subId: null,
                dirty: false,
                results: [],
                loading: true,
                scrollDisabled: true,
                pageSize: 24,
                offset: 0,
                selection: [],
                settings: settings,
                filter: filter,
                lastFilter: {},
                routeName: routeName,
                titleRule: v => v.length <= 25 || this.$gettext("Title too long"),
                labels: {
                    search: this.$gettext("Search"),
                    name: this.$gettext("Album Name"),
                },
            };
        },
        methods: {
            clearQuery() {
                this.filter.q = '';
                this.search();
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

                Album.search(params).then(response => {
                    this.results = this.results.concat(response.models);

                    this.scrollDisabled = (response.models.length < this.pageSize);

                    if (this.scrollDisabled) {
                        this.$notify.info(this.$gettext("All ") + this.results.length + this.$gettext(" albums loaded"));
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

                Album.search(params).then(response => {
                    this.loading = false;
                    this.dirty = false;
                    this.results = response.models;

                    this.scrollDisabled = (response.models.length < this.pageSize);

                    if (this.scrollDisabled) {
                        if (!this.results.length) {
                            this.$notify.warning(this.$gettext("No albums found"));
                        } else if (this.results.length === 1) {
                            this.$notify.info(this.$gettext("One album found"));
                        } else {
                            this.$notify.info(this.results.length + this.$gettext(" albums found"));
                        }
                    } else {
                        this.$notify.info(this.$gettext('More than 20 albums found'));

                        this.$nextTick(() => this.$emit("scrollRefresh"));
                    }
                }).catch(() => this.loading = false);
            },
            refresh() {
                this.lastFilter = {};
                const pageSize = this.pageSize;
                this.pageSize = this.offset + pageSize;
                this.search();
                this.offset = this.pageSize;
                this.pageSize = pageSize;
            },
            create() {
                let name = DateTime.local().toFormat("LLLL yyyy");

                if (this.results.findIndex(a => a.AlbumName.startsWith(name)) !== -1) {
                    const existing = this.results.filter(a => a.AlbumName.startsWith(name));
                    name = `${name} (${existing.length + 1})`
                }

                const album = new Album({"AlbumName": name, "AlbumFavorite": true});

                album.save().then(() => {
                    this.filter.q = "";
                    this.lastFilter = {};

                    this.search();
                })
            },
            onSave(album) {
                album.update();
            },
            toggleSelection(uuid) {
                const pos = this.selection.indexOf(uuid);

                if (pos !== -1) {
                    this.selection.splice(pos, 1);
                } else {
                    this.selection.push(uuid)
                }
            },
            onCount() {
                this.dirty = true;

                if(!this.selection && this.offset === 0) {
                    this.refresh();
                }
            }
        },
        created() {
            this.search();
            this.subId = Event.subscribe("count.albums", (ev, data) => this.onCount(ev, data));
        },
        destroyed() {
            Event.unsubscribe(this.subId);
        },
    };
</script>
