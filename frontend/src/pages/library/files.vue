<template>
  <div class="p-page p-page-files">
    <v-form ref="form" class="p-files-search" lazy-validation @submit.prevent="updateQuery" dense>
      <v-toolbar flat color="secondary">
        <v-toolbar-title>
          <router-link to="/library/files">
            <translate key="Originals">Originals</translate>
          </router-link>

          <router-link v-for="(item, index) in breadcrumbs" :key="index" :to="item.path">
            <v-icon>navigate_next</v-icon>
            {{item.name}}
          </router-link>
        </v-toolbar-title>

        <v-spacer></v-spacer>

        <v-btn icon @click.stop="refresh">
          <v-icon>refresh</v-icon>
        </v-btn>
      </v-toolbar>
    </v-form>

    <v-container fluid class="pa-4" v-if="loading">
      <v-progress-linear color="secondary-dark" :indeterminate="true"></v-progress-linear>
    </v-container>
    <v-container fluid class="pa-0" v-else>
      <p-file-clipboard :refresh="refresh" :selection="selection"
                        :clear-selection="clearSelection"></p-file-clipboard>

      <p-scroll-top></p-scroll-top>

      <v-container grid-list-xs fluid class="pa-2 p-files p-files-cards">
        <v-card v-if="results.length === 0" class="p-files-empty secondary-light lighten-1 ma-1" flat>
          <v-card-title primary-title>
            <div>
              <h3 class="title ma-0 pa-0">
                <translate>Couldn't find anything</translate>
              </h3>
              <p class="mt-4 mb-0 pa-0">
                <translate>Duplicates will be skipped and only appear once.</translate>
                <translate>If a file you expect is missing, please re-index your library and wait until indexing has been completed.</translate>
              </p>
            </div>
          </v-card-title>
        </v-card>
        <v-layout row wrap class="p-files-results">
          <v-flex
                  v-for="(model, index) in results"
                  :key="index"
                  :data-uid="model.UID"
                  class="p-file"
                  xs6 sm4 md3 lg2 d-flex
          >
            <v-hover>
              <v-card tile class="accent lighten-3 clickable"
                      slot-scope="{ hover }"
                      @contextmenu="onContextMenu($event, index)"
                      :dark="selection.includes(model.UID)"
                      :class="selection.includes(model.UID) ? 'elevation-10 ma-0 darken-1 white--text' : 'elevation-0 ma-1 lighten-3'">
                <v-img
                        :src="model.thumbnailUrl('tile_500')"
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

                  <v-btn v-if="hover || selection.includes(model.UID)" :flat="!hover" :ripple="false"
                         icon large absolute
                         :class="selection.includes(model.UID) ? 'p-file-select' : 'p-file-select opacity-50'"
                         @click.stop.prevent="onSelect($event, index)">
                    <v-icon v-if="selection.includes(model.UID)" color="white" class="t-select t-on">check_circle
                    </v-icon>
                    <v-icon v-else color="accent lighten-3" class="t-select t-off">radio_button_off</v-icon>
                  </v-btn>
                </v-img>

                <v-card-title primary-title class="pa-3 p-photo-desc" style="user-select: none;"
                              v-if="model.isFile()">
                  <div>
                    <h3 class="body-2 mb-2" :title="model.Name">
                      <button @click.exact="openFile(index)">
                        {{ model.baseName() }}
                      </button>
                    </h3>
                    <div class="caption" title="Info">
                      {{ model.getInfo() }}
                    </div>
                  </div>
                </v-card-title>
                <v-card-title primary-title class="pa-3 p-photo-desc" v-else>
                  <div>
                    <h3 class="body-2 mb-2" :title="model.Title">
                      <button @click.exact="openFile(index)">
                        {{ model.baseName() }}
                      </button>
                    </h3>
                    <div class="caption" title="Path">
                      <translate key="Folder">Folder</translate>
                    </div>
                  </div>
                </v-card-title>
              </v-card>
            </v-hover>
          </v-flex>
        </v-layout>
      </v-container>
    </v-container>
  </div>
</template>

