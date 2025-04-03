<template>
  <div class="p-page p-page-subjects not-selectable">
    <v-form ref="form" validate-on="invalid-input" class="p-people-search" @submit.prevent="updateQuery()">
      <v-toolbar density="compact" class="page-toolbar" color="secondary-light">
        <v-text-field
          v-if="canSearch"
          :model-value="filter.q"
          hide-details
          clearable
          single-line
          overflow
          rounded
          validate-on="invalid-input"
          :placeholder="$gettext('Search')"
          prepend-inner-icon="mdi-magnify"
          autocomplete="off"
          autocorrect="off"
          autocapitalize="none"
          color="surface-variant"
          density="compact"
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

        <v-btn :title="$gettext('Refresh')" icon="mdi-refresh" class="action-reload" @click.stop="refresh"></v-btn>

        <template v-if="canManage">
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
            @click.stop="onExcludeHidden()"
          >
          </v-btn>
        </template>
      </v-toolbar>
    </v-form>

    <div v-if="loading" class="p-page__loading">
      <p-loading></p-loading>
    </div>
    <div v-else style="min-height: 100vh" class="p-page__content">
      <p-people-clipboard
        :refresh="refresh"
        :selection="selection"
        :clear-selection="clearSelection"
      ></p-people-clipboard>

      <p-scroll
        :load-more="loadMore"
        :load-disabled="scrollDisabled"
        :load-distance="scrollDistance"
        :loading="loading"
      ></p-scroll>

      <div v-if="results.length === 0" class="pa-3">
        <v-alert color="surface-variant" icon="mdi-lightbulb-outline" class="no-results" variant="outlined">
          <div class="font-weight-bold">
            {{ $gettext(`No people found`) }}
          </div>
          <div class="mt-2">
            {{ $gettext(`Try again using other filters or keywords.`) }}
            {{ $gettext(`You may rescan your library to find additional faces.`) }}
            {{ $gettext(`Recognition starts after indexing has been completed.`) }}
          </div>
        </v-alert>
      </div>
      <div
        v-else
        class="v-row search-results subject-results cards-view"
        :class="{ 'select-results': selection.length > 0 }"
      >
        <div v-for="(m, index) in results" :key="m.UID" class="v-col-6 v-col-sm-4 v-col-md-3 v-col-xl-2">
          <div
            :data-uid="m.UID"
            class="result not-selectable"
            :class="m.classes(selection.includes(m.UID))"
            @contextmenu.stop="onContextMenu($event, index)"
          >
            <v-img
              :src="m.thumbnailUrl('tile_320')"
              :alt="m.Name"
              aspect-ratio="1"
              class="preview not-selectable"
              @touchstart.passive="input.touchStart($event, index)"
              @touchend.stop="onClick($event, index)"
              @mousedown.stop.prevent="input.mouseDown($event, index)"
              @click.stop.prevent="onClick($event, index)"
            >
              <v-btn
                v-if="canManage"
                :ripple="false"
                class="input-hidden"
                icon
                variant="text"
                density="comfortable"
                position="absolute"
                @touchstart.stop="input.touchStart($event, index)"
                @touchend.stop="onToggleHidden($event, index)"
                @touchmove.stop.prevent
                @click.stop.prevent="onToggleHidden($event, index)"
              >
                <v-icon color="white" class="select-on" :title="$gettext('Show')">mdi-eye-off</v-icon>
                <v-icon color="white" class="select-off" :title="$gettext('Hide')">mdi-close</v-icon>
              </v-btn>
              <v-btn
                :ripple="false"
                icon
                variant="text"
                position="absolute"
                class="input-select"
                @touchstart.stop="input.touchStart($event, index)"
                @touchend.stop="onSelect($event, index)"
                @touchmove.stop.prevent
                @click.stop.prevent="onSelect($event, index)"
              >
                <v-icon color="white" class="select-on">mdi-check-circle</v-icon>
                <v-icon color="white" class="select-off">mdi-radiobox-blank</v-icon>
              </v-btn>

              <v-btn
                :ripple="false"
                icon
                variant="text"
                position="absolute"
                class="input-favorite"
                @touchstart.stop="input.touchStart($event, index)"
                @touchend.stop="toggleLike($event, index)"
                @touchmove.stop.prevent
                @click.stop.prevent="toggleLike($event, index)"
              >
                <v-icon icon="mdi-star" color="favorite" class="select-on"></v-icon>
                <v-icon icon="mdi-star-outline" color="white" class="select-off"></v-icon>
              </v-btn>
            </v-img>

            <div class="meta" @click.stop.prevent="">
              <div v-if="canManage" class="meta-title inline-edit clickable" @click.stop.prevent="edit(m)">
                {{ m.Name }}
              </div>
              <div v-else class="meta-title">
                {{ m.Name }}
              </div>

              <div v-if="m.About" class="meta-about text-truncate" :title="$gettext('About')">
                {{ m.About }}
              </div>

              <div v-if="m.PhotoCount === 1" class="meta-count" @click.stop.prevent="">
                {{ $gettext(`Contains one picture.`) }}
              </div>
              <div v-else-if="m.PhotoCount > 0" class="meta-count" @click.stop.prevent="">
                {{ $gettext(`Contains %{n} pictures.`, { n: m.PhotoCount }) }}
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>

    <p-people-edit-dialog
      :visible="dialog.edit"
      :person="model"
      @close="dialog.edit = false"
      @confirm="onSave"
    ></p-people-edit-dialog>
    <p-people-merge-dialog
      :visible="merge.visible"
      :subj1="merge.subj1"
      :subj2="merge.subj2"
      @close="onCancelMerge"
      @confirm="onMerge"
    ></p-people-merge-dialog>
  </div>
