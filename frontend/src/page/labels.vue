<template>
  <div v-infinite-scroll="loadMore" :class="$config.aclClasses('labels')" class="p-page p-page-labels" style="user-select: none" :infinite-scroll-disabled="scrollDisabled" :infinite-scroll-distance="scrollDistance" :infinite-scroll-listen-for-event="'scrollRefresh'">
    <v-form ref="form" class="p-labels-search" lazy-validation dense @submit.stop.prevent="updateQuery()">
      <v-toolbar flat :dense="$vuetify.breakpoint.smAndDown" class="page-toolbar" color="secondary">
        <v-text-field
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

        <v-btn icon class="action-reload" :title="$gettext('Reload')" @click.stop="refresh()">
          <v-icon>refresh</v-icon>
        </v-btn>

        <v-btn v-if="!filter.all" icon class="action-show-all" :title="$gettext('Show more')" @click.stop="showAll()">
          <v-icon>unfold_more</v-icon>
        </v-btn>
        <v-btn v-else icon class="action-show-important" :title="$gettext('Show less')" @click.stop="showImportant()">
          <v-icon>unfold_less</v-icon>
        </v-btn>
      </v-toolbar>
    </v-form>

    <v-container v-if="loading" fluid class="pa-4">
      <v-progress-linear color="secondary-dark" :indeterminate="true"></v-progress-linear>
    </v-container>
    <v-container v-else fluid class="pa-0">
      <p-label-clipboard v-if="canSelect" :refresh="refresh" :selection="selection" :clear-selection="clearSelection"></p-label-clipboard>

      <p-scroll-top></p-scroll-top>

      <v-container grid-list-xs fluid class="pa-2">
        <v-alert :value="results.length === 0" color="secondary-dark" icon="lightbulb_outline" class="no-results ma-2 opacity-70" outline>
          <h3 class="body-2 ma-0 pa-0">
            <translate>No labels found</translate>
          </h3>
          <p class="body-1 mt-2 mb-0 pa-0">
            <translate>Try again using other filters or keywords.</translate>
            <translate>In case pictures you expect are missing, please rescan your library and wait until indexing has been completed.</translate>
          </p>
        </v-alert>
        <v-layout row wrap class="search-results label-results cards-view" :class="{ 'select-results': selection.length > 0 }">
          <v-flex v-for="(label, index) in results" :key="label.UID" xs6 sm4 md3 lg2 xxl1 d-flex>
            <v-card tile :data-uid="label.UID" style="user-select: none" class="result card" :class="label.classes(selection.includes(label.UID))" :to="label.route(view)" @contextmenu.stop="onContextMenu($event, index)">
              <div class="card-background card"></div>
              <v-img
                :src="label.thumbnailUrl('tile_500')"
                :alt="label.Name"
                :transition="false"
                aspect-ratio="1"
                style="user-select: none"
                class="card darken-1 clickable"
                @touchstart.passive="input.touchStart($event, index)"
                @touchend.stop.prevent="onClick($event, index)"
                @mousedown.stop.prevent="input.mouseDown($event, index)"
                @click.stop.prevent="onClick($event, index)"
              >
                <v-btn v-if="canSelect" :ripple="false" icon flat absolute class="input-select" @touchstart.stop.prevent="input.touchStart($event, index)" @touchend.stop.prevent="onSelect($event, index)" @touchmove.stop.prevent @click.stop.prevent="onSelect($event, index)">
                  <v-icon color="white" class="select-on">check_circle</v-icon>
                  <v-icon color="white" class="select-off">radio_button_off</v-icon>
                </v-btn>

                <v-btn :ripple="false" icon flat absolute class="input-favorite" @touchstart.stop.prevent="input.touchStart($event, index)" @touchend.stop.prevent="toggleLike($event, index)" @touchmove.stop.prevent @click.stop.prevent="toggleLike($event, index)">
                  <v-icon color="#FFD600" class="select-on">star</v-icon>
                  <v-icon color="white" class="select-off">star_border</v-icon>
                </v-btn>
              </v-img>

              <v-card-title primary-title class="pa-3 card-details" style="user-select: none" @click.stop.prevent="">
                <v-edit-dialog v-if="canManage" :return-value.sync="label.Name" lazy class="inline-edit" @save="onSave(label)">
                  <span v-if="label.Name" class="body-2 ma-0">
                    {{ label.Name }}
                  </span>
                  <span v-else>
                    <v-icon>edit</v-icon>
                  </span>
                  <template #input>
                    <v-text-field v-model="label.Name" :rules="[titleRule]" :label="$gettext('Name')" color="secondary-dark" class="input-rename background-inherit elevation-0" single-line autofocus solo hide-details></v-text-field>
                  </template>
                </v-edit-dialog>
                <span v-else class="body-2 ma-0">
                  {{ label.Name }}
                </span>
              </v-card-title>

              <v-card-text primary-title class="pb-2 pt-0 card-details" style="user-select: none" @click.stop.prevent="">
                <div class="caption mb-2">
                  <button v-if="label.PhotoCount === 1">
                    <translate>Contains one picture.</translate>
                  </button>
                  <button v-else-if="label.PhotoCount > 0">
                    <translate :translate-params="{ n: label.PhotoCount }">Contains %{n} pictures.</translate>
                  </button>
                </div>
              </v-card-text>
            </v-card>
          </v-flex>
        </v-layout>
      </v-container>
    </v-container>
  </div>