<script>
    import Event from "pubsub-js";
    import RestModel from "model/rest";
    import Thumb from "model/thumb";
    import {Folder} from "model/folder";
    import {Photo, TypeJpeg} from "model/photo";

    export default {
        name: 'p-page-files',
        props: {
            staticFilter: Object
        },
        watch: {
            '$route'() {
                const query = this.$route.query;

                this.filter.q = query['q'] ? query['q'] : '';
                this.filter.all = query['all'] ? query['all'] : '';
                this.lastFilter = {};
                this.routeName = this.$route.name;
                this.path = this.$route.params.pathMatch;
                this.search();
            }
        },
        data() {
            const query = this.$route.query;
            const routeName = this.$route.name;
            const q = query['q'] ? query['q'] : '';
            const all = query['all'] ? query['all'] : '';
            const filter = {q: q, all: all};
            const settings = {};

            return {
                config: this.$config.values,
                subscriptions: [],
                listen: false,
                dirty: false,
                results: [],
                loading: true,
                selection: [],
                settings: settings,
                filter: filter,
                lastFilter: {},
                routeName: routeName,
                path: "",
                page: 0,
                files: {
                    limit: Folder.limit(),
                    offset: 0,
                },
                labels: {
                    search: this.$gettext("Search"),
                    name: this.$gettext("Folder Name"),
                },
                titleRule: v => v.length <= this.$config.get('clip') || this.$gettext("Name too long"),
                mouseDown: {
                    index: -1,
                    timeStamp: -1,
                },
                lastId: "",
                breadcrumbs: [],
            };
        },
        methods: {
            getBreadcrumbs() {
                let result = [];
                let path = "/library/files";

                const crumbs = this.path.split("/");

                crumbs.forEach(dir => {
                    if (dir) {
                        path += "/" + dir
                        result.push({path: path, name: dir})
                    }
                })

                return result;
            },
            openFile(index) {
                const model = this.results[index];

                if (model.isFile()) {
                    if (model.Type === TypeJpeg) {
                        const photo = new Photo({
                            UID: model.PhotoUID,
                            Title: model.Name,
                            TakenAt: model.Modified,
                            Description: "",
                            Favorite: false,
                            Files: [model]
                        });
                        this.$viewer.show(Thumb.fromPhotos([photo]), 0);
                    } else {
                        this.downloadFile(index);
                    }
                } else {
                    this.$router.push({path: '/library/files/' + model.Path});
                }
            },
            downloadFile(index) {
                const model = this.results[index];
                const link = document.createElement('a')
                link.href = `/api/v1/dl/${model.Hash}?t=${this.$config.downloadToken()}`;
                link.download = model.Name;
                link.click()
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
                } else {
                    this.openFile(index);
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
            onSave(model) {
                model.update();
            },
            showAll() {
                this.filter.all = "true";
                this.updateQuery();
            },
            showImportant() {
                this.filter.all = "";
                this.updateQuery();
            },
            clearQuery() {
                this.filter.q = '';
                this.updateQuery();
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
                    files: true,
                    uncached: true,
                    count: this.files.limit,
                    offset: this.files.offset,
                };

                Object.assign(params, this.filter);

                if (this.staticFilter) {
                    Object.assign(params, this.staticFilter);
                }

                return params;
            },
            refresh() {
                if (this.loading) return;
                this.loading = true;
                this.page = 0;
                this.dirty = true;
                this.search();
            },
            search() {
                // Don't query the same data more than once
                if (!this.dirty && (JSON.stringify(this.lastFilter) === JSON.stringify(this.filter))) {
                    this.loading = false;
                    this.listen = true;
                    return;
                }

                Object.assign(this.lastFilter, this.filter);

                this.files.offset = 0;
                this.page = 0;
                this.loading = true;
                this.listen = false;

                const params = this.searchParams();

                Folder.originals(this.path, params).then(response => {
                    this.files.offset = this.files.limit;

                    this.results = response.models;
                    this.breadcrumbs = this.getBreadcrumbs();

                    if (response.count === 0) {
                        this.$notify.warn(this.$gettext('Folder is empty'));
                    } else if (response.files === 1) {
                        this.$notify.info(this.$gettext('One file found'));
                    } else if (response.files === 0 && response.folders === 1) {
                        this.$notify.info(this.$gettext('One folder found'));
                    } else if (response.files === 0 && response.folders > 1) {
                        this.$notify.info(this.$gettextInterpolate(this.$gettext("%{n} folders found"), {n: response.folders}));
                    } else if (response.files < this.files.limit) {
                        this.$notify.info(this.$gettextInterpolate(this.$gettext("Folder contains %{n} files"), {n: response.files}));
                    } else {
                        this.$notify.warn(this.$gettextInterpolate(this.$gettext("Limit reached, showing first %{n} files"), {n: response.files}));
                    }
                }).finally(() => {
                    this.dirty = false;
                    this.loading = false;
                    this.listen = true;
                });
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

                            for (let key in values) {
                                if (values.hasOwnProperty(key)) {
                                    model[key] = values[key];
                                }
                            }
                        }
                        break;
                    case 'deleted':
                        this.dirty = true;

                        for (let i = 0; i < data.entities.length; i++) {
                            const ppid = data.entities[i];
                            const index = this.results.findIndex((m) => m.UID === ppid);

                            if (index >= 0) {
                                this.results.splice(index, 1);
                            }

                            this.removeSelection(ppid)
                        }

                        break;
                    case 'created':
                        this.dirty = true;
                        break;
                    default:
                        console.warn("unexpected event type", ev);
                }
            }
        },
        created() {
            this.path = this.$route.params.pathMatch;

            this.search();

            this.subscriptions.push(Event.subscribe("folders", (ev, data) => this.onUpdate(ev, data)));

            this.subscriptions.push(Event.subscribe("touchmove.top", () => this.refresh()));
        },
        destroyed() {
            for (let i = 0; i < this.subscriptions.length; i++) {
                Event.unsubscribe(this.subscriptions[i]);
            }
        },
    };
</script>
