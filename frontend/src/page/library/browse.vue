<template>
  <div class="p-page p-page-files">
    <v-form ref="form" class="p-files-search" lazy-validation dense @submit.prevent="updateQuery">
      <v-toolbar flat color="secondary" :dense="$vuetify.breakpoint.smAndDown">
        <v-toolbar-title>
          <router-link to="/index/files">
            <translate key="Originals">Originals</translate>
          </router-link>

          <router-link v-for="(item, index) in breadcrumbs" :key="index" :to="item.path">
            <v-icon>{{ navIcon }}</v-icon>
            {{ item.name }}
          </router-link>
        </v-toolbar-title>

        <v-spacer></v-spacer>

        <v-btn icon :title="$gettext('Reload')" class="action-reload" @click.stop="refresh">
          <v-icon>refresh</v-icon>
        </v-btn>
      </v-toolbar>
    </v-form>

    <v-container v-if="loading" fluid class="pa-4">
      <v-progress-linear color="secondary-dark" :indeterminate="true"></v-progress-linear>
    </v-container>
    <v-container v-else fluid class="pa-0">
      <p-file-clipboard :refresh="refresh" :selection="selection"
                        :clear-selection="clearSelection"></p-file-clipboard>

      <p-scroll-top></p-scroll-top>

      <v-container grid-list-xs fluid class="pa-2 p-files p-files-cards">
        <v-alert
            :value="results.length === 0"
            color="secondary-dark" icon="lightbulb_outline" class="no-results ma-2 opacity-70" outline
        >
          <h3 class="body-2 ma-0 pa-0">
            <translate>No pictures found</translate>
          </h3>
          <p class="body-1 mt-2 mb-0 pa-0">
            <translate>Duplicates will be skipped and only appear once.</translate>
            <translate>In case pictures you expect are missing, please rescan your library and wait until indexing has been completed.</translate>
          </p>
        </v-alert>
        <v-layout row wrap class="search-results file-results cards-view" :class="{'select-results': selection.length > 0}">
          <v-flex
              v-for="(model, index) in results"
              :key="model.UID"
              xs6 sm4 md3 lg2 xxl1 d-flex
          >
            <v-card tile
                    :data-uid="model.UID"
                    class="result card"
                    :class="model.classes(selection.includes(model.UID))"
                    @contextmenu.stop="onContextMenu($event, index)"
            >
              <div class="card-background card"></div>
              <v-img
                  :src="model.thumbnailUrl('tile_500')"
                  :alt="model.Name"
                  :transition="false"
                  loading="lazy"
                  aspect-ratio="1"
                  class="card darken-1 clickable"
                  @touchstart.passive="input.touchStart($event, index)"
                  @touchend.stop.prevent="onClick($event, index)"
                  @mousedown.stop.prevent="input.mouseDown($event, index)"
                  @click.stop.prevent="onClick($event, index)"
              >
                <v-btn :ripple="false"
                       icon flat absolute
                       class="input-select"
                       @touchstart.stop.prevent="input.touchStart($event, index)"
                       @touchend.stop.prevent="onSelect($event, index)"
                       @touchmove.stop.prevent
                       @click.stop.prevent="onSelect($event, index)">
                  <v-icon color="white" class="select-on">check_circle</v-icon>
                  <v-icon color="white" class="select-off">radio_button_off</v-icon>
                </v-btn>
              </v-img>

              <v-card-title v-if="model.isFile()" primary-title class="pa-3 card-details"
                            style="user-select: none;">
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
              <v-card-title v-else primary-title class="pa-3 card-details">
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
          </v-flex>
        </v-layout>
      </v-container>
    </v-container>
  </div>
</template>

<script>
import Event from "pubsub-js";
import RestModel from "model/rest";
import {Folder} from "model/folder";
import Notify from "common/notify";
import {MaxItems} from "common/clipboard";
import download from "common/download";
import {Input, InputInvalid, ClickShort, ClickLong} from "common/input";

export default {
  name: 'PPageFiles',
  props: {
    staticFilter: {
      type: Object,
      default: () => {},
    },
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
      navIcon: this.$rtl ? 'navigate_before' : 'navigate_next',
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
        limit: 999,
        offset: 0,
      },
      titleRule: v => v.length <= this.$config.get('clip') || this.$gettext("Name too long"),
      input: new Input(),
      lastId: "",
      breadcrumbs: [],
    };
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
  created() {
    if (this.$config.deny("files", "access_library")) {
      this.$router.push({ name: "albums" });
      return;
    }

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
  methods: {
    getBreadcrumbs() {
      let result = [];
      let path = "/index/files";

      const crumbs = this.path.split("/");

      crumbs.forEach(dir => {
        if (dir) {
          path += "/" + dir;
          result.push({path: path, name: dir});
        }
      });

      return result;
    },
    openFile(index) {
      const model = this.results[index];

      if (model.isFile()) {
        // Open Edit Dialog
        Event.publish("dialog.edit", {selection: [model.PhotoUID], album: null, index: 0});
      } else {
        this.$router.push({path: '/index/files/' + model.Path});
      }
    },
    downloadFile(index) {
      Notify.success(this.$gettext("Downloading…"));

      const model = this.results[index];
      download(`${this.$config.apiUri}/dl/${model.Hash}?t=${this.$config.downloadToken}`, model.Name);
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
      const inputType = this.input.eval(ev, index);

      if (inputType !== ClickShort) {
        return;
      }

      if (ev.shiftKey) {
        this.selectRange(index, this.results);
      } else {
        this.toggleSelection(this.results[index].getId());
      }
    },
    onClick(ev, index) {
      const inputType = this.input.eval(ev, index);
      const longClick = inputType === ClickLong;

      if (inputType === InputInvalid) {
        return;
      }

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
        if (this.selection.length >= MaxItems) {
          Notify.warn(this.$gettext("Can't select more items"));
          return;
        }

        this.selection.push(uid);
        this.lastId = uid;
      }
    },
    toggleSelection(uid) {
      const pos = this.selection.indexOf(uid);

      if (pos !== -1) {
        this.selection.splice(pos, 1);
        this.lastId = "";
      } else {
        if (this.selection.length >= MaxItems) {
          Notify.warn(this.$gettext("Can't select more items"));
          return;
        }

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
      this.filter.q = this.filter.q.trim();

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
        return;
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

      if (!data || !data.entities || !Array.isArray(data.entities)) {
        return;
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

            this.removeSelection(ppid);
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
};
</script>
