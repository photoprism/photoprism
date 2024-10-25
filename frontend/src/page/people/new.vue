<template>
  <div class="p-page p-page-faces" style="user-select: none">
    <v-form ref="form" class="p-faces-search" lazy-validation @submit.prevent="updateQuery">
      <v-toolbar dense class="page-toolbar" flat color="secondary-light pa-0">
        <v-spacer></v-spacer>
        <v-divider direction="vertical"></v-divider>

        <v-btn icon variant="text" color="secondary-dark" class="action-reload" :title="$gettext('Reload')" @click.stop="refresh">
          <v-icon>mdi-refresh</v-icon>
        </v-btn>

        <v-btn v-if="!filter.hidden" icon class="action-show-hidden" :title="$gettext('Show hidden')" @click.stop="onShowHidden">
          <v-icon>mdi-eye</v-icon>
        </v-btn>
        <v-btn v-else icon class="action-exclude-hidden" :title="$gettext('Exclude hidden')" @click.stop="onExcludeHidden">
          <v-icon>mdi-eye-off</v-icon>
        </v-btn>
      </v-toolbar>
    </v-form>

    <v-container v-if="loading" fluid class="pa-6">
      <v-progress-linear color="secondary-dark" :indeterminate="true"></v-progress-linear>
    </v-container>
    <v-container v-else fluid class="pa-0">
      <p-scroll-top></p-scroll-top>

      <v-container grid-list-xs fluid class="pa-2">
        <v-alert :value="results.length === 0" color="secondary-dark" icon="mdi-check-circle-outline" class="no-results ma-2 opacity-70" variant="outlined">
          <h3 class="text-body-2 ma-0 pa-0">
            <translate>No people found</translate>
          </h3>
          <p class="text-body-1 mt-2 mb-0 pa-0">
            <translate>You may rescan your library to find additional faces.</translate>
            <translate>Recognition starts after indexing has been completed.</translate>
          </p>
        </v-alert>
        <v-row class="search-results face-results cards-view" :class="{ 'select-results': selection.length > 0 }">
          <v-col v-for="model in results" :key="model.ID" cols="12" sm="6" md="4" lg="3" xl="2" xxl="1" class="d-flex">
            <v-card :data-id="model.ID" tile style="user-select: none" :class="model.classes()" class="result card flex-grow-1">
              <div class="card-background card"></div>
              <v-img :src="model.thumbnailUrl('tile_320')" :transition="false" aspect-ratio="1" class="card clickable" @click.stop.prevent="onView(model)">
                <v-btn :ripple="false" class="input-hidden" icon variant="text" size="small" absolute @click.stop.prevent="toggleHidden(model)">
                  <v-icon color="white" class="select-on" :title="$gettext('Show')">mdi-eye-off</v-icon>
                  <v-icon color="white" class="select-off" :title="$gettext('Hide')">mdi-close</v-icon>
                </v-btn>
              </v-img>

              <v-card-actions class="card-details pa-0">
                <v-row v-if="model.SubjUID" align="center">
                  <v-col cols="12" class="text-left pa-0">
                    <v-text-field
                      :model-value="model.Name"
                      :rules="[textRule]"
                      :readonly="readonly"
                      autocomplete="off"
                      class="input-name pa-0 ma-0"
                      hide-details
                      single-line
                      variant="solo-inverted"
                      @change="
                        (newName) => {
                          onRename(model, newName);
                        }
                      "
                      @keyup.enter="
                        (event) => {
                          onRename(model, event.target.value);
                        }
                      "
                    ></v-text-field>
                  </v-col>
                </v-row>
                <v-row align="center">
                  <v-col cols="12" class="text-left pa-0">
                    <!-- TODO: check property allow-overflow TEST -->
                    <v-combobox
                      :model-value="model.Name"
                      style="z-index: 250"
                      :items="$config.values.people"
                      item-value="Name"
                      item-title="Name"
                      :readonly="readonly"
                      :return-object="false"
                      :menu-props="menuProps"
                      :hint="$gettext('Name')"
                      hide-details
                      single-line
                      variant="solo-inverted"
                      open-on-clear
                      hide-no-data
                      append-icon=""
                      prepend-inner-icon="mdi-account-plus"
                      autocomplete="off"
                      class="input-name pa-0 ma-0"
                      @update:model-value="
                        (newName) => {
                          onRename(model, newName);
                        }
                      "
                      @keyup.enter.native="
                        (event) => {
                          onRename(model, event.target.value);
                        }
                      "
                    >
                    </v-combobox>
                  </v-col>
                </v-row>
              </v-card-actions>
            </v-card>
          </v-col>
        </v-row>
        <div class="text-center mt-4 mb-2">
          <v-btn color="secondary" rounded variant="flat" :to="{ name: 'all', query: { q: 'face:new' } }">
            <translate>Show all new faces</translate>
          </v-btn>
        </div>
      </v-container>
    </v-container>
  </div>
