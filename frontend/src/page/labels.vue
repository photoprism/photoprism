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

        <p-action-menu :items="menuActions" :tabindex="3" button-class="ms-1"></p-action-menu>
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
      <div v-if="results.length && !filter.all && !filter.q" class="d-flex justify-center my-8">
        <v-btn color="button" rounded variant="flat" @click.stop="showAll">
          {{ $gettext(`Show All Labels`) }}
        </v-btn>
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
    defaultOrder: {
      type: String,
      default: "relevance",
    },
  },
  expose: ["onShortCut"],
  data() {
    const query = this.$route.query;
    const routeName = this.$route.name;
    const order = this.sortOrder();
    const q = query["q"] ? query["q"] : "";
    const all = query["all"] ? query["all"] : "";
    const settings = {};

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
      settings: settings,
      filter: { q, order, all },
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
      restoreKey: "",
      restoreConsumed: false,
      restoreTargetCount: 0,
      restorePending: 0,
      restoring: false,
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
      this.filter.order = this.sortOrder();
      this.filter.all = query["all"] ? query["all"] : "";

      this.initRestoreState();
      this.search();
    },
  },
  created() {
    this.initRestoreState();
    this.search();

    this.subscriptions.push(this.$event.subscribe("labels", (ev, data) => this.onUpdate(ev, data)));

    this.subscriptions.push(this.$event.subscribe("touchmove.top", () => this.refresh()));
    this.subscriptions.push(this.$event.subscribe("touchmove.bottom", () => this.loadMore()));
  },
  mounted() {
    this.$view.enter(this, this.$refs?.page);
  },
  activated() {
    this.initRestoreState();

    if (this.restoring && this.restorePending > 0) {
      this.ensureRestoreTarget();
    }
  },
  beforeUnmount() {
    this.persistRestoreState();
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
          visible: this.$vuetify.display.mdAndUp,
          click: () => {
            this.refresh();
          },
        },
        {
          name: "show-all",
          icon: "mdi-eye",
          text: this.$gettext("Show All Labels"),
          visible: !this.filter.all,
          click: () => {
            this.showAll();
          },
        },
        {
          name: "show-important",
          icon: "mdi-eye-off",
          text: this.$gettext("Show Important Only"),
          visible: this.filter.all,
          click: () => {
            this.showImportant();
          },
        },
        {
          name: "sort-by-relevance",
          icon: "mdi-star",
          text: this.$gettext("Sort by Relevance"),
          visible: this.filter?.order !== "relevance",
          click: () => {
            this.updateQuery({ order: "relevance" });
          },
        },
        {
          name: "sort-by-name",
          icon: "mdi-sort-alphabetical-descending-variant",
          text: this.$gettext("Sort by Name (A–Z)"),
          visible: this.filter?.order !== "slug",
          click: () => {
            this.updateQuery({ order: "slug" });
          },
        },
        {
          name: "sort-by-count",
          icon: "mdi-sort-numeric-descending-variant",
          text: this.$gettext("Sort by Photo Count"),
          visible: this.filter?.order !== "count",
          click: () => {
            this.updateQuery({ order: "count" });
          },
        },
        {
          name: "upload",
          icon: "mdi-cloud-upload",
          text: this.$gettext("Upload") + "…",
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
    sortOrder() {
      const keyName = "labels.order";
      const queryParam = this.$route.query["order"];
      const storedOrder = window.localStorage.getItem(keyName);

      if (queryParam) {
        window.localStorage.setItem(keyName, queryParam);
        return queryParam;
      } else if (storedOrder) {
        return storedOrder;
      }

      return this.defaultOrder;
    },
    sortReverse() {
      return !!this.$route?.query["reverse"] && this.$route.query["reverse"] === "true";
    },
    searchCount() {
      if (this.restoring && this.restoreTargetCount > 0) {
        const cap = Label.restoreCap(this.batchSize);
        const desired = Math.max(this.batchSize, this.restoreTargetCount);
        const buffered = desired + this.batchSize;

        if (cap > 0) {
          return Math.min(cap, buffered);
        }

        return buffered;
      }

      const storedOffset = parseInt(window.localStorage.getItem("labels.offset"));

      if (this.offset > 0 || !Number.isFinite(storedOffset) || storedOffset <= 0) {
        return this.batchSize;
      }

      const cap = Label.restoreCap(this.batchSize);
      const total = storedOffset + this.batchSize;

      if (cap > 0) {
        return Math.min(cap, total);
      }

      return total;
    },
    setOffset(offset) {
      const value = Number.isFinite(Number(offset)) ? Number(offset) : 0;
      this.offset = value;
      window.localStorage.setItem("labels.offset", String(value));
    },
    buildRestoreKey() {
      const staticFilter = JSON.stringify(this.staticFilter) || "";
      const filter = JSON.stringify(this.filter) || "";
      const parts = [this.$route?.name || "", this.view || "", staticFilter, filter];

      return parts.join("|");
    },
    initRestoreState() {
      this.restoreKey = this.buildRestoreKey();

      if (!this.$view.wasBackwardNavigation()) {
        this.restoreConsumed = false;
        this.resetRestoreState();
        return;
      }

      if (this.restoreConsumed) {
        this.restoring = this.restorePending > 0;
        return;
      }

      const state = this.$view.consumeRestoreState(this.restoreKey);
      this.restoreConsumed = true;

      if (!state || typeof state !== "object") {
        this.resetRestoreState();
        return;
      }

      const cap = Label.restoreCap(this.batchSize);
      const count = Number(state.count);
      const offset = Number(state.offset);

      this.restoreTargetCount = Math.max(0, Number.isFinite(count) ? count : 0);

      if (cap > 0 && this.restoreTargetCount > 0) {
        this.restoreTargetCount = Math.min(cap, this.restoreTargetCount);
      }

      this.restorePending = this.restoreTargetCount;
      this.restoring = this.restorePending > 0;

      if (Number.isFinite(offset) && offset >= 0) {
        this.setOffset(offset);
      }
    },
    resetRestoreState() {
      this.restoreTargetCount = 0;
      this.restorePending = 0;
      this.restoring = false;
    },
    finishRestore() {
      if (this.restorePending > 0) {
        return;
      }

      this.restorePending = 0;
      this.restoring = false;

      window.setTimeout(() => {
        if (!this.$view.wasBackwardNavigation()) {
          this.restoreConsumed = false;
        }
      }, 0);
    },
    ensureRestoreTarget() {
      if (this.restorePending <= 0) {
        this.finishRestore();
        return;
      }

      if (this.scrollDisabled || this.$view.isHidden(this)) {
        this.finishRestore();
        return;
      }

      this.$nextTick(() => {
        if (this.restorePending > 0) {
          this.loadMore(true);
        }
      });
    },
    persistRestoreState() {
      const key = this.buildRestoreKey();

      if (!key) {
        return false;
      }

      const hasResults = Array.isArray(this.results) && this.results.length > 0;

      if (!hasResults) {
        this.$view.clearRestoreState(key);
        return false;
      }

      const scrollTop = window.scrollY ?? window.pageYOffset ?? 0;
      const offset = Number.isFinite(this.offset) && this.offset > 0 ? this.offset : this.results.length;

      return this.$view.saveRestoreState(key, {
        filterKey: key,
        count: this.results.length,
        offset: offset,
        scrollTop: Math.max(0, Math.round(scrollTop)),
      });
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
      this.$view.saveWindowScrollPos();
      this.filter.all = "true";
      if (!this.updateQuery()) {
        this.$view.clearWindowScrollPos();
      }
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
    loadMore(force = false) {
      const restoring = this.restorePending > 0;

      if (!force && (this.scrollDisabled || this.$view.isHidden(this))) {
        return;
      }

      this.scrollDisabled = true;
      this.listen = false;

      let count;
      let offset;

      if (this.dirty) {
        count = (this.page + 2) * this.batchSize;
        offset = 0;
        this.resetRestoreState();
      } else if (restoring) {
        const cap = Label.restoreCap(this.batchSize);
        const buffer = Math.min(this.batchSize, this.restorePending);
        count = Math.min(cap, this.restorePending + buffer);
        offset = this.results.length;
      } else {
        count = this.batchSize;
        offset = this.offset;
      }

      if (!Number.isFinite(count) || count <= 0) {
        count = this.batchSize;
      }

      if (!Number.isFinite(offset) || offset < 0) {
        offset = 0;
      }

      const params = {
        count: count,
        offset: offset,
      };

      Object.assign(params, this.lastFilter);

      if (this.staticFilter) {
        Object.assign(params, this.staticFilter);
      }

      if ((this.dirty || offset === 0) && !restoring) {
        this.results = [];
      }

      let shouldEnsureRestore = false;

      Label.search(params)
        .then((resp) => {
          this.results = this.dirty || offset === 0 ? resp.models : this.results.concat(resp.models);

          this.scrollDisabled = resp.count < resp.limit;

          const nextOffset = resp.offset + resp.limit;
          this.setOffset(Number.isFinite(nextOffset) ? nextOffset : this.results.length);

          if (this.scrollDisabled) {
            if (this.results.length > 1) {
              this.$notify.info(
                this.$gettextInterpolate(this.$gettext("All %{n} labels loaded"), { n: this.results.length })
              );
            }
          } else {
            this.page++;

            this.$nextTick(() => {
              if (this.$root.$el.clientHeight <= window.document.documentElement.clientHeight + 300) {
                this.loadMore();
              }
            });
          }

          if (restoring) {
            this.restorePending = Math.max(0, this.restoreTargetCount - this.results.length);
            this.restoring = this.restorePending > 0;
            shouldEnsureRestore = this.restoring && !this.scrollDisabled;

            if (!this.restoring) {
              this.finishRestore();
            }
          }

          this.$nextTick(() => this.persistRestoreState());
        })
        .catch(() => {
          this.scrollDisabled = false;
        })
        .finally(() => {
          this.dirty = false;
          this.loading = false;
          this.listen = true;

          if (shouldEnsureRestore) {
            this.ensureRestoreTarget();
          } else if (!this.restoring) {
            this.finishRestore();
          }
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

      if (this.loading) {
        return false;
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
    refresh(props) {
      this.updateSettings(props);

      // Do not refresh results if the view is already loading
      // or should not be listening for events.
      if (this.loading || !this.listen) {
        return;
      }

      this.resetRestoreState();

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
      this.resetRestoreState();
      this.$view.clearRestoreState(this.buildRestoreKey());
    },
    search() {
      /**
       * re-creating the last scroll-position should only ever happen when using
       * back-navigation. We therefore reset the remembered scroll-position
       * in any other scenario
       */
      const restoring = this.restoring && this.restoreTargetCount > 0;

      if (!restoring && !this.$view.wasBackwardNavigation()) {
        this.setOffset(0);
        this.resetRestoreState();
      }

      this.scrollDisabled = true;

      // Don't query the same data more than once
      if (!restoring && JSON.stringify(this.lastFilter) === JSON.stringify(this.filter)) {
        // this.$nextTick(() => this.$emit("scrollRefresh"));
        return;
      }

      Object.assign(this.lastFilter, this.filter);

      this.offset = 0;
      this.page = 0;
      this.loading = true;
      this.listen = false;

      const params = this.searchParams();

      let shouldEnsureRestore = false;

      Label.search(params)
        .then((resp) => {
          this.results = resp.models;

          this.scrollDisabled = resp.count < resp.limit;

          const nextOffset = resp.offset + resp.limit;
          this.setOffset(Number.isFinite(nextOffset) ? nextOffset : this.results.length);

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

          if (restoring) {
            this.restorePending = Math.max(0, this.restoreTargetCount - this.results.length);
            this.restoring = this.restorePending > 0;
            shouldEnsureRestore = this.restoring && !this.scrollDisabled;

            if (!this.restoring) {
              this.finishRestore();
            }
          } else {
            this.finishRestore();
          }

          this.$nextTick(() => this.persistRestoreState());
        })
        .catch(() => {
          this.reset();
        })
        .finally(() => {
          this.dirty = false;
          this.loading = false;
          this.listen = true;

          if (shouldEnsureRestore) {
            this.ensureRestoreTarget();
          }
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