</template>

<script>
import Label from "model/label";
import Event from "pubsub-js";
import RestModel from "model/rest";
import { MaxItems } from "common/clipboard";
import Notify from "common/notify";
import { Input, InputInvalid, ClickShort, ClickLong } from "common/input";

export default {
  name: "PPageLabels",
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

    const canManage = this.$config.allow("labels", "manage");
    const canAddAlbums = this.$config.allow("albums", "create") && this.$config.feature("albums");

    return {
      canManage: canManage,
      canSelect: canManage || canAddAlbums,
      view: "all",
      config: this.$config.values,
      subscriptions: [],
      listen: false,
      dirty: false,
      results: [],
      scrollDisabled: true,
      scrollDistance: window.innerHeight * 2,
      loading: true,
      batchSize: Label.batchSize(),
      offset: 0,
      page: 0,
      selection: [],
      settings: {},
      filter: { q, all },
      lastFilter: {},
      routeName: routeName,
      titleRule: (v) => v.length <= this.$config.get("clip") || this.$gettext("Name too long"),
      input: new Input(),
      lastId: "",
    };
  },
  watch: {
    $route() {
      const query = this.$route.query;

      this.routeName = this.$route.name;
      this.lastFilter = {};
      this.filter.q = query["q"] ? query["q"] : "";
      this.filter.all = query["all"] ? query["all"] : "";

      this.search();
    },
  },
  created() {
    this.search();

    this.subscriptions.push(Event.subscribe("labels", (ev, data) => this.onUpdate(ev, data)));

    this.subscriptions.push(Event.subscribe("touchmove.top", () => this.refresh()));
    this.subscriptions.push(Event.subscribe("touchmove.bottom", () => this.loadMore()));
  },
  destroyed() {
    for (let i = 0; i < this.subscriptions.length; i++) {
      Event.unsubscribe(this.subscriptions[i]);
    }
  },
  methods: {
    searchCount() {
      const offset = parseInt(window.localStorage.getItem("labels_offset"));

      if (this.offset > 0 || !offset) {
        return this.batchSize;
      }

      return offset + this.batchSize;
    },
    setOffset(offset) {
      this.offset = offset;
      window.localStorage.setItem("labels_offset", offset);
    },
    toggleLike(ev, index) {
      if (!this.canManage) {
        return;
      }

      const inputType = this.input.eval(ev, index);

      if (inputType !== ClickShort) {
        return;
      }

      const label = this.results[index];

      if (!label) {
        return;
      }

      label.toggleLike();
    },
    selectRange(rangeEnd, models) {
      if (!this.canSelect) {
        return;
      } else if (!models || !models[rangeEnd] || !(models[rangeEnd] instanceof RestModel)) {
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
      if (!this.canSelect) {
        return;
      }

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
      if (!this.canSelect) {
        return;
      }

      if (this.$isMobile) {
        ev.preventDefault();
        ev.stopPropagation();

        if (this.results[index]) {
          this.selectRange(index, this.results);
        }
      }
    },
    onSave(label) {
      if (!this.canManage) {
        return;
      }

      label.update();
    },
    showAll() {
      this.filter.all = "true";
      this.updateQuery();
    },
    showImportant() {
      this.filter.all = "";
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
      if (!this.canSelect) {
        return;
      }

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
      if (this.scrollDisabled || this.$scrollbar.disabled()) return;

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

      if (offset === 0) {
        this.results = [];
      }

      Label.search(params)
        .then((resp) => {
          this.results = offset === 0 ? resp.models : this.results.concat(resp.models);

          this.scrollDisabled = resp.count < resp.limit;

          if (this.scrollDisabled) {
            this.setOffset(resp.offset);
            if (this.results.length > 1) {
              this.$notify.info(this.$gettextInterpolate(this.$gettext("All %{n} labels loaded"), { n: this.results.length }));
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

        window.localStorage.setItem("labels_" + key, this.settings[key]);
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

      if (this.loading) return;

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

      if (this.loading) return;

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

      Label.search(params)
        .then((resp) => {
          this.offset = resp.limit;
          this.results = resp.models;

          this.scrollDisabled = resp.count < resp.limit;

          if (this.scrollDisabled) {
            if (!this.results.length) {
              this.$notify.warn(this.$gettext("No labels found"));
            } else if (this.results.length === 1) {
              this.$notify.info(this.$gettext("One label found"));
            } else {
              this.$notify.info(this.$gettextInterpolate(this.$gettext("%{n} labels found"), { n: this.results.length }));
            }
          } else {
            // this.$notify.info(this.$gettext('More than 20 labels found'));
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
