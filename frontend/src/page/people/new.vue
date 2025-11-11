<template>
  <div class="p-page p-page-faces not-selectable">
    <v-form ref="form" validate-on="invalid-input" class="p-faces-search" @submit.prevent="updateQuery">
      <v-toolbar density="compact" class="page-toolbar" color="secondary-light">
        <v-spacer></v-spacer>

        <v-btn :title="$gettext('Refresh')" icon="mdi-refresh" class="action-reload" @click.stop="refresh">
        </v-btn>

        <v-btn
          v-if="!filter.hidden"
          :title="$gettext('Show hidden')"
          icon="mdi-eye"
          class="action-show-hidden"
          @click.stop="onShowHidden"
        >
        </v-btn>
        <v-btn
          v-else
          :title="$gettext('Exclude hidden')"
          icon="mdi-eye-off"
          class="action-exclude-hidden"
          @click.stop="onExcludeHidden"
        >
        </v-btn>
      </v-toolbar>
    </v-form>

    <div v-if="loading" class="p-page__loading">
      <p-loading></p-loading>
    </div>
    <div v-else class="p-page__content">
      <p-scroll
        :load-more="loadMore"
        :load-disabled="scrollDisabled"
        :load-distance="scrollDistance"
        :loading="loading"
      ></p-scroll>

      <div v-if="results.length === 0" class="pa-3">
        <v-alert color="primary" icon="mdi-check-circle-outline" class="no-results" variant="outlined">
          <div class="font-weight-bold">
            {{ $gettext(`No people found`) }}
          </div>
          <div class="mt-2">
            {{ $gettext(`You may rescan your library to find additional faces.`) }}
            {{ $gettext(`Recognition starts after indexing has been completed.`) }}
          </div>
        </v-alert>
        <div class="d-flex justify-center my-8">
          <v-btn color="button" rounded variant="flat" :to="{ name: 'all', query: { q: 'face:new' } }">
            {{ $gettext(`Show all new faces`) }}
          </v-btn>
        </div>
      </div>
      <div v-else>
        <div class="v-row search-results face-results cards-view" :class="{ 'select-results': selection.length > 0 }">
          <div v-for="m in results" :key="m.ID" class="v-col-12 v-col-sm-6 v-col-md-4 v-col-lg-3 v-col-xl-2">
            <div :data-id="m.ID" :class="m.classes()" class="result flex-grow-1 not-selectable">
              <v-img :src="m.thumbnailUrl('tile_320')" aspect-ratio="1" class="preview" @click.stop.prevent="onView(m)">
                <v-btn
                  :ripple="false"
                  class="input-hidden"
                  icon
                  variant="text"
                  density="comfortable"
                  position="absolute"
                  @click.stop.prevent="toggleHidden(m)"
                >
                  <v-icon color="white" class="select-on" :title="$gettext('Show')">mdi-eye-off</v-icon>
                  <v-icon color="white" class="select-off" :title="$gettext('Hide')">mdi-close</v-icon>
                </v-btn>
              </v-img>

              <v-card-actions class="meta pa-0">
                <v-text-field
                  v-if="m.SubjUID"
                  v-model="m.Name"
                  :rules="[textRule]"
                  :readonly="readonly"
                  autocomplete="off"
                  hide-details
                  single-line
                  density="comfortable"
                  class="input-name pa-0 ma-0"
                  @blur="(ev) => onSetName(m, ev)"
                  @keyup.enter="(ev) => onSetName(m, ev)"
                ></v-text-field>
                <v-combobox
                  v-else
                  v-model:search="m.Name"
                  :items="$config.values.people"
                  item-title="Name"
                  item-value="Name"
                  :readonly="readonly"
                  :menu-props="menuProps"
                  return-object
                  hide-no-data
                  hide-details
                  single-line
                  open-on-clear
                  append-icon=""
                  prepend-inner-icon="mdi-account-plus"
                  autocomplete="off"
                  density="comfortable"
                  class="input-name pa-0 ma-0 text-selectable"
                  @update:model-value="(person) => onSetPerson(m, person)"
                  @blur="(ev) => onSetName(m, ev)"
                  @keyup.enter="(ev) => onSetName(m, ev)"
                >
                </v-combobox>
              </v-card-actions>
            </div>
          </div>
        </div>
        <div class="d-flex justify-center my-8">
          <v-btn color="button" rounded variant="flat" :to="{ name: 'all', query: { q: 'face:new' } }">
            {{ $gettext(`Show all new faces`) }}
          </v-btn>
        </div>
      </div>
    </div>
    <p-confirm-dialog
      :visible="confirm.visible"
      icon="mdi-account-plus"
      :icon-size="42"
      :text="confirm?.model?.Name ? $gettext('Add %{s}?', { s: confirm.model.Name }) : $gettext('Add person?')"
      @close="onCancelRename"
      @confirm="onConfirmRename"
    ></p-confirm-dialog>
  </div>
