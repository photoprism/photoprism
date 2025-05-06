<template>
  <div ref="page" tabindex="1" class="p-page p-page-labels not-selectable" :class="$config.aclClasses('labels')">
    <v-form
      ref="form"
      validate-on="invalid-input"
      class="p-labels-search p-page__navigation"
      @submit.stop.prevent="updateQuery()"
    >
      <v-toolbar
        flat
        :density="$vuetify.display.smAndDown ? 'compact' : 'default'"
        color="secondary"
        class="page-toolbar"
      >
        <v-text-field
          :model-value="filter.q"
          hide-details
          clearable
          overflow
          single-line
          rounded
          tabindex="1"
          variant="solo-filled"
          :density="density"
          validate-on="invalid-input"
          :placeholder="$gettext('Search')"
          prepend-inner-icon="mdi-magnify"
          autocomplete="off"
          autocorrect="off"
          autocapitalize="none"
          color="surface-variant"
          class="input-search input-search--focus background-inherit elevation-0"
          @update:model-value="
            (v) => {
              updateFilter({ q: v });
            }
          "
          @keyup.enter="() => updateQuery()"
          @click:clear="
            () => {
              updateQuery({ q: '' });
            }
          "
        ></v-text-field>

        <v-btn
          v-if="!filter.all"
          icon="mdi-eye"
          tabindex="2"
          :title="$gettext('Show more')"
          class="action-show-all ms-1"
          @click.stop="showAll"
        >
        </v-btn>
        <v-btn
          v-else
          icon="mdi-eye-off"
          tabindex="2"
          :title="$gettext('Show less')"
          class="action-show-important ms-1"
          @click.stop="showImportant"
        >
        </v-btn>
        <p-action-menu
          v-if="$vuetify.display.mdAndUp"
          :items="menuActions"
          :tabindex="3"
          button-class="ms-1"
        ></p-action-menu>
      </v-toolbar>
    </v-form>

    <div v-if="loading" class="p-page__loading">
      <p-loading></p-loading>
    </div>
    <div v-else class="p-page__content">
      <p-label-clipboard
        v-if="canSelect"
        :refresh="refresh"
        :selection="selection"
        :clear-selection="clearSelection"
      ></p-label-clipboard>

      <p-scroll
        :load-more="loadMore"
        :load-disabled="scrollDisabled"
        :load-distance="scrollDistance"
        :loading="loading"
      ></p-scroll>

      <div v-if="results.length === 0" class="pa-3">
        <v-alert color="surface-variant" icon="mdi-lightbulb-outline" class="no-results" variant="outlined">
          <div class="font-weight-bold">
            {{ $gettext(`No labels found`) }}
          </div>
          <div class="mt-2">
            {{ $gettext(`Try again using other filters or keywords.`) }}
            {{
              $gettext(
                `In case pictures you expect are missing, please rescan your library and wait until indexing has been completed.`
              )
            }}
          </div>
        </v-alert>
      </div>
      <div
        v-else
        class="v-row search-results label-results cards-view"
        :class="{ 'select-results': selection.length > 0 }"
      >
        <div
          v-for="(label, index) in results"
          :key="label.UID"
          ref="items"
          class="v-col-6 v-col-sm-4 v-col-md-3 v-col-xl-2"
        >
          <div
            :data-uid="label.UID"
            class="result not-selectable"
            :class="label.classes(selection.includes(label.UID))"
            @click="$router.push(label.route(view))"
            @contextmenu.stop="onContextMenu($event, index)"
          >
            <div
              :title="label.Name"
              :style="`background-image: url(${label.thumbnailUrl('tile_500')})`"
              class="preview"
              @touchstart.passive="input.touchStart($event, index)"
              @touchend.stop="onClick($event, index)"
              @mousedown.stop.prevent="input.mouseDown($event, index)"
              @click.stop.prevent="onClick($event, index)"
            >
              <div class="preview__overlay"></div>
              <button
                v-if="canSelect"
                class="input-select"
                @touchstart.stop="input.touchStart($event, index)"
                @touchend.stop="onSelect($event, index)"
                @touchmove.stop.prevent
                @click.stop.prevent="onSelect($event, index)"
              >
                <i class="mdi mdi-check-circle select-on" />
                <i class="mdi mdi-circle-outline select-off" />
              </button>
              <button
                class="input-favorite"
                @touchstart.stop="input.touchStart($event, index)"
                @touchend.stop="toggleLike($event, index)"
                @touchmove.stop.prevent
                @click.stop.prevent="toggleLike($event, index)"
              >
                <i v-if="label.Favorite" class="mdi mdi-star text-favorite" />
                <i v-else class="mdi mdi-star-outline" />
              </button>
            </div>

            <div class="meta" @click.stop.prevent="">
              <div v-if="canManage" class="meta-title inline-edit clickable" @click.stop.prevent="edit(label)">
                {{ label.Name }}
              </div>
              <div v-else class="meta-title">
                {{ label.Name }}
              </div>

              <div v-if="label.PhotoCount === 1" class="meta-count" @click.stop.prevent="">
                {{ $gettext(`Contains one picture.`) }}
              </div>
              <div v-else-if="label.PhotoCount > 0" class="meta-count" @click.stop.prevent="">
                {{ $gettext(`Contains %{n} pictures.`, { n: label.PhotoCount }) }}
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>

    <p-label-edit-dialog :visible="dialog.edit" :label="model" @close="dialog.edit = false"></p-label-edit-dialog>
  </div>
