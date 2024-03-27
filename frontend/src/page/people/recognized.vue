<template>
  <div v-infinite-scroll="loadMore" class="p-page p-page-subjects" style="user-select: none" :infinite-scroll-disabled="scrollDisabled" :infinite-scroll-distance="scrollDistance" :infinite-scroll-listen-for-event="'scrollRefresh'">
    <v-form ref="form" class="p-people-search" lazy-validation dense @submit.prevent="updateQuery()">
      <v-toolbar dense flat class="page-toolbar" color="secondary-light pa-0">
        <v-text-field
          v-if="canSearch"
          :value="filter.q"
          solo
          hide-details
          clearable
          overflow
          single-line
          validate-on-blur
          class="input-search background-inherit elevation-0"
          :label="$gettext('Search')"
          prepend-inner-icon="search"
          browser-autocomplete="off"
          autocorrect="off"
          autocapitalize="none"
          color="secondary-dark"
          @change="
            (v) => {
              updateFilter({ q: v });
            }
          "
          @keyup.enter.native="(e) => updateQuery({ q: e.target.value })"
          @click:clear="
            () => {
              updateQuery({ q: '' });
            }
          "
        ></v-text-field>

        <v-divider vertical></v-divider>

        <v-btn icon overflow flat depressed color="secondary-dark" class="action-reload" :title="$gettext('Reload')" @click.stop="refresh()">
          <v-icon>refresh</v-icon>
        </v-btn>

        <template v-if="canManage">
          <v-btn v-if="!filter.hidden" icon class="action-show-hidden" :title="$gettext('Show hidden')" @click.stop="onShowHidden()">
            <v-icon>visibility</v-icon>
          </v-btn>
          <v-btn v-else icon class="action-exclude-hidden" :title="$gettext('Exclude hidden')" @click.stop="onExcludeHidden()">
            <v-icon>visibility_off</v-icon>
          </v-btn>
        </template>
      </v-toolbar>
    </v-form>

    <v-container v-if="loading" fluid class="pa-4">
      <v-progress-linear color="secondary-dark" :indeterminate="true"></v-progress-linear>
    </v-container>
    <v-container v-else fluid class="pa-0">
      <p-subject-clipboard :refresh="refresh" :selection="selection" :clear-selection="clearSelection"></p-subject-clipboard>

      <p-scroll-top></p-scroll-top>

      <v-container grid-list-xs fluid class="pa-2">
        <v-alert :value="results.length === 0" color="secondary-dark" icon="lightbulb_outline" class="no-results ma-2 opacity-70" outline>
          <h3 class="body-2 ma-0 pa-0">
            <translate>No people found</translate>
          </h3>
          <p class="body-1 mt-2 mb-0 pa-0">
            <translate>Try again using other filters or keywords.</translate>
            <translate>You may rescan your library to find additional faces.</translate>
            <translate>Recognition starts after indexing has been completed.</translate>
          </p>
        </v-alert>
        <v-layout row wrap class="search-results subject-results cards-view" :class="{ 'select-results': selection.length > 0 }">
          <v-flex v-for="(model, index) in results" :key="model.UID" xs6 sm4 md3 lg2 xxl1 d-flex>
            <v-card tile :data-uid="model.UID" style="user-select: none" class="result card" :class="model.classes(selection.includes(model.UID))" :to="model.route(view)" @contextmenu.stop="onContextMenu($event, index)">
              <div class="card-background card"></div>
              <v-img
                :src="model.thumbnailUrl('tile_320')"
                :alt="model.Name"
                :transition="false"
                aspect-ratio="1"
                style="user-select: none"
                class="card darken-1 clickable"
                @touchstart.passive="input.touchStart($event, index)"
                @touchend.stop.prevent="onClick($event, index)"
                @mousedown.stop.prevent="input.mouseDown($event, index)"
                @click.stop.prevent="onClick($event, index)"
              >
                <v-btn
                  v-if="canManage"
                  :ripple="false"
                  :depressed="false"
                  class="input-hidden"
                  icon
                  flat
                  small
                  absolute
                  @touchstart.stop.prevent="input.touchStart($event, index)"
                  @touchend.stop.prevent="onToggleHidden($event, index)"
                  @touchmove.stop.prevent
                  @click.stop.prevent="onToggleHidden($event, index)"
                >
                  <v-icon color="white" class="select-on" :title="$gettext('Show')">visibility_off</v-icon>
                  <v-icon color="white" class="select-off" :title="$gettext('Hide')">clear</v-icon>
                </v-btn>
                <v-btn :ripple="false" icon flat absolute class="input-select" @touchstart.stop.prevent="input.touchStart($event, index)" @touchend.stop.prevent="onSelect($event, index)" @touchmove.stop.prevent @click.stop.prevent="onSelect($event, index)">
                  <v-icon color="white" class="select-on">check_circle</v-icon>
                  <v-icon color="white" class="select-off">radio_button_off</v-icon>
                </v-btn>

                <v-btn :ripple="false" icon flat absolute class="input-favorite" @touchstart.stop.prevent="input.touchStart($event, index)" @touchend.stop.prevent="toggleLike($event, index)" @touchmove.stop.prevent @click.stop.prevent="toggleLike($event, index)">
                  <v-icon color="#FFD600" class="select-on">star</v-icon>
                  <v-icon color="white" class="select-off">star_border</v-icon>
                </v-btn>
              </v-img>

              <v-card-title primary-title class="pa-3 card-details" style="user-select: none" @click.stop.prevent="">
                <v-edit-dialog v-if="canManage" :return-value.sync="model.Name" lazy class="inline-edit" @save="onSave(model)">
                  <span v-if="model.Name" class="body-2 ma-0">
                    {{ model.Name }}
                  </span>
                  <span v-else>
                    <v-icon>edit</v-icon>
                  </span>
                  <template #input>
                    <v-text-field v-model="model.Name" :rules="[titleRule]" :readonly="readonly" :label="$gettext('Name')" color="secondary-dark" class="input-rename background-inherit elevation-0" single-line autofocus solo hide-details></v-text-field>
                  </template>
                </v-edit-dialog>
                <span v-else class="body-2 ma-0">
                  {{ model.Name }}
                </span>
              </v-card-title>

              <v-card-text primary-title class="pb-2 pt-0 card-details" style="user-select: none" @click.stop.prevent="">
                <div v-if="model.About" class="caption mb-2" :title="$gettext('About')">
                  {{ model.About | truncate(100) }}
                </div>

                <div class="caption mb-2">
                  <button v-if="model.PhotoCount === 1">
                    <translate>Contains one picture.</translate>
                  </button>
                  <button v-else-if="model.PhotoCount > 0">
                    <translate :translate-params="{ n: model.PhotoCount }">Contains %{n} pictures.</translate>
                  </button>
                </div>
              </v-card-text>
            </v-card>
          </v-flex>
        </v-layout>
      </v-container>
    </v-container>
    <p-people-merge-dialog lazy :show="merge.show" :subj1="merge.subj1" :subj2="merge.subj2" @cancel="onCancelMerge" @confirm="onMerge"></p-people-merge-dialog>
  </div>