</template>

<script>
import Face from "model/face";
import RestModel from "model/rest";
import { MaxItems } from "common/clipboard";
import $notify from "common/notify";
import { ClickLong, ClickShort, Input, InputInvalid } from "common/input";
import PConfirmDialog from "component/confirm/dialog.vue";
import PLoading from "component/loading.vue";

export default {
  name: "PPageFaces",
  components: {
    PLoading,
    PConfirmDialog,
  },
  props: {
    staticFilter: {
      type: Object,
      default: () => {},
    },
    active: Boolean,
  },
  emits: ["updateFaceCount"],
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
      confirm: {
        visible: false,
        model: new Face(),
        text: this.$gettext("Add person?"),
      },
      menuProps: {
        openOnClick: false,
        openOnFocus: true,
        closeOnBack: true,
        closeOnContentClick: true,
        disableInitialFocus: true,
        persistent: false,
        scrim: true,
        openDelay: 0,
        closeDelay: 0,
        opacity: 0,
        density: "compact",
        maxHeight: 300,
        locationStrategy: "connected",
        scrollStrategy: "reposition",
        origin: "auto",
      },
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

    this.subscriptions.push(this.$event.subscribe("faces", (ev, data) => this.onUpdate(ev, data)));
    this.subscriptions.push(this.$event.subscribe("touchmove.top", () => this.refresh()));
  },
  beforeUnmount() {
    for (let i = 0; i < this.subscriptions.length; i++) {
      this.$event.unsubscribe(this.subscriptions[i]);
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
    onUpdate(ev, data) {
      if (!this.listen) {
        return;
      }

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
      if (this.loading || !this.active) {
        return false;
      }

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
        return false;
      }

      this.$router.replace({ query: query });

      return true;
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
      if (this.loading || !this.active || !this.listen || this.busy) {
        return;
      }

      // Make sure enough results are loaded to maintain the scroll position.
      if (this.page > 2) {
        this.page = this.page - 1;
      } else {
        this.page = 1;
      }

      // Flag results as dirty to force a refresh.
      this.dirty = true;

      // Enable infinite scrolling if it was disabled.
      this.scrollDisabled = false;

      this.loadMore();
    },
    reset() {
      this.results = [];
    },
    search() {
      this.scrollDisabled = true;

      // Don't query the same data more than once.
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
        .catch(() => {
          this.reset();
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
    onSetPerson(model, person) {
      if (typeof person === "object" && model?.ID && person?.UID && person?.Name) {
        model.Name = person.Name;
        model.SubjUID = person.UID;
        this.setName(model, person.Name);
      }

      return true;
    },
    onSetName(model, ev) {
      if (this.busy || !model) {
        return;
      }

      // If there's a pending confirmation for a different face, don't process new input
      if (this.confirm.visible && this.confirm.model && this.confirm.model.ID !== model.ID) {
        return;
      }

      const name = model?.Name;

      if (!name) {
        this.onCancelRename();
        return;
      }

      this.confirm.model = model;

      const people = this.$config.values?.people;

      if (people) {
        const found = people.find((person) => person.Name.localeCompare(name, "en", { sensitivity: "base" }) === 0);
        if (found) {
          model.Name = found.Name;
          model.SubjUID = found.UID;
          if (model.wasChanged()) {
            this.setName(model, model.Name);
          }
          return;
        }
      }

      model.Name = name;
      model.SubjUID = "";

      if (model.Name) {
        if (ev && ev.key === "Enter" && !ev.isComposing && !ev.repeat) {
          this.setName(model, model.Name);
        } else {
          this.confirm.visible = true;
        }
      }
    },
    onConfirmRename() {
      if (!this.confirm?.model?.Name) {
        return;
      }

      if (this.confirm.model.wasChanged()) {
        this.setName(this.confirm.model, this.confirm?.model?.Name);
      } else {
        this.confirm.model = null;
        this.confirm.visible = false;
      }
    },
    onCancelRename() {
      if (this.confirm && this.confirm.model) {
        this.confirm.model.Name = "";
        this.confirm.model.SubjUID = "";
      }
      this.confirm.visible = false;
    },
    setName(model, newName) {
      if (this.busy || !model || !newName || newName.trim() === "") {
        // Ignore if busy, refuse to save empty name.
        return;
      }

      this.busy = true;
      this.$notify.blockUI("busy");

      return model.setName(newName).finally(() => {
        this.$notify.unblockUI();
        this.busy = false;
        this.confirm.model = null;
        this.confirm.visible = false;
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
  },
};
</script>
