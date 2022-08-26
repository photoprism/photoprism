<template>
  <div v-infinite-scroll="loadMore" class="p-page p-page-albums" style="user-select: none"
       :infinite-scroll-disabled="scrollDisabled" :infinite-scroll-distance="scrollDistance"
       :infinite-scroll-listen-for-event="'scrollRefresh'">

    <v-form ref="form" class="p-albums-search" lazy-validation dense @submit.prevent="updateQuery()">
      <v-toolbar flat :dense="$vuetify.breakpoint.smAndDown" class="page-toolbar" color="secondary">
        <v-text-field :value="filter.q"
                      solo hide-details clearable overflow single-line validate-on-blur
                      class="input-search background-inherit elevation-0"
                      :label="$gettext('Search')"
                      browser-autocomplete="off"
                      autocorrect="off"
                      autocapitalize="none"
                      prepend-inner-icon="search"
                      color="secondary-dark"
                      @change="(v) => {updateFilter({'q': v})}"
                      @keyup.enter.native="(e) => updateQuery({'q': e.target.value})"
                      @click:clear="() => {updateQuery({'q': ''})}"
        ></v-text-field>

        <v-overflow-btn :value="filter.category"
                  solo hide-details single-line
                  :label="$gettext('Category')"
                  color="secondary-dark"
                  background-color="secondary"
                  prepend-icon="local_offer"
                  append-icon=""
                  :items="categories"
                  class="hidden-xs-only input-category background-inherit elevation-0"
                  @change="(v) => {updateQuery({'category': v})}"
        >
        </v-overflow-btn>

        <v-btn icon class="action-reload" :title="$gettext('Reload')" @click.stop="refresh()">
          <v-icon>refresh</v-icon>
        </v-btn>

        <v-btn v-if="!$config.values.readonly && $config.feature('upload')" icon class="hidden-sm-and-down action-upload"
               :title="$gettext('Upload')" @click.stop="showUpload()">
          <v-icon>cloud_upload</v-icon>
        </v-btn>

        <v-btn v-if="staticFilter.type === 'album'" icon class="action-add" :title="$gettext('Add Album')"
               @click.prevent="create()">
          <v-icon>add</v-icon>
        </v-btn>
      </v-toolbar>
    </v-form>

    <v-container v-if="loading" fluid class="pa-4">
      <v-progress-linear color="secondary-dark" :indeterminate="true"></v-progress-linear>
    </v-container>
    <v-container v-else fluid class="pa-0">
      <p-scroll-top></p-scroll-top>

      <p-album-clipboard :refresh="refresh" :selection="selection" :share="share" :edit="edit"
                         :clear-selection="clearSelection" :context="context"></p-album-clipboard>

      <v-container grid-list-xs fluid class="pa-2">
        <v-alert
            :value="results.length === 0"
            color="secondary-dark" icon="lightbulb_outline" class="no-results ma-2 opacity-70" outline
        >
          <h3 class="body-2 ma-0 pa-0">
            <translate>No albums found</translate>
          </h3>
          <p class="body-1 mt-2 mb-0 pa-0">
            <translate>Try again using other filters or keywords.</translate>
            <template v-if="staticFilter.type === 'album'">
              <translate>After selecting pictures from search results, you can add them to an album using the context menu.</translate>
            </template>
            <template v-else>
              <translate>Your library is continuously analyzed to automatically create albums of special moments, trips, and places.</translate>
            </template>
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
                <v-btn v-if="featureShare && album.LinkCount > 0" :ripple="false"
                       icon flat absolute
                       class="action-share"
                       @touchstart.stop.prevent="input.touchStart($event, index)"
                       @touchend.stop.prevent="onShare($event, index)"
                       @touchmove.stop.prevent
                       @click.stop.prevent="onShare($event, index)">
                  <v-icon color="white">share</v-icon>
                </v-btn>

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

                <v-btn :ripple="false"
                       icon flat absolute
                       class="input-favorite"
                       @touchstart.stop.prevent="input.touchStart($event, index)"
                       @touchend.stop.prevent="toggleLike($event, index)"
                       @touchmove.stop.prevent
                       @click.stop.prevent="toggleLike($event, index)">
                  <v-icon color="#FFD600" class="select-on">star</v-icon>
                  <v-icon color="white" class="select-off">star_border</v-icon>
                </v-btn>
              </v-img>

              <v-card-title primary-title class="pl-3 pt-3 pr-3 pb-2 card-details" style="user-select: none;">
                <div>
                  <h3 class="body-2 mb-0">
                    <button v-if="album.Type !== 'month'" class="action-title-edit" :data-uid="album.UID"
                            @click.stop.prevent="edit(album)">
                      {{ album.Title | truncate(80) }}
                    </button>
                    <button v-else class="action-title-edit" :data-uid="album.UID"
                    @click.stop.prevent="edit(album)">
                      {{ album.getDateString() | capitalize }}
                    </button>
                  </h3>
                </div>
              </v-card-title>

              <v-card-text primary-title class="pb-2 pt-0 card-details" style="user-select: none;"
                           @click.stop.prevent="">
                <div v-if="album.Description" class="caption mb-2" :title="$gettext('Description')">
                  <button @click.exact="edit(album)">
                    {{ album.Description | truncate(100) }}
                  </button>
                </div>

                <div v-else-if="album.Type === 'album'" class="caption mb-2">
                  <button v-if="album.PhotoCount === 1" @click.exact="edit(album)">
                    <translate>Contains one picture.</translate>
                  </button>
                  <button v-else-if="album.PhotoCount > 0">
                    <translate :translate-params="{n: album.PhotoCount}">Contains %{n} pictures.</translate>
                  </button>
                  <button v-else @click.stop.prevent="$router.push({name: 'browse'})">
                    <translate>Add pictures from search results by selecting them.</translate>
                  </button>
                </div>
                <div v-else-if="album.Type === 'folder'" class="caption mb-2">
                  <button @click.exact="edit(album)">
                    /{{ album.Path | truncate(100) }}
                  </button>
                </div>
                <div v-if="album.Category !== ''" class="caption mb-2 d-inline-block">
                  <button @click.exact="edit(album)">
                    <v-icon size="14">local_offer</v-icon>
                    {{ album.Category }}
                  </button>
                </div>
                <div v-if="album.getLocation() !== ''" class="caption mb-2 d-inline-block">
                  <button @click.exact="edit(album)">
                    <v-icon size="14">location_on</v-icon>
                    {{ album.getLocation() }}
                  </button>
                </div>
              </v-card-text>
            </v-card>
          </v-flex>
        </v-layout>
        <div v-if="staticFilter.type === 'album' && config.count.albums === 0" class="text-xs-center my-2">
          <v-btn class="action-add" color="secondary" round @click.prevent="create">
            <translate>Add Album</translate>
          </v-btn>
        </div>
      </v-container>
    </v-container>
    <p-share-dialog :show="dialog.share" :model="model" @upload="webdavUpload"
                    @close="dialog.share = false"></p-share-dialog>
    <p-share-upload-dialog :show="dialog.upload" :items="{albums: selection}" :model="model" @cancel="dialog.upload = false"
                           @confirm="dialog.upload = false"></p-share-upload-dialog>
    <p-album-edit-dialog :show="dialog.edit" :album="model" @close="dialog.edit = false"></p-album-edit-dialog>
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
    view: {
      type: String,
      default: "",
    },
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
      config: this.$config.values,
      featureShare: this.$config.feature('share'),
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
      q: q,
      filter: filter,
      lastFilter: {},
      routeName: routeName,
      titleRule: v => v.length <= this.$config.get('clip') || this.$gettext("Title too long"),
      input: new Input(),
      lastId: "",
      dialog: {
        share: false,
        upload: false,
        edit: false,
      },
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

      this.routeName = this.$route.name;
      this.lastFilter = {};
      this.q = query["q"] ? query["q"] : "";
      this.filter.q = this.q;
      this.filter.category = query["category"] ? query["category"] : "";

      this.search();
    }
  },
  created() {
    if (!window.popStateDetected) {
      this.setOffset(0);
    }
    this.search();

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
      const offset = parseInt(window.localStorage.getItem("albums_offset"));

      if(this.offset > 0 || !offset) {
        return this.batchSize;
      }

      return offset + this.batchSize;
    },
    setOffset(offset) {
      this.offset = offset;
      window.localStorage.setItem("albums_offset", offset);
    },
    share(album) {
      if (!album) {
        return;
      }

      this.model = album;
      this.dialog.share = true;
    },
    edit(album) {
      if (!album) {
        return;
      }

      this.model = album;
      this.dialog.edit = true;
    },
    webdavUpload() {
      this.dialog.share = false;
      this.dialog.upload = true;
    },
    showUpload() {
      Event.publish("dialog.upload");
    },
    toggleLike(ev, index) {
      const inputType = this.input.eval(ev, index);

      if (inputType !== ClickShort) {
        return;
      }

      const album = this.results[index];

      if (!album) {
        return;
      }

      album.toggleLike();
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
    onShare(ev, index) {
      const inputType = this.input.eval(ev, index);

      if (inputType !== ClickShort) {
        return;
      }

      return this.share(this.results[index]);
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

        window.localStorage.setItem("albums_"+key, this.settings[key]);
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
    refresh(props) {
      this.updateSettings(props);

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
      if (!this.listen) {
        console.log("albums.onUpdate currently not listening", ev, data);
        return;
      } else if (!data || !data.entities || !Array.isArray(data.entities)) {
        console.log("albums.onUpdate received empty data", ev, data);
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
