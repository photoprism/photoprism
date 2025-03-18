<template>
  <div ref="page" tabindex="1" class="p-page p-page-files">
    <v-form
      ref="form"
      validate-on="invalid-input"
      class="p-files-search p-page__navigation"
      @submit.prevent="updateQuery"
    >
      <v-toolbar flat color="secondary" :density="$vuetify.display.smAndDown ? 'compact' : 'default'">
        <v-toolbar-title>
          <router-link to="/index/files">
            {{ $gettext(`Originals`) }}
          </router-link>

          <router-link v-for="dir in breadcrumbs" :key="dir.key" :to="dir.uri">
            <v-icon>{{ navIcon }}</v-icon>
            {{ dir.name }}
          </router-link>
        </v-toolbar-title>

        <v-btn :title="$gettext('Refresh')" icon="mdi-refresh" class="action-reload" @click.stop="refresh"> </v-btn>
      </v-toolbar>
    </v-form>

    <div v-if="loading" class="p-page__loading">
      <p-loading></p-loading>
    </div>
    <div v-else class="p-page__content">
      <p-file-clipboard :refresh="refresh" :selection="selection" :clear-selection="clearSelection"></p-file-clipboard>

      <p-scroll :loading="loading"></p-scroll>

      <div class="p-files p-files-cards">
        <v-alert
          v-if="results.length === 0"
          color="surface-variant"
          icon="mdi-lightbulb-outline"
          class="ma-3 no-results opacity-60"
          variant="outlined"
        >
          <div class="font-weight-bold">
            {{ $gettext(`No pictures found`) }}
          </div>
          <div class="mt-2">
            {{ $gettext(`Duplicates will be skipped and only appear once.`) }}
            {{
              $gettext(
                `In case pictures you expect are missing, please rescan your library and wait until indexing has been completed.`
              )
            }}
          </div>
        </v-alert>
        <div
          v-else
          class="v-row search-results file-results cards-view"
          :class="{ 'select-results': selection.length > 0 }"
        >
          <div v-for="(m, index) in results" :key="m.UID" ref="items" class="v-col-6 v-col-sm-4 v-col-md-3 v-col-xl-2">
            <div
              :data-uid="m.UID"
              class="result"
              :class="m.classes(selection.includes(m.UID))"
              @contextmenu.stop="onContextMenu($event, index)"
            >
              <div
                :title="m.Name"
                :style="`background-image: url(${m.thumbnailUrl('tile_500')})`"
                class="preview"
                @touchstart.passive="input.touchStart($event, index)"
                @touchend.stop="onClick($event, index)"
                @mousedown.stop.prevent="input.mouseDown($event, index)"
                @click.stop.prevent="onClick($event, index)"
              >
                <div class="preview__overlay"></div>

                <button
                  class="input-select"
                  @touchstart.stop="input.touchStart($event, index)"
                  @touchend.stop="onSelect($event, index)"
                  @touchmove.stop.prevent
                  @click.stop.prevent="onSelect($event, index)"
                >
                  <i class="mdi mdi-check-circle select-on" />
                  <i class="mdi mdi-circle-outline select-off" />
                </button>
              </div>

              <div v-if="m.isFile()" class="meta">
                <button :title="m.Name" class="meta-title" @click.exact="openFile(index)">
                  {{ m.baseName() }}
                </button>
                <div class="meta-description">
                  {{ m.getInfo() }}
                </div>
              </div>
              <div v-else class="meta">
                <button :title="m.Title" class="meta-title" @click.exact="openFile(index)">
                  {{ m.baseName() }}
                </button>
                <div class="meta-description">
                  {{ $gettext(`Folder`) }}
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script>
import RestModel from "model/rest";
import { Folder } from "model/folder";
import $notify from "common/notify";
import { MaxItems } from "common/clipboard";
import download from "common/download";
import { Input, InputInvalid, ClickShort, ClickLong } from "common/input";
import PLoading from "component/loading.vue";

