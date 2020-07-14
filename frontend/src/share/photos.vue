<template>
  <div class="p-page p-page-album-photos" v-infinite-scroll="loadMore" :infinite-scroll-disabled="scrollDisabled"
       :infinite-scroll-distance="10" :infinite-scroll-listen-for-event="'scrollRefresh'">

    <v-form lazy-validation dense
            ref="form" autocomplete="off" class="p-photo-toolbar p-album-toolbar" accept-charset="UTF-8">
      <v-toolbar flat color="secondary">
        <v-toolbar-title>
          {{ model.Title }}
        </v-toolbar-title>

        <v-spacer></v-spacer>

        <v-btn icon @click.stop="refresh" class="hidden-xs-only action-reload">
          <v-icon>refresh</v-icon>
        </v-btn>

        <v-btn icon v-if="settings.view === 'cards'" @click.stop="setView('list')">
          <v-icon>view_list</v-icon>
        </v-btn>
        <v-btn icon v-else-if="settings.view === 'list'" @click.stop="setView('mosaic')">
          <v-icon>view_comfy</v-icon>
        </v-btn>
        <v-btn icon v-else @click.stop="setView('cards')">
          <v-icon>view_column</v-icon>
        </v-btn>
      </v-toolbar>
      <template v-if="model.Description">
        <v-card flat class="px-2 py-1 hidden-sm-and-down"
                color="secondary-light"
        >
          <v-card-text>
            {{ model.Description }}
          </v-card-text>
        </v-card>
        <v-card flat class="pa-0 hidden-md-and-up"
                color="secondary-light"
        >
          <v-card-text>
            {{ model.Description }}
          </v-card-text>
        </v-card>
      </template>
    </v-form>

    <v-container fluid class="pa-4" v-if="loading">
      <v-progress-linear color="secondary-dark" :indeterminate="true"></v-progress-linear>
    </v-container>
    <v-container fluid class="pa-0" v-else>
      <p-scroll-top></p-scroll-top>

      <p-photo-clipboard :refresh="refresh"
                         :selection="selection"
                         :album="model" context="album"></p-photo-clipboard>

      <p-photo-mosaic v-if="settings.view === 'mosaic'"
                      :photos="results"
                      :selection="selection"
                      :filter="filter"
                      :album="model"
                      :edit-photo="editPhoto"
                      :open-photo="openPhoto"></p-photo-mosaic>
      <p-photo-list v-else-if="settings.view === 'list'"
                    :photos="results"
                    :selection="selection"
                    :filter="filter"
                    :album="model"
                    :open-photo="openPhoto"
                    :edit-photo="editPhoto"
                    :open-location="openLocation"></p-photo-list>
      <p-photo-cards v-else
                     :photos="results"
                     :selection="selection"
                     :filter="filter"
                     :album="model"
                     :open-photo="openPhoto"
                     :edit-photo="editPhoto"
                     :open-location="openLocation"></p-photo-cards>
    </v-container>
  </div>
</template>

