<template>
  <div v-infinite-scroll="loadMore" class="p-page p-page-faces" style="user-select: none"
       :infinite-scroll-disabled="scrollDisabled" :infinite-scroll-distance="1200"
       :infinite-scroll-listen-for-event="'scrollRefresh'">

    <v-form ref="form" class="p-faces-search" lazy-validation dense @submit.prevent="updateQuery">
      <v-toolbar dense flat color="secondary-light pa-0">
        <!-- v-text-field id="search"
                      v-model="filter.q"
                      class="input-search background-inherit elevation-0"
                      solo hide-details
                      :label="$gettext('Search')"
                      prepend-inner-icon="search"
                      browser-autocomplete="off"
                      clearable overflow
                      color="secondary-dark"
                      @click:clear="clearQuery"
                      @keyup.enter.native="updateQuery"
        ></v-text-field -->
        <v-spacer></v-spacer>
        <v-divider vertical></v-divider>

        <v-btn icon overflow flat depressed color="secondary-dark" class="action-reload" :title="$gettext('Reload')" @click.stop="refresh">
          <v-icon>refresh</v-icon>
        </v-btn>
      </v-toolbar>
    </v-form>

    <v-container v-if="loading" fluid class="pa-4">
      <v-progress-linear color="secondary-dark" :indeterminate="true"></v-progress-linear>
    </v-container>
    <v-container v-else fluid class="pa-0">
      <p-scroll-top></p-scroll-top>

      <v-container grid-list-xs fluid class="pa-2">
        <v-card v-if="results.length === 0" class="no-results secondary-light lighten-1 ma-1" flat>
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
        <v-layout row wrap class="search-results face-results cards-view" :class="{'select-results': selection.length > 0}">
          <v-flex
              v-for="(model, index) in results"
              :key="index"
              xs12 sm6 md4 lg3 xl2 xxl1 d-flex
          >
            <v-card v-if="model.Marker"
                    :data-id="model.Marker.UID"
                    tile style="user-select: none;"
                    :class="model.classes()"
                    class="result accent lighten-3">
              <div class="card-background accent lighten-3"></div>
              <v-img :src="model.Marker.thumbnailUrl('tile_320')"
                     :transition="false"
                     aspect-ratio="1"
                     class="accent lighten-2">
                <v-btn v-if="!model.Marker.SubjUID && !model.Hidden" :ripple="false" :depressed="false" class="input-hide"
                       icon flat small absolute :title="$gettext('Hide')"
                       @click.stop.prevent="onHide(model)">
                  <v-icon color="white" class="action-hide">clear</v-icon>
                </v-btn>
              </v-img>

              <v-card-actions class="card-details pa-0">
                <v-layout v-if="model.Hidden" row wrap align-center>
                  <v-flex xs12 class="text-xs-center pa-0">
                    <v-btn color="transparent" :disabled="busy"
                           large depressed block :round="false"
                           class="action-undo text-xs-center"
                           :title="$gettext('Undo')" @click.stop="onShow(model)">
                      <v-icon dark>undo</v-icon>
                    </v-btn>
                  </v-flex>
                </v-layout>
                <v-layout v-else-if="model.Marker.SubjUID" row wrap align-center>
                  <v-flex xs12 class="text-xs-left pa-0">
                    <v-text-field
                        v-model="model.Marker.Name"
                        :rules="[textRule]"
                        :disabled="busy"
                        browser-autocomplete="off"
                        class="input-name pa-0 ma-0"
                        hide-details
                        single-line
                        solo-inverted
                        clearable
                        clear-icon="eject"
                        @click:clear="onClearSubject(model.Marker)"
                        @change="onRename(model.Marker)"
                        @keyup.enter.native="onRename(model.Marker)"
                    ></v-text-field>
                  </v-flex>
                </v-layout>
                <v-layout v-else row wrap align-center>
                  <v-flex xs12 class="text-xs-left pa-0">
                    <v-combobox
                        v-model="model.Marker.Name"
                        style="z-index: 250"
                        :items="$config.values.people"
                        item-value="Name"
                        item-text="Name"
                        :disabled="busy"
                        :return-object="false"
                        :menu-props="menuProps"
                        :allow-overflow="false"
                        :hint="$gettext('Name')"
                        hide-details
                        single-line
                        solo-inverted
                        open-on-clear
                        append-icon=""
                        prepend-inner-icon="person_add"
                        browser-autocomplete="off"
                        class="input-name pa-0 ma-0"
                        @change="onRename(model.Marker)"
                        @keyup.enter.native="onRename(model.Marker)"
                    >
                    </v-combobox>
                  </v-flex>
                </v-layout>
              </v-card-actions>
            </v-card>
          </v-flex>
        </v-layout>
      </v-container>
    </v-container>
  </div>