export default {
  name: "PPageFiles",
  components: { PLoading },
  props: {
    staticFilter: {
      type: Object,
      default: () => {},
    },
  },
  data() {
    const query = this.$route.query;
    const routeName = this.$route.name;
    const q = query["q"] ? query["q"] : "";
    const all = query["all"] ? query["all"] : "";
    const filter = { q: q, all: all };
    const settings = {};

    return {
      config: this.$config.values,
      navIcon: this.$isRtl ? "mdi-chevron-left" : "mdi-chevron-right",
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
      path: [],
      page: 0,
      files: {
        limit: 999,
        offset: 0,
      },
      titleRule: (v) => v.length <= this.$config.get("clip") || this.$gettext("Name too long"),
      input: new Input(),
      lastId: "",
      breadcrumbs: [],
    };
  },
  watch: {
    $route() {
      if (!this.$view.isActive(this)) {
        return;
      }

      this.$view.focus(this.$refs?.page);

      const query = this.$route.query;

      this.filter.q = query["q"] ? query["q"] : "";
      this.filter.all = query["all"] ? query["all"] : "";
      this.lastFilter = {};
      this.routeName = this.$route.name;
      this.path = this.$route.params.pathMatch;

      this.search();
    },
  },
  created() {
    if (this.$config.deny("files", "access_library") || this.$config.deny("files", "access_private")) {
      this.$router.push({ name: this.$session.getDefaultRoute() });
      return;
    }

    this.path = this.$route.params.pathMatch;

    this.search();

    this.subscriptions.push(this.$event.subscribe("folders", (ev, data) => this.onUpdate(ev, data)));
    this.subscriptions.push(this.$event.subscribe("touchmove.top", () => this.refresh()));
  },
  mounted() {
    this.$view.enter(this);
  },
  beforeUnmount() {
    for (let i = 0; i < this.subscriptions.length; i++) {
      this.$event.unsubscribe(this.subscriptions[i]);
    }
  },
  unmounted() {
    this.$view.leave(this);
  },
  methods: {
    getBreadcrumbs() {
      let result = [];
      let uri = "/index/files";
      let key = "B";

      const crumbs = [...this.path];

      crumbs.forEach((dir) => {
        if (dir) {
          key += "_" + dir;
          uri += "/" + dir;
          result.push({ key, uri, name: dir });
        }
      });

      return result;
    },
    openFile(index) {
      const model = this.results[index];

      if (model.isFile()) {
        // Open Edit Dialog
        this.$event.publish("dialog.edit", { selection: [model.PhotoUID], album: null, index: 0 });
      } else {
        // "#" chars in path names must be explicitly escaped,
        // see https://github.com/photoprism/photoprism/issues/3695
        const path = model.Path.replaceAll(":", "%3A").replaceAll("#", "%23");
        this.$router.push({ path: "/index/files/" + path });
      }
    },
    downloadFile(index) {
      $notify.success(this.$gettext("Downloading…"));

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

      return rangeEnd - rangeStart + 1;
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
      this.filter.q = "";
      this.updateQuery();
    },
    addSelection(uid) {
      const pos = this.selection.indexOf(uid);

      if (pos === -1) {
        if (this.selection.length >= MaxItems) {
          $notify.warn(this.$gettext("Can't select more items"));
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
          $notify.warn(this.$gettext("Can't select more items"));
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
        view: this.settings.view,
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

      this.$router.replace({ query: query });
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
    getPathAsString() {
      if (Array.isArray(this.path)) {
        return this.path.join("/");
      }

      return "";
    },
    search() {
      // Don't query the same data more than once
      if (!this.dirty && JSON.stringify(this.lastFilter) === JSON.stringify(this.filter)) {
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

      Folder.originals(this.getPathAsString(), params)
        .then((response) => {
          this.files.offset = this.files.limit;

          this.results = response.models;
          this.breadcrumbs = this.getBreadcrumbs();

          if (response.count === 0) {
            this.$notify.warn(this.$gettext("Folder is empty"));
          } else if (response.files === 1) {
            this.$notify.info(this.$gettext("One file found"));
          } else if (response.files === 0 && response.folders === 1) {
            this.$notify.info(this.$gettext("One folder found"));
          } else if (response.files === 0 && response.folders > 1) {
            this.$notify.info(this.$gettextInterpolate(this.$gettext("%{n} folders found"), { n: response.folders }));
          } else if (response.files < this.files.limit) {
            this.$notify.info(
              this.$gettextInterpolate(this.$gettext("Folder contains %{n} files"), { n: response.files })
            );
          } else {
            this.$notify.warn(
              this.$gettextInterpolate(this.$gettext("Limit reached, showing first %{n} files"), { n: response.files })
            );
          }
        })
        .finally(() => {
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

      const type = ev.split(".")[1];

      switch (type) {
        case "updated":
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
        case "deleted":
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
        case "created":
          this.dirty = true;
          break;
        default:
          console.warn("unexpected event type", ev);
      }
    },
  },
};
</script>