<script>
    import Photo from "model/photo";
    import Album from "model/album";
    import Event from "pubsub-js";
    import Thumb from "model/thumb";

    export default {
        name: 'p-page-album-photos',
        props: {
            staticFilter: Object
        },
        watch: {
            '$route'() {
                const query = this.$route.query;

                this.filter.q = query['q'] ? query['q'] : '';
                this.filter.camera = query['camera'] ? parseInt(query['camera']) : 0;
                this.filter.country = query['country'] ? query['country'] : '';
                this.settings.view = this.viewType();
                this.lastFilter = {};
                this.routeName = this.$route.name;

                if (this.uid !== this.$route.params.uid) {
                    this.uid = this.$route.params.uid;
                    this.findAlbum().then(() => this.search());
                } else {
                    this.search();
                }
            }
        },
        data() {
            const uid = this.$route.params.uid;
            const query = this.$route.query;
            const routeName = this.$route.name;
            const order = query['order'] ? query['order'] : 'oldest';
            const camera = query['camera'] ? parseInt(query['camera']) : 0;
            const q = query['q'] ? query['q'] : '';
            const country = query['country'] ? query['country'] : '';
            const view = this.viewType();
            const filter = {country: country, camera: camera, order: order, q: q};
            const settings = {view: view};

            return {
                subscriptions: [],
                listen: false,
                dirty: false,
                complete: false,
                model: new Album(),
                uid: uid,
                results: [],
                scrollDisabled: true,
                pageSize: 60,
                offset: 0,
                page: 0,
                selection: this.$clipboard.selection,
                settings: settings,
                filter: filter,
                lastFilter: {},
                routeName: routeName,
                loading: true,
                token: this.$route.params.token,
                viewer: {
                    results: [],
                    loading: false,
                },
            };
        },
        methods: {
            setView(name) {
                this.settings.view = name;
                this.updateQuery();
            },
            viewType() {
                let queryParam = this.$route.query['view'];
                let defaultType = window.localStorage.getItem("photo_view_type");
                let storedType = window.localStorage.getItem("album_view_type");

                if (queryParam) {
                    window.localStorage.setItem("album_view_type", queryParam);
                    return queryParam;
                } else if (storedType) {
                    return storedType;
                } else if (defaultType) {
                    return defaultType;
                } else if (window.innerWidth < 960) {
                    return 'mosaic';
                }

                return 'cards';
            },
            openLocation(index) {
                const photo = this.results[index];

                if (photo.CellID && photo.CellID !== "zz") {
                    this.$router.push({name: "place", params: {q: photo.CellID}});
                } else if (photo.PlaceID && photo.PlaceID !== "zz") {
                    this.$router.push({name: "place", params: {q: photo.PlaceID}});
                } else if (photo.Country && photo.Country !== "zz") {
                    this.$router.push({name: "place", params: {q: "country:" + photo.Country}});
                } else {
                    this.$notify.warn("unknown location");
                }
            },
            editPhoto(index) {
                let selection = this.results.map((p) => {
                    return p.getId()
                });

                // Open Edit Dialog
                Event.publish("dialog.edit", {selection: selection, album: this.album, index: index});
            },
            openPhoto(index, showMerged) {
                if (this.loading || this.viewer.loading || !this.results[index]) {
                    return false;
                }

                const selected = this.results[index];

                if (showMerged && (selected.Type === 'video' || selected.Type === 'live')) {
                    if (this.results[index].isPlayable()) {
                        this.$modal.show('video', {video: selected, album: this.album});
                    } else {
                        this.$viewer.show(Thumb.fromPhotos(this.results), index);
                    }
                } else if (showMerged) {
                    this.$viewer.show(Thumb.fromFiles([selected]), 0)
                } else {
                    this.viewerResults().then((results) => {
                        const thumbsIndex = results.findIndex(result => result.UID === selected.UID);

                        if(thumbsIndex < 0) {
                            this.$viewer.show(Thumb.fromPhotos(this.results), index);
                        } else {
                            this.$viewer.show(Thumb.fromPhotos(results), thumbsIndex);
                        }
                    });
                }

                return true;
            },
            viewerResults() {
                if (this.complete || this.loading || this.viewer.loading) {
                    return Promise.resolve(this.results);
                }

                if (this.viewer.results.length >= this.results.length) {
                    return Promise.resolve(this.viewer.results);
                }

                this.viewer.loading = true;

                const count = Photo.limit();
                const offset = 0;

                const params = {
                    count: count,
                    offset: offset,
                    album: this.uid,
                    filter: this.model.Filter ? this.model.Filter : "",
                    merged: true,
                };

                Object.assign(params, this.lastFilter);

                if (this.staticFilter) {
                    Object.assign(params, this.staticFilter);
                }

                return Photo.search(params).then(resp => {
                    // Success.
                    this.viewer.loading = false;
                    this.viewer.results = resp.models;
                    return Promise.resolve(this.viewer.results);
                }, () => {
                    // Error.
                    this.viewer.loading = false;
                    return Promise.resolve(this.results);
                });
            },
            loadMore() {
                if (this.scrollDisabled) return;

                this.scrollDisabled = true;
                this.listen = false;

                const count = this.dirty ? (this.page + 2) * this.pageSize : this.pageSize;
                const offset = this.dirty ? 0 : this.offset;

                const params = {
                    count: count,
                    offset: offset,
                    album: this.uid,
                    filter: this.model.Filter ? this.model.Filter : "",
                    merged: true,
                };

                Object.assign(params, this.lastFilter);

                if (this.staticFilter) {
                    Object.assign(params, this.staticFilter);
                }

                Photo.search(params).then(response => {
                    this.results = Photo.mergeResponse(this.results, response);
                    this.complete = (response.count < count);
                    this.scrollDisabled = this.complete;

                    if (this.complete) {
                        this.offset = offset;

                        if (this.results.length > 1) {
                            this.$notify.info(this.$gettextInterpolate(this.$gettext("All %{n} entries loaded"), {n: this.results.length}));
                        }
                    } else if (this.results.length >= Photo.limit()) {
                        this.offset = offset;
                        this.scrollDisabled = true;
                        this.complete = true;
                        this.$notify.warn(this.$gettext("Can't load more, limit reached"));
                    } else {
                        this.offset = offset + count;
                        this.page++;
                    }
                }).catch(() => {
                    this.scrollDisabled = false;
                }).finally(() => {
                    this.dirty = false;
                    this.loading = false;
                    this.listen = true;

                    if(offset === 0) {
                        this.viewerResults();
                    }
                });
            },
            updateQuery() {
                const len = this.filter.q.length;

                if (len > 1 && len < 3) {
                    this.$notify.error(this.$gettext("Search term too short"));
                    return;
                }

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
                    album: this.uid,
                    filter: this.model.Filter ? this.model.Filter : "",
                    merged: true,
                };

                Object.assign(params, this.filter);

                if (this.staticFilter) {
                    Object.assign(params, this.staticFilter);
                }

                return params;
            },
            refresh() {
                if (this.loading) {
                    return;
                }

                this.loading = true;
                this.page = 0;
                this.dirty = true;
                this.complete = false;
                this.scrollDisabled = false;

                this.loadMore();
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
                this.page = 0;
                this.loading = true;
                this.listen = false;
                this.complete = false;

                const params = this.searchParams();

                Photo.search(params).then(response => {
                    this.offset = this.pageSize;
                    this.results = response.models;
                    this.complete = (response.count < this.pageSize);
                    this.scrollDisabled = this.complete;

                    if (this.complete) {
                        if (!this.results.length) {
                            this.$notify.warn(this.$gettext("No entries found"));
                        } else if (this.results.length === 1) {
                            this.$notify.info(this.$gettext("One entry found"));
                        } else {
                            this.$notify.info(this.$gettextInterpolate(this.$gettext("%{n} entries found"), {n: this.results.length}));
                        }
                    } else {
                        this.$notify.info(this.$gettext('More than 50 entries found'));

                        this.$nextTick(() => this.$emit("scrollRefresh"));
                    }
                }).finally(() => {
                    this.dirty = false;
                    this.loading = false;
                    this.listen = true;

                    this.viewerResults();
                });
            },
            findAlbum() {
                return this.model.find(this.uid).then(m => {
                    this.model = m;

                    this.filter.order = m.Order;
                    window.document.title = this.model.Title;

                    return Promise.resolve(this.model)
                });
            },
            onAlbumsUpdated(ev, data) {
                if (!this.listen) return;

                if (!data || !data.entities) {
                    return
                }

                for (let i = 0; i < data.entities.length; i++) {
                    if (this.model.UID === data.entities[i].UID) {
                        let values = data.entities[i];

                        for (let key in values) {
                            if (values.hasOwnProperty(key)) {
                                this.model[key] = values[key];
                            }
                        }

                        window.document.title = `${this.$config.get("siteTitle")}: ${this.model.Title}`

                        this.dirty = true;
                        this.complete = false;
                        this.scrollDisabled = false;

                        if (this.filter.order !== this.model.Order) {
                            this.filter.order = this.model.Order;
                            this.updateQuery();
                        } else {
                            this.loadMore();
                        }

                        return;
                    }
                }
            },
            updateResult(results, values) {
                const model = results.find((m) => m.UID === values.UID);

                if (model) {
                    for (let key in values) {
                        if (values.hasOwnProperty(key) && values[key] != null && typeof values[key] !== "object") {
                            model[key] = values[key];
                        }
                    }
                }
            },
            removeResult(results, uid) {
                const index = results.findIndex((m) => m.UID === uid);

                if (index >= 0) {
                    results.splice(index, 1);
                }
            },
            onUpdate(ev, data) {
                if (!this.listen) return;

                if (!data || !data.entities) {
                    return
                }

                const type = ev.split('.')[1];

                switch (type) {
                    case 'updated':
                        for (let i = 0; i < data.entities.length; i++) {
                            const values = data.entities[i];
                            this.updateResult(this.results, values);
                            this.updateResult(this.viewer.results, values);
                        }
                        break;
                    case 'restored':
                        this.dirty = true;
                        this.scrollDisabled = false;
                        this.complete = false;

                        this.loadMore();

                        break;
                    case 'archived':
                        this.dirty = true;
                        this.complete = false;

                        for (let i = 0; i < data.entities.length; i++) {
                            const uid = data.entities[i];

                            this.removeResult(this.results, uid);
                            this.removeResult(this.viewer.results, uid);
                            this.$clipboard.removeId(uid);
                        }

                        break;
                }
            },
        },
        created() {
            const token = this.$route.params.token;

            if (this.$session.hasToken(token)) {
                this.findAlbum().then(() => this.search());
            } else {
                this.$session.redeemToken(token).then(() => {
                    this.findAlbum().then(() => this.search());
                });
            }

            this.subscriptions.push(Event.subscribe("albums.updated", (ev, data) => this.onAlbumsUpdated(ev, data)));
            this.subscriptions.push(Event.subscribe("photos", (ev, data) => this.onUpdate(ev, data)));

            this.subscriptions.push(Event.subscribe("touchmove.top", () => this.refresh()));
            this.subscriptions.push(Event.subscribe("touchmove.bottom", () => this.loadMore()));
        },
        destroyed() {
            for (let i = 0; i < this.subscriptions.length; i++) {
                Event.unsubscribe(this.subscriptions[i]);
            }
        },
    };
</script>