</template>

<script>
import Face from "model/face";
import Event from "pubsub-js";
import RestModel from "model/rest";
import {MaxItems} from "common/clipboard";
import Notify from "common/notify";
import {ClickLong, ClickShort, Input, InputInvalid} from "common/input";

export default {
  name: 'PPageFaces',
  props: {
    staticFilter: Object
  },
  data() {
    const query = this.$route.query;
    const routeName = this.$route.name;
    const q = query['q'] ? query['q'] : '';
    const all = query['all'] ? query['all'] : '';
    const order = this.sortOrder();
    const filter = {q: q, all: all, order: order};
    const settings = {};

    return {
      view: 'all',
      config: this.$config.values,
      subscriptions: [],
      listen: false,
      dirty: false,
      results: [],
      scrollDisabled: true,
      loading: true,
      busy: false,
      batchSize: Face.batchSize(),
      offset: 0,
      page: 0,
      selection: [],
      settings: settings,
      filter: filter,
      lastFilter: {},
      routeName: routeName,
      titleRule: v => v.length <= this.$config.get("clip") || this.$gettext("Name too long"),
      input: new Input(),
      lastId: "",
      menuProps:{"closeOnClick":false, "closeOnContentClick":true, "openOnClick":false, "maxHeight":300},
      textRule: (v) => {
        if (!v || !v.length) {
          return this.$gettext("Name");
        }

        return v.length <= this.$config.get('clip') || this.$gettext("Text too long");
      },
    };
  },
  watch: {
    '$route'() {
      const query = this.$route.query;

      this.filter.q = query["q"] ? query["q"] : "";
      this.filter.all = query["all"] ? query["all"] : "";
      this.filter.order = this.sortOrder();
      this.lastFilter = {};
      this.routeName = this.$route.name;
      this.search();
    }
  },
  created() {
    this.search();

    // this.subscriptions.push(Event.subscribe("subjects", (ev, data) => this.onUpdate(ev, data)));

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
      const offset = parseInt(window.localStorage.getItem("faces_offset"));

      if(this.offset > 0 || !offset) {
        return this.batchSize;
      }

      return offset + this.batchSize;
    },
    sortOrder() {
      return "relevance";
    },
    setOffset(offset) {
      this.offset = offset;
      window.localStorage.setItem("faces_offset", offset);
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
    onSave(m) {
      m.update();
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
    loadMore() {
      if (this.scrollDisabled) return;

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

      Face.search(params).then(resp => {
        this.results = this.dirty ? resp.models : this.results.concat(resp.models);

        this.scrollDisabled = (resp.count < resp.limit);

        if (this.scrollDisabled) {
          this.setOffset(resp.offset);
          if (this.results.length > 1) {
            this.$notify.info(this.$gettextInterpolate(this.$gettext("All %{n} faces loaded"), {n: this.results.length}));
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
      }).catch(() => {
        this.scrollDisabled = false;
      }).finally(() => {
        this.dirty = false;
        this.loading = false;
        this.listen = true;
      });
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
      if (this.loading) return;
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
        this.$nextTick(() => this.$emit("scrollRefresh"));
        return;
      }

      Object.assign(this.lastFilter, this.filter);

      this.offset = 0;
      this.page = 0;
      this.loading = true;
      this.listen = false;

      const params = this.searchParams();

      Face.search(params).then(resp => {
        this.offset = resp.limit;
        this.results = resp.models;

        this.scrollDisabled = (resp.count < resp.limit);

        if (this.scrollDisabled) {
          this.$notify.info(this.$gettextInterpolate(this.$gettext("%{n} faces found"), {n: this.results.length}));
        } else {
          this.$notify.info(this.$gettext('More than 20 faces found'));

          this.$nextTick(() => {
            if (this.$root.$el.clientHeight <= window.document.documentElement.clientHeight + 300) {
              this.$emit("scrollRefresh");
            }
          });
        }
      }).finally(() => {
        this.dirty = false;
        this.loading = false;
        this.listen = true;
      });
    },
    onShow(face) {
      this.busy = true;
      face.show().finally(() => this.busy = false);
    },
    onHide(face) {
      this.busy = true;
      face.hide().finally(() => this.busy = false);
    },
    onClearSubject(marker) {
      this.busy = true;
      marker.clearSubject(marker).finally(() => this.busy = false);
    },
    onRename(marker) {
      this.busy = true;
      marker.rename().finally(() => this.busy = false);
    },
    onUpdate(ev, data) {
      if (!this.listen) return;

      if (!data || !data.entities) {
        return;
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

            this.removeSelection(uid);
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