</template>

<script>
import Subject from "model/subject";
import Event from "pubsub-js";
import RestModel from "model/rest";
import { MaxItems } from "common/clipboard";
import Notify from "common/notify";
import { ClickLong, ClickShort, Input, InputInvalid } from "common/input";

export default {
  name: "PPageSubjects",
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

    return {
      canView: this.$config.allow("people", "view"),
      canSearch: this.$config.allow("people", "search"),
      canManage: this.$config.allow("people", "manage"),
      view: "all",
      config: this.$config.values,
      subscriptions: [],
      listen: false,
      dirty: false,
      results: [],
      scrollDisabled: true,
      scrollDistance: window.innerHeight * 2,
      loading: true,
      batchSize: Subject.batchSize(),
      offset: 0,
      page: 0,
      selection: [],
      settings: {},
      filter: { q, hidden, order },
      lastFilter: {},
      routeName: routeName,
      titleRule: (v) => v.length <= this.$config.get("clip") || this.$gettext("Name too long"),
      input: new Input(),
      lastId: "",
      merge: {
        subj1: null,
        subj2: null,
        show: false,
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

      this.routeName = this.$route.name;
      this.filter.q = query["q"] ? query["q"] : "";
      this.filter.hidden = query["hidden"] ? query["hidden"] : "";
      this.filter.order = this.sortOrder();

      this.search();
    },
  },
  created() {
    this.search();

    this.subscriptions.push(Event.subscribe("subjects", (ev, data) => this.onUpdate(ev, data)));

    this.subscriptions.push(Event.subscribe("touchmove.top", () => this.refresh()));
    this.subscriptions.push(Event.subscribe("touchmove.bottom", () => this.loadMore()));
  },
  destroyed() {
    for (let i = 0; i < this.subscriptions.length; i++) {
      Event.unsubscribe(this.subscriptions[i]);
    }
  },
  methods: {
    onSave(m) {
      if (!this.canManage || !m.Name || m.Name.trim() === "") {
        // Refuse to save empty name.
        return;
      }

      const existing = this.$config.getPerson(m.Name);

      if (!existing) {
        this.busy = true;
        m.update().finally(() => {
          this.busy = false;
        });
      } else if (existing.UID !== m.UID) {
        this.merge.subj1 = m;
        this.merge.subj2 = existing;
        this.merge.show = true;
      }
    },
    onCancelMerge() {
      this.merge.subj1.Name = this.merge.subj1.originalValue("Name");
      this.merge.show = false;
      this.merge.subj1 = null;
      this.merge.subj2 = null;
    },
    onMerge() {
      if (!this.canManage) {
        return;
      }

      this.busy = true;
      this.merge.show = false;
      this.$notify.blockUI();
      this.merge.subj1.update().finally(() => {
        this.busy = false;
        this.merge.subj1 = null;
        this.merge.subj2 = null;
        this.$notify.unblockUI();
        this.refresh();
      });
    },
    searchCount() {
      const offset = parseInt(window.localStorage.getItem("subjects_offset"));

      if (this.offset > 0 || !offset) {
        return this.batchSize;
      }

      return offset + this.batchSize;
    },
    sortOrder() {
      return "relevance";
    },
    setOffset(offset) {
      this.offset = offset;
      window.localStorage.setItem("subjects_offset", offset);
    },
    toggleLike(ev, index) {
      if (!this.canManage) {
        return;
      }

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
    onShowHidden() {
      if (!this.canManage) {
        return;
      }

      this.showHidden("yes");
    },
    onExcludeHidden() {
      if (!this.canManage) {
        return;
      }

      this.showHidden("");
    },
    showHidden(value) {
      if (!this.canManage) {
        return;
      }

      this.filter.hidden = value;
      this.updateQuery();
    },
    onToggleHidden(ev, index) {
      if (!this.canManage) {
        return;
      }

      const inputType = this.input.eval(ev, index);

      if (inputType !== ClickShort) {
        return;
      }

      this.toggleHidden(this.results[index]);
    },
    toggleHidden(model) {
      if (!model || !this.canManage) {
        return;
      }
      this.busy = true;
      model.toggleHidden().finally(() => {
        this.busy = false;
      });
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

      const count = this.dirty ? (this.page + 2) * this.batchSize : this.batchSize;
      const offset = this.dirty ? 0 : this.offset;

      const params = {
        count: count,
        offset: offset,
      };

      Object.assign(params, this.lastFilter);

      if (this.staticFilter) {
        Object.assign(params, this.staticFilter);
      }

      Subject.search(params)
        .then((resp) => {
          this.results = this.dirty ? resp.models : this.results.concat(resp.models);

          this.scrollDisabled = resp.count < resp.limit;

          if (this.scrollDisabled) {
            this.setOffset(resp.offset);
            if (this.results.length > 1) {
              this.$notify.info(this.$gettextInterpolate(this.$gettext("All %{n} people loaded"), { n: this.results.length }));
            }
          } else {
            this.setOffset(resp.offset + resp.limit);
            this.page++;

            this.$nextTick(() => {
              if (this.$root.$el.clientHeight <= window.document.documentElement.clientHeight + 300) {
                this.$emit("scrollRefresh");
              }
            });
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
    updateSettings(props) {
      if (!props || typeof props !== "object" || props.target) {
        return;
      }

      for (const [key, value] of Object.entries(props)) {
        if (!this.settings.hasOwnProperty(key)) {
          continue;
        }
        switch (typeof value) {
          case "string":
            this.settings[key] = value.trim();
            break;
          default:
            this.settings[key] = value;
        }

        window.localStorage.setItem("people_" + key, this.settings[key]);
      }
    },
    updateFilter(props) {
      if (!props || typeof props !== "object" || props.target) {
        return;
      }

      for (const [key, value] of Object.entries(props)) {
        if (!this.filter.hasOwnProperty(key)) {
          continue;
        }
        switch (typeof value) {
          case "string":
            this.filter[key] = value.trim();
            break;
          default:
            this.filter[key] = value;
        }
      }
    },
    updateQuery(props) {
      this.updateFilter(props);

      if (this.loading || !this.active) {
        return;
      }

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
    refresh(props) {
      this.updateSettings(props);

      if (this.loading || !this.active) return;

      this.loading = true;
      this.page = 0;
      this.dirty = true;
      this.scrollDisabled = false;

      this.loadMore();
    },
    search() {
      /**
       * re-creating the last scroll-position should only ever happen when using
       * back-navigation. We therefore reset the remembered scroll-position
       * in any other scenario
       */
      if (!window.backwardsNavigationDetected) {
        this.setOffset(0);
      }

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

      Subject.search(params)
        .then((resp) => {
          this.offset = resp.limit;
          this.results = resp.models;

          this.scrollDisabled = resp.count < resp.limit;

          if (this.scrollDisabled) {
            if (!this.results.length) {
              this.$notify.warn(this.$gettext("No people found"));
            } else if (this.results.length === 1) {
              this.$notify.info(this.$gettext("One person found"));
            } else {
              this.$notify.info(this.$gettextInterpolate(this.$gettext("%{n} people found"), { n: this.results.length }));
            }
          } else {
            // this.$notify.info(this.$gettext('More than 20 people found'));
            this.$nextTick(() => {
              if (this.$root.$el.clientHeight <= window.document.documentElement.clientHeight + 300) {
                this.$emit("scrollRefresh");
              }
            });
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