</template>

<script>
import Subject from "model/subject";
import RestModel from "model/rest";
import { MaxItems } from "common/clipboard";
import $notify from "common/notify";
import { ClickLong, ClickShort, Input, InputInvalid } from "common/input";
import PLoading from "component/loading.vue";

export default {
  name: "PPageSubjects",
  components: { PLoading },
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
        visible: false,
        subj1: null,
        subj2: null,
      },
      dialog: {
        edit: false,
      },
      model: new Subject(false),
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

    this.subscriptions.push(this.$event.subscribe("subjects", (ev, data) => this.onUpdate(ev, data)));

    this.subscriptions.push(this.$event.subscribe("touchmove.top", () => this.refresh()));
    this.subscriptions.push(this.$event.subscribe("touchmove.bottom", () => this.loadMore()));
  },
  beforeUnmount() {
    for (let i = 0; i < this.subscriptions.length; i++) {
      this.$event.unsubscribe(this.subscriptions[i]);
    }
  },
  methods: {
    edit(subject) {
      if (!subject) {
        return;
      } else if (!this.canManage) {
        this.$router.push(subject.route(this.view));
        return;
      }

      this.model = subject;
      this.dialog.edit = true;
    },
    onSave(m) {
      if (!this.canManage || !m.Name || m.Name.trim() === "") {
        // Refuse to save empty name.
        return;
      }

      const existing = this.$config.getPerson(m.Name);
      if (!existing || existing.UID === m.UID) {
        this.busy = true;
        m.update()
          .then((m) => {
            this.$notify.success(this.$gettext("Changes successfully saved"));
            this.dialog.edit = false;
          })
          .finally(() => {
            this.busy = false;
            this.dialog.edit = false;
          });
      } else {
        this.merge.subj1 = m;
        this.merge.subj2 = existing;
        this.dialog.edit = false;
        this.merge.visible = true;
      }
    },
    onCancelMerge() {
      this.merge.visible = false;
      this.merge.subj1 = null;
      this.merge.subj2 = null;
    },
    onMerge() {
      if (!this.canManage) {
        return;
      }

      this.busy = true;
      this.merge.visible = false;
      this.dialog.edit = false;
      this.$notify.blockUI("busy");
      this.merge.subj1.update().finally(() => {
        this.busy = false;
        this.merge.subj1 = null;
        this.merge.subj2 = null;
        this.$notify.unblockUI();
        this.refresh();
      });
    },
    searchCount() {
      const offset = parseInt(window.localStorage.getItem("people.recognized.offset"));

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
      window.localStorage.setItem("people.recognized.offset", offset);
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
              this.$notify.info(
                this.$gettextInterpolate(this.$gettext("All %{n} people loaded"), { n: this.results.length })
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

        window.localStorage.setItem("people.recognized." + key, this.settings[key]);
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

      if (this.loading || !this.active || !this.listen) {
        return;
      }

      /*
      TODO: Leaving "loading" untouched here avoids flickering when refreshing the results, which might lead to a
       smoother experience. If it doesn't cause any problems or unwanted side effects, this line can be removed.

      this.loading = true;
      */
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
        // this.$nextTick(() => this.$emit("scrollRefresh"));
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
              this.$notify.info(
                this.$gettextInterpolate(this.$gettext("%{n} people found"), { n: this.results.length })
              );
            }
          } else {
            // this.$notify.info(this.$gettext('More than 20 people found'));
            this.$nextTick(() => {
              if (this.$root.$el.clientHeight <= window.document.documentElement.clientHeight + 300) {
                this.loadMore();
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