</template>

<script>
import Label from "model/label";
import RestModel from "model/rest";
import { MaxItems } from "common/clipboard";
import $notify from "common/notify";
import { Input, InputInvalid, ClickShort, ClickLong } from "common/input";

import PLoading from "component/loading.vue";
import PActionMenu from "component/action/menu.vue";

export default {
  name: "PPageLabels",
  components: {
    PLoading,
    PActionMenu,
  },
  props: {
    staticFilter: {
      type: Object,
      default: () => {},
    },
  },
  expose: ["onShortCut"],
  data() {
    const query = this.$route.query;
    const routeName = this.$route.name;
    const q = query["q"] ? query["q"] : "";
    const all = query["all"] ? query["all"] : "";

    const features = this.$config.getSettings().features;
    const canManage = this.$config.allow("labels", "manage");
    const canAddAlbums = this.$config.allow("albums", "create") && features.albums;

    return {
      canManage: canManage,
      canUpload: this.$config.allow("files", "upload") && features.upload,
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
      labelToRename: "",
      dialog: {
        edit: false,
      },
      model: new Label(false),
    };
  },
  computed: {
    density() {
      return this.$vuetify.display.smAndDown ? "compact" : "comfortable";
    },
  },
  watch: {
    $route() {
      if (!this.$view.isActive(this)) {
        return;
      }

      this.$view.focus(this.$refs?.page);

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

    this.subscriptions.push(this.$event.subscribe("labels", (ev, data) => this.onUpdate(ev, data)));

    this.subscriptions.push(this.$event.subscribe("touchmove.top", () => this.refresh()));
    this.subscriptions.push(this.$event.subscribe("touchmove.bottom", () => this.loadMore()));
  },
  mounted() {
    this.$view.enter(this, this.$refs?.page);
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
    menuActions() {
      return [
        {
          name: "refresh",
          icon: "mdi-refresh",
          text: this.$gettext("Refresh"),
          shortcut: "Ctrl-R",
          visible: true,
          click: () => {
            this.refresh();
          },
        },
        {
          name: "upload",
          icon: "mdi-cloud-upload",
          text: this.$gettext("Upload"),
          shortcut: "Ctrl-U",
          visible: this.canUpload,
          click: () => {
            this.$event.publish("dialog.upload");
          },
        },
        {
          name: "docs",
          icon: "mdi-book-open-page-variant-outline",
          text: this.$gettext("Learn More"),
          visible: true,
          href: "https://docs.photoprism.app/user-guide/organize/labels/",
          target: "_blank",
        },
      ];
    },
    onShortCut(ev) {
      switch (ev.code) {
        case "KeyR":
          this.refresh();
          return true;
        case "KeyF":
          this.$view.focus(this.$refs?.form, ".input-search input", true);
          return true;
        case "KeyU":
          if (this.$config.allow("files", "upload") && this.$config.feature("upload")) {
            this.$event.publish("dialog.upload");
          }
          return true;
      }
    },
    edit(label) {
      if (!label) {
        return;
      } else if (!this.canManage) {
        this.$router.push(label.route(this.view));
        return;
      }

      this.model = label;
      this.dialog.edit = true;
    },
    searchCount() {
      const offset = parseInt(window.localStorage.getItem("labels.offset"));

      if (this.offset > 0 || !offset) {
        return this.batchSize;
      }

      return offset + this.batchSize;
    },
    setOffset(offset) {
      this.offset = offset;
      window.localStorage.setItem("labels.offset", offset);
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
      label.Name = this.labelToRename;
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
          $notify.warn(this.$gettext("Can't select more items"));
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
      if (this.scrollDisabled || this.$view.isHidden(this)) return;

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
              this.$notify.info(
                this.$gettextInterpolate(this.$gettext("All %{n} labels loaded"), { n: this.results.length })
              );
            }
          } else {
            this.setOffset(resp.offset + resp.limit);
            this.page++;

            this.$nextTick(() => {
              if (this.$root.$el.clientHeight <= window.document.documentElement.clientHeight + 300) {
                this.loadMore();
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

        window.localStorage.setItem("labels." + key, this.settings[key]);
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

      // Do not refresh results if the view is already loading
      // or should not be listening for events.
      if (this.loading || !this.listen) {
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
        // this.$nextTick(() => this.$emit("scrollRefresh"));
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
              this.$notify.info(
                this.$gettextInterpolate(this.$gettext("%{n} labels found"), { n: this.results.length })
              );
            }
          } else {
            // this.$notify.info(this.$gettext('More than 20 labels found'));
            this.$nextTick(() => {
              if (this.$root.$el.clientHeight <= window.document.documentElement.clientHeight + 300) {
                this.loadMore();
              }
            });
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
  },
};
</script>
