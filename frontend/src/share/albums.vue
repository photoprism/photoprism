<template>
  <div class="p-page p-page-albums" v-infinite-scroll="loadMore" :infinite-scroll-disabled="scrollDisabled"
       :infinite-scroll-distance="10" :infinite-scroll-listen-for-event="'scrollRefresh'">

    <v-container fluid class="pa-4" v-if="loading">
      <v-progress-linear color="secondary-dark" :indeterminate="true"></v-progress-linear>
    </v-container>
    <v-container fluid class="pa-0" v-else>
      <p-scroll-top></p-scroll-top>

      <p-album-clipboard :refresh="refresh" :selection="selection"
                         :clear-selection="clearSelection"></p-album-clipboard>

      <v-container grid-list-xs fluid class="pa-2 p-albums p-albums-cards">
        <v-card v-if="results.length === 0" class="p-albums-empty secondary-light lighten-1 ma-1" flat>
          <v-card-title primary-title>
            <div>
              <h3 class="title ma-0 pa-0">
                <translate>Couldn't find anything</translate>
              </h3>
              <p class="mt-4 mb-0 pa-0">
                <translate>Try again using other filters or keywords.</translate>
              </p>
            </div>
          </v-card-title>
        </v-card>
        <v-layout row wrap class="p-album-results">
          <v-flex
                  v-for="(album, index) in results"
                  :key="index"
                  :data-uid="album.UID"
                  class="p-album"
                  xs6 sm4 lg3 xl2 d-flex
          >
            <v-hover>
              <v-card tile class="accent lighten-3"
                      slot-scope="{ hover }"
                      @contextmenu="onContextMenu($event, index)"
                      :dark="selection.includes(album.UID)"
                      :class="selection.includes(album.UID) ? 'elevation-10 ma-0 accent darken-1 white--text' : 'elevation-0 ma-1 accent lighten-3'"
                      :to="{name: view, params: {uid: album.UID, slug: album.Slug, year: album.Year, month: album.Month}}"
              >
                <v-img
                        :src="album.thumbnailUrl('tile_500')"
                        @mousedown="onMouseDown($event, index)"
                        @click="onClick($event, index)"
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

                  <v-btn v-if="hover || selection.includes(album.UID)" :flat="!hover" :ripple="false"
                         icon large absolute
                         :class="selection.includes(album.UID) ? 'action-select' : 'action-select opacity-50'"
                         @click.stop.prevent="onSelect($event, index)">
                    <v-icon v-if="selection.includes(album.UID)" color="white"
                            class="t-select t-on">check_circle
                    </v-icon>
                    <v-icon v-else color="accent lighten-3" class="t-select t-off">
                      radio_button_off
                    </v-icon>
                  </v-btn>
                </v-img>

                <v-card-title primary-title class="pa-3 p-album-desc" style="user-select: none;">
                  <h3 class="body-2 ma-0">
                    {{ album.Title }}
                  </h3>
                </v-card-title>
                <v-card-text class="pl-3 pr-3 pt-0 pb-3 p-album-desc">
                  <div class="caption" title="Description" v-if="album.Description">
                    {{ album.Description | truncate(100) }}
                  </div>
                  <div class="caption" title="Description" v-else>
                    <translate>Shared with you.</translate>
                  </div>
                </v-card-text>
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
    import RestModel from "model/rest";

    export default {
        name: 'p-page-albums',
        props: {
            staticFilter: Object,
            view: String,
        },
        watch: {
            '$route'() {
                const query = this.$route.query;

                this.filter.q = query["q"] ? query["q"] : "";
                this.filter.category = query["category"] ? query["category"] : "";
                this.lastFilter = {};
                this.routeName = this.$route.name;
                this.search();
            }
        },
        data() {
            const query = this.$route.query;
            const routeName = this.$route.name;
            const q = query["q"] ? query["q"] : "";
            const category = query["category"] ? query["category"] : "";
            const filter = {q, category};
            const settings = {};

            let categories = [{"value": "", "text": this.$gettext("All Categories")}];

            if (this.$config.values.albumCategories) {
                categories = categories.concat(this.$config.values.albumCategories.map(cat => {
                    return {"value": cat, "text": cat};
                }));
            }

            return {
                categories: categories,
                subscriptions: [],
                listen: false,
                dirty: false,
                results: [],
                loading: true,
                scrollDisabled: true,
                pageSize: 24,
                offset: 0,
                page: 0,
                selection: [],
                settings: settings,
                filter: filter,
                lastFilter: {},
                routeName: routeName,
                titleRule: v => v.length <= this.$config.get('clip') || this.$gettext("Title too long"),
                labels: {
                    search: this.$gettext("Search"),
                    title: this.$gettext("Album Name"),
                    category: this.$gettext("Category"),
                },
                mouseDown: {
                    index: -1,
                    timeStamp: -1,
                },
                lastId: "",
            };
        },
        methods: {
            showUpload() {
                Event.publish("dialog.upload");
            },
            selectRange(rangeEnd, models) {
                if (!models || !models[rangeEnd] || !(models[rangeEnd] instanceof RestModel)) {
                    console.warn("selectRange() - invalid arguments:", rangeEnd, models);
                    return;
                }

                let rangeStart = models.findIndex((m) => m.getId() === this.lastId);

                if (rangeStart === -1) {
                    this.toggleSelection(models[rangeEnd].getId());
                    return 1;
                }

                if (rangeStart > rangeEnd) {
                    const newEnd = rangeStart;
                    rangeStart = rangeEnd;
                    rangeEnd = newEnd;
                }

                for (let i = rangeStart; i <= rangeEnd; i++) {
                    this.addSelection(models[i].getId());
                }

                return (rangeEnd - rangeStart) + 1;
            },
            onSelect(ev, index) {
                if (ev.shiftKey) {
                    this.selectRange(index, this.results);
                } else {
                    this.toggleSelection(this.results[index].getId());
                }
            },
            onMouseDown(ev, index) {
                this.mouseDown.index = index;
                this.mouseDown.timeStamp = ev.timeStamp;
            },
            onClick(ev, index) {
                let longClick = (this.mouseDown.index === index && ev.timeStamp - this.mouseDown.timeStamp > 400);

                if (longClick || this.selection.length > 0) {
                    ev.preventDefault();
                    ev.stopPropagation();

                    if (longClick || ev.shiftKey) {
                        this.selectRange(index, this.results);
                    } else {
                        this.toggleSelection(this.results[index].getId());
                    }
                }
            },
            onContextMenu(ev, index) {
                if (this.$isMobile) {
                    ev.preventDefault();
                    ev.stopPropagation();

                    if (this.results[index]) {
                        this.selectRange(index, this.results);
                    }
                }
            },
            clearQuery() {
                this.filter.q = '';
                this.search();
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
                };

                Object.assign(params, this.lastFilter);

                if (this.staticFilter) {
                    Object.assign(params, this.staticFilter);
                }

                Album.search(params).then(response => {
                    this.results = this.dirty ? response.models : this.results.concat(response.models);

                    this.scrollDisabled = (response.models.length < count);

                    if (this.scrollDisabled) {
                        this.offset = offset;

                        if (this.results.length > 1) {
                            this.$notify.info(this.$gettextInterpolate(this.$gettext("All %{n} albums loaded"), {n: this.results.length}));
                        }
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
                this.page = 0;
                this.loading = true;
                this.listen = false;

                const params = this.searchParams();

                Album.search(params).then(response => {
                    this.offset = this.pageSize;

                    this.results = response.models;

                    this.scrollDisabled = (response.models.length < this.pageSize);

                    if (this.scrollDisabled) {
                        if (!this.results.length) {
                            this.$notify.warn(this.$gettext("No albums found"));
                        } else if (this.results.length === 1) {
                            this.$notify.info(this.$gettext("One album found"));
                        } else {
                            this.$notify.info(this.$gettextInterpolate(this.$gettext("%{n} albums found"), {n: this.results.length}));
                        }
                    } else {
                        this.$notify.info(this.$gettext('More than 20 albums found'));

                        this.$nextTick(() => this.$emit("scrollRefresh"));
                    }
                }).finally(() => {
                    this.dirty = false;
                    this.loading = false;
                    this.listen = true;
                });
            },
            refresh() {
                if (this.loading) return;
                this.loading = true;
                this.page = 0;
                this.dirty = true;
                this.scrollDisabled = false;
                this.loadMore();
            },
            create() {
                let title = DateTime.local().toFormat("LLLL yyyy");

                if (this.results.findIndex(a => a.Title.startsWith(title)) !== -1) {
                    const existing = this.results.filter(a => a.Title.startsWith(title));
                    title = `${title} (${existing.length + 1})`
                }

                const album = new Album({"Title": title, "Favorite": true});

                album.save();
            },
            onSave(album) {
                album.update();
            },
            addSelection(uid) {
                const pos = this.selection.indexOf(uid);

                if (pos === -1) {
                    this.selection.push(uid)
                    this.lastId = uid;
                }
            },
            toggleSelection(uid) {
                const pos = this.selection.indexOf(uid);

                if (pos !== -1) {
                    this.selection.splice(pos, 1);
                    this.lastId = "";
                } else {
                    this.selection.push(uid);
                    this.lastId = uid;
                }
            },
            removeSelection(uid) {
                const pos = this.selection.indexOf(uid);

                if (pos !== -1) {
                    this.selection.splice(pos, 1);
                    this.lastId = "";
                }
            },
            clearSelection() {
                this.selection.splice(0, this.selection.length);
                this.lastId = "";
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
                            const model = this.results.find((m) => m.UID === values.UID);

                            if (model) {
                                for (let key in values) {
                                    if (values.hasOwnProperty(key) && values[key] != null && typeof values[key] !== "object") {
                                        model[key] = values[key];
                                    }
                                }
                            }
                        }
                        break;
                    case 'deleted':
                        this.dirty = true;

                        for (let i = 0; i < data.entities.length; i++) {
                            const uid = data.entities[i];
                            const index = this.results.findIndex((m) => m.UID === uid);

                            if (index >= 0) {
                                this.results.splice(index, 1);
                            }

                            this.removeSelection(uid)
                        }

                        break;
                    case 'created':
                        this.dirty = true;

                        for (let i = 0; i < data.entities.length; i++) {
                            const values = data.entities[i];
                            const index = this.results.findIndex((m) => m.UID === values.UID);

                            if (index === -1 && this.staticFilter.type === values.Type) {
                                this.results.unshift(new Album(values));
                            }
                        }
                        break;
                    default:
                        console.warn("unexpected event type", ev);
                }
            }
        },
        created() {
            const token = this.$route.params.token;

            if (this.$session.hasToken(token)) {
                this.search();
            } else {
                this.$session.redeemToken(token).then(() => {
                    this.search();
                });
            }

            this.subscriptions.push(Event.subscribe("albums", (ev, data) => this.onUpdate(ev, data)));

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
