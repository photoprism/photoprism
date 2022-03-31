<template>
  <div v-infinite-scroll="loadMore" class="p-page p-page-albums" style="user-select: none"
       :infinite-scroll-disabled="scrollDisabled" :infinite-scroll-distance="scrollDistance"
       :infinite-scroll-listen-for-event="'scrollRefresh'">
    <v-toolbar flat color="secondary" :dense="$vuetify.breakpoint.smAndDown">
      <v-toolbar-title>
        <translate>Albums</translate>
      </v-toolbar-title>
    </v-toolbar>
    <v-container v-if="loading" fluid class="pa-4">
      <v-progress-linear color="secondary-dark" :indeterminate="true"></v-progress-linear>
    </v-container>
    <v-container v-else fluid class="pa-0">
      <p-scroll-top></p-scroll-top>

      <p-album-clipboard :refresh="refresh" :selection="selection"
                         :clear-selection="clearSelection" :context="context"></p-album-clipboard>

      <v-container grid-list-xs fluid class="pa-2">
        <v-alert
            :value="results.length === 0"
            color="secondary-dark" icon="bookmark" class="no-results ma-2 opacity-70" outline
        >
          <h3 class="body-2 ma-0 pa-0">
            <translate>No albums found</translate>
          </h3>
          <p class="body-1 mt-2 mb-0 pa-0">
            <translate>Try again using other filters or keywords.</translate>
          </p>
        </v-alert>
        <v-layout row wrap class="search-results album-results cards-view" :class="{'select-results': selection.length > 0}">
          <v-flex
              v-for="(album, index) in results"
              :key="album.UID"
              xs6 sm4 md3 xlg2 xxl1 d-flex
          >
            <v-card tile
                    :data-uid="album.UID"
                    style="user-select: none"
                    class="result accent lighten-3"
                    :class="album.classes(selection.includes(album.UID))"
                    :to="album.route(view)"
                    @contextmenu.stop="onContextMenu($event, index)"
            >
              <div class="card-background accent lighten-3" style="user-select: none"></div>
              <v-img
                  :src="album.thumbnailUrl('tile_500')"
                  :alt="album.Title"
                  :transition="false"
                  aspect-ratio="1"
                  style="user-select: none"
                  class="accent lighten-2 clickable"
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

              <v-card-title primary-title class="pl-3 pt-3 pr-3 pb-2 card-details" style="user-select: none;">
                <div>
                  <h3 v-if="album.Type !== 'month'" class="body-2 mb-0" :title="album.Title">
                    {{ album.Title | truncate(80) }}
                  </h3>
                  <h3 v-else class="body-2 mb-0">
                    {{ album.getDateString() | capitalize }}
                  </h3>
                </div>
              </v-card-title>

              <v-card-text class="pb-2 pt-0  card-details">
                <div v-if="album.Description" class="caption mb-2" :title="$gettext('Description')">
                  {{ album.Description }}
                </div>
                <div v-else class="caption mb-2">
                  <translate>Shared with you.</translate>
                </div>
                <div v-if="album.Category !== ''" class="caption mb-2 d-inline-block">
                  <button @click.stop="">
                    <v-icon size="14">local_offer</v-icon>
                    {{ album.Category }}
                  </button>
                </div>
                <div v-if="album.getLocation() !== ''" class="caption mb-2 d-inline-block">
                  <button @click.stop="">
                    <v-icon size="14">location_on</v-icon>
                    {{ album.getLocation() }}
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
import Album from "model/album";
import {DateTime} from "luxon";
import Event from "pubsub-js";
import RestModel from "model/rest";
import {MaxItems} from "common/clipboard";
import Notify from "common/notify";
import {Input, InputInvalid, ClickShort, ClickLong} from "common/input";

export default {
  name: 'PPageAlbums',
  props: {
    staticFilter: {
      type: Object,
      default: () => {},
    },
    view: String,
  },
  data() {
    const query = this.$route.query;
    const routeName = this.$route.name;
    const q = query["q"] ? query["q"] : "";
    const category = query["category"] ? query["category"] : "";
    const filter = {q, category};
    const settings = {};

    let categories = [{"value": "", "text": this.$gettext("All Categories")}];

    if (this.$config.albumCategories().length > 0) {
      categories = categories.concat(this.$config.albumCategories().map(cat => {
        return {"value": cat, "text": cat};
      }));
    }

    return {
      site: this.$config.page,
      categories: categories,
      subscriptions: [],
      listen: false,
      dirty: false,
      results: [],
      loading: true,
      scrollDisabled: true,
      scrollDistance: window.innerHeight*2,
      batchSize: Album.batchSize(),
      offset: 0,
      page: 0,
      selection: [],
      settings: settings,
      filter: filter,
      lastFilter: {},
      routeName: routeName,
      titleRule: v => v.length <= this.$config.get('clip') || this.$gettext("Title too long"),
      input: new Input(),
      lastId: "",
      model: new Album(),
    };
  },
  computed: {
    context: function () {
      if (!this.staticFilter) {
        return "album";
      }

      if (this.staticFilter.type) {
        return this.staticFilter.type;
      }

      return "";
    }
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
  methods: {
    searchCount() {
      const offset = parseInt(window.localStorage.getItem("share_albums_offset"));

      if(this.offset > 0 || !offset) {
        return this.batchSize;
      }

      return offset + this.batchSize;
    },
    setOffset(offset) {
      this.offset = offset;
      window.localStorage.setItem("share_albums_offset", offset);
    },
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
    clearQuery() {
      this.filter.q = '';
      this.search();
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

      Album.search(params).then(resp => {
        this.results = this.dirty ? resp.models : this.results.concat(resp.models);

        this.scrollDisabled = (resp.count < resp.limit);

        if (this.scrollDisabled) {
          this.setOffset(resp.offset);

          if (this.results.length > 1) {
            this.$notify.info(this.$gettextInterpolate(this.$gettext("All %{n} albums loaded"), {n: this.results.length}));
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

      Album.search(params).then(resp => {
        this.offset = resp.limit;
        this.results = resp.models;

        this.scrollDisabled = (resp.count < resp.limit);

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
        title = `${title} (${existing.length + 1})`;
      }

      const album = new Album({"Title": title, "Favorite": false});

      album.save();
    },
    onSave(album) {
      album.update();
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

            if (model) {
              for (let key in values) {
                if (values.hasOwnProperty(key) && values[key] != null && typeof values[key] !== "object") {
                  model[key] = values[key];
                }
              }
            }
          }

          let categories = [{"value": "", "text": this.$gettext("All Categories")}];

          if (this.$config.albumCategories().length > 0) {
            categories = categories.concat(this.$config.albumCategories().map(cat => {
              return {"value": cat, "text": cat};
            }));
          }

          this.categories = categories;

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
};
</script>