</template>

<script>
import Face from "model/face";
import Event from "pubsub-js";
import RestModel from "model/rest";
import { MaxItems } from "common/clipboard";
import Notify from "common/notify";
import { ClickLong, ClickShort, Input, InputInvalid } from "common/input";

export default {
  name: "PPageFaces",
  props: {
    staticFilter: {
      type: Object,
      default: () => {},
    },
    active: Boolean,
  },
  data() {
    const query = this.$route.query;
    const routeName = this.$route.name;
    const q = query["q"] ? query["q"] : "";
    const hidden = query["hidden"] ? query["hidden"] : "";
    const order = this.sortOrder();
    const filter = { q, hidden, order };
    const settings = {};

    return {
      view: "all",
      config: this.$config.values,
      subscriptions: [],
      listen: false,
      dirty: false,
      results: [],
      scrollDisabled: true,
      scrollDistance: window.innerHeight * 2,
      loading: true,
      busy: false,
      batchSize: 999,
      offset: 0,
      page: 0,
      faceCount: 0,
      selection: [],
      settings: settings,
      filter: filter,
      lastFilter: {},
      routeName: routeName,
      titleRule: (v) => v.length <= this.$config.get("clip") || this.$gettext("Name too long"),
      input: new Input(),
      lastId: "",
      menuProps: { closeOnClick: false, closeOnContentClick: true, openOnClick: false, maxHeight: 300 },
      textRule: (v) => {
        if (!v || !v.length) {
          return this.$gettext("Name");
        }

        return v.length <= this.$config.get("clip") || this.$gettext("Text too long");
      },
    };
  },
  computed: {
    readonly: function () {
      return this.busy || this.loading;
    },
  },
  watch: {
    $route() {
      // Tab inactive?
      if (!this.active) {
        // Ignore event.
        return;
      }

      const query = this.$route.query;

      this.filter.q = query["q"] ? query["q"] : "";
      this.filter.hidden = query["hidden"] ? query["hidden"] : "";
      this.filter.order = this.sortOrder();
      this.routeName = this.$route.name;

      this.search();
    },
  },
  created() {
    this.search();

    this.subscriptions.push(Event.subscribe("faces", (ev, data) => this.onUpdate(ev, data)));

    this.subscriptions.push(Event.subscribe("touchmove.top", () => this.refresh()));
  },
  unmounted() {
    for (let i = 0; i < this.subscriptions.length; i++) {
      Event.unsubscribe(this.subscriptions[i]);
    }
  },
  methods: {
    searchCount() {
      return this.batchSize;
    },
    sortOrder() {
      return "samples";
    },
    setOffset(offset) {
      this.offset = offset;
    },
    toggleLike(ev, index) {
      const inputType = this.input.eval(ev, index);

      if (inputType !== ClickShort) {
        return;
      }

      const m = this.results[index];

      if (!m) {
        return;
      }

      m.toggleLike();
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
        if (longClick || ev.shiftKey) {
          this.selectRange(index, this.results);
        } else {
          this.toggleSelection(this.results[index].getId());
        }
      } else {
        this.$router.push(this.results[index].route(this.view));
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
    onView(model) {
      if (this.loading || this.busy || !this.active) {
        // Don't redirect if page is not ready or active.
        return;
      }

      this.$router.push(model.route(this.view));
    },
    onSave(m) {
      m.update();
    },
    onShowHidden() {
      this.showHidden("yes");
    },
    onExcludeHidden() {
      this.showHidden("");
    },
    showHidden(value) {
      this.filter.hidden = value;
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
    loadMore() {
      if (this.scrollDisabled || !this.active) {
        return;
      }

      this.scrollDisabled = true;
      this.listen = false;

      // Always refresh all faces for now.
      this.dirty = true;

      const count = this.batchSize;
      const offset = 0;

      const params = {
        count: count,
        offset: offset,
      };

      Object.assign(params, this.lastFilter);

      if (this.staticFilter) {
        Object.assign(params, this.staticFilter);
      }

      Face.search(params)
        .then((resp) => {
          this.results = this.dirty ? resp.models : this.results.concat(resp.models);

          this.setFaceCount(this.results.length);

          if (!this.results.length) {
            this.$notify.warn(this.$gettext("No people found"));
          } else if (this.results.length === 1) {
            this.$notify.info(this.$gettext("One person found"));
          } else {
            this.$notify.info(this.$gettextInterpolate(this.$gettext("%{n} people found"), { n: this.results.length }));
          }
        })
        .catch(() => {
          this.scrollDisabled = false;
        })
        .finally(() => {
          this.dirty = false;
          this.loading = false;
          this.listen = true;
        });
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
        count: this.searchCount(),
        offset: this.offset,
      };

      Object.assign(params, this.filter);

      if (this.staticFilter) {
        Object.assign(params, this.staticFilter);
      }

      return params;
    },
    refresh() {
      if (this.loading || !this.active || this.busy) {
        return;
      }

      this.loading = true;
      this.page = 0;
      this.dirty = true;
      this.scrollDisabled = false;

      this.loadMore();
    },
    search() {
      this.scrollDisabled = true;

      // Don't query the same data more than once
      if (JSON.stringify(this.lastFilter) === JSON.stringify(this.filter)) {
        this.refresh();
        return;
      }

      Object.assign(this.lastFilter, this.filter);

      this.offset = 0;
      this.page = 0;
      this.loading = true;
      this.listen = false;

      const params = this.searchParams();

      Face.search(params)
        .then((resp) => {
          this.offset = resp.limit;
          this.results = resp.models;

          this.setFaceCount(this.results.length);

          if (!this.results.length) {
            this.$notify.warn(this.$gettext("No people found"));
          } else if (this.results.length === 1) {
            this.$notify.info(this.$gettext("One person found"));
          } else {
            this.$notify.info(this.$gettextInterpolate(this.$gettext("%{n} people found"), { n: this.results.length }));
          }
        })
        .finally(() => {
          this.dirty = false;
          this.loading = false;
          this.listen = true;
        });
    },
    onShow(model) {
      if (this.busy || !model) return;

      this.busy = true;
      model.show().finally(() => {
        this.busy = false;
        this.changeFaceCount(1);
      });
    },
    onHide(model) {
      if (this.busy || !model) return;

      this.busy = true;
      model.hide().finally(() => {
        this.busy = false;
        this.changeFaceCount(-1);
      });
    },
    toggleHidden(model) {
      if (this.busy || !model) return;

      this.busy = true;

      model.toggleHidden().finally(() => {
        this.busy = false;

        if (model.Hidden) {
          this.changeFaceCount(-1);
        } else {
          this.changeFaceCount(1);
        }
      });
    },
    onRename(model, newName) {
      if (this.busy || !model || !newName || newName.trim() === "") {
        // Ignore if busy, refuse to save empty name.
        return;
      }

      this.busy = true;
      this.$notify.blockUI();

      model.setName(newName).finally(() => {
        this.$notify.unblockUI();
        this.busy = false;
        this.changeFaceCount(-1);
      });
    },
    changeFaceCount(count) {
      this.faceCount = this.faceCount + count;
      this.$emit("updateFaceCount", this.faceCount);
    },
    setFaceCount(count) {
      this.faceCount = count;
      this.$emit("updateFaceCount", this.faceCount);
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

            if (model) {
              for (let key in values) {
                if (values.hasOwnProperty(key) && values[key] != null && typeof values[key] !== "object") {
                  model[key] = values[key];
                }
              }
            }
          }
          break;
        case "deleted":
          this.dirty = true;

          for (let i = 0; i < data.entities.length; i++) {
            const uid = data.entities[i];
            const index = this.results.findIndex((m) => m.UID === uid);

            if (index >= 0) {
              this.results.splice(index, 1);
            }

            this.removeSelection(uid);
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
