<template>
  <div ref="page" tabindex="-1" class="p-page p-page-errors">
    <v-form
      ref="form"
      validate-on="invalid-input"
      class="p-errors-search p-page__navigation"
      @submit.prevent="updateQuery()"
    >
      <v-toolbar
        flat
        :density="$vuetify.display.smAndDown ? 'compact' : 'default'"
        class="page-toolbar"
        color="secondary"
      >
        <v-text-field
          :model-value="filter.q"
          hide-details
          clearable
          overflow
          single-line
          rounded
          variant="solo-filled"
          :density="density"
          validate-on="invalid-input"
          autocomplete="off"
          autocorrect="off"
          autocapitalize="none"
          :placeholder="$gettext('Search')"
          prepend-inner-icon="mdi-magnify"
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
          v-if="!isPublic"
          :title="$gettext('Delete All')"
          icon="mdi-delete-sweep"
          class="action-delete action-delete-all ms-1"
          @click.stop="onDelete"
        >
        </v-btn>
        <p-action-menu
          v-if="$vuetify.display.mdAndUp"
          :items="menuActions"
          button-class="ms-1"
        ></p-action-menu>
      </v-toolbar>
    </v-form>
    <div v-if="loading" class="p-page__loading">
      <p-loading></p-loading>
    </div>
    <div v-else-if="errors.length > 0" fluid class="pa-0">
      <p-scroll
        :load-more="loadMore"
        :load-disabled="scrollDisabled"
        :load-distance="scrollDistance"
        :loading="loading"
      ></p-scroll>

      <v-list lines="one" bg-color="table" density="compact" class="py-0">
        <v-list-item
          v-for="err in errors"
          :key="err.ID"
          :prepend-icon="err.Level === 'error' ? 'mdi-alert-circle-outline' : 'mdi-alert'"
          density="default"
          :title="err.Message"
          :subtitle="localTime(err.Time)"
          class="py-2"
          @click="showDetails(err)"
        >
          <template #prepend>
            <v-icon v-if="err.Level === 'error'" icon="mdi-alert-circle-outline" color="error"></v-icon>
            <v-icon v-else-if="err.Level === 'warning'" icon="mdi-alert" color="warning"></v-icon>
            <v-icon v-else icon="mdi-information-outline" color="info"></v-icon>
          </template>
          <template #title="{ title }">
            <div class="text-body-2 text-truncate">{{ title }}</div>
          </template>
        </v-list-item>
      </v-list>
    </div>
    <div v-else class="pa-3">
      <v-alert color="primary" icon="mdi-check-circle-outline" class="no-results" variant="outlined">
        <div v-if="filter.q">
          {{ $gettext(`No warnings or error containing this keyword. Note that search is case-sensitive.`) }}
        </div>
        <div v-else>
          {{
            $gettext(
              `Log messages appear here whenever PhotoPrism comes across broken files, or there are other potential issues.`
            )
          }}
        </div>
      </v-alert>
    </div>
    <p-confirm-dialog
      :visible="dialog.delete"
      :text="$gettext(`Delete all?`)"
      icon="mdi-delete-sweep-outline"
      @close="dialog.delete = false"
      @confirm="onConfirmDelete"
    ></p-confirm-dialog>
    <v-dialog
      :model-value="details.visible"
      max-width="550"
      class="p-dialog"
      @keydown.esc.exact="details.visible = false"
    >
      <v-card>
        <v-card-title class="d-flex justify-start align-center ga-3">
          <v-icon v-if="details.err.Level === 'error'" icon="mdi-alert-circle-outline" color="error"></v-icon>
          <v-icon v-else-if="details.err.Level === 'warning'" icon="mdi-alert" color="warning"></v-icon>
          <v-icon v-else icon="mdi-information-outline" color="info"></v-icon>
          <h6 class="text-h6 text-capitalize">{{ formatLevel(details.err.Level) }}</h6>
        </v-card-title>

        <v-card-text>
          <div :class="'p-log-' + details.err.Level" class="p-log-message text-body-2 text-selectable" dir="ltr">
            <div :title="utcTime(details.err.Time)" class="p-log-message__time cursor-help mb-3">
              {{ localTime(details.err.Time) }}
            </div>
            <div class="text-break p-log-message__text">{{ details.err.Message }}</div>
          </div>
        </v-card-text>

        <v-card-actions>
          <v-btn color="button" variant="flat" class="action-close" @click="details.visible = false">
            {{ $gettext(`Close`) }}
          </v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>
  </div>
</template>

<script>
import { DateTime } from "luxon";
import $api from "common/api";
import links from "common/links";
import * as formats from "options/formats";

import PLoading from "component/loading.vue";
import PActionMenu from "component/action/menu.vue";
import PConfirmDialog from "component/confirm/dialog.vue";

export default {
  name: "PPageErrors",
  components: {
    PLoading,
    PActionMenu,
    PConfirmDialog,
  },
  expose: ["onShortCut"],
  data() {
    const query = this.$route.query;
    const q = query["q"] ? query["q"] : "";

    return {
      dirty: false,
      loading: false,
      scrollDisabled: false,
      scrollDistance: window.innerHeight * 2,
      filter: { q },
      isPublic: this.$config.get("public"),
      timeZone: this.$config.getTimeZone(),
      batchSize: 100,
      offset: 0,
      page: 0,
      errors: [],
      dialog: {
        delete: false,
      },
      details: {
        visible: false,
        err: { Level: "", Message: "", Time: "" },
      },
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
      this.filter.q = query["q"] ? query["q"] : "";
      this.onReload();
    },
  },
  created() {
    if (this.$config.deny("logs", "access_all")) {
      this.$router.push({ name: this.$session.getDefaultRoute() });
      return;
    }

    this.loadMore();
  },
  mounted() {
    this.$view.enter(this);
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
            this.onReload();
          },
        },
        {
          name: "troubleshooting",
          icon: "mdi-book-open-page-variant-outline",
          text: this.$gettext("Troubleshooting"),
          visible: true,
          href: links.troubleshooting,
          target: "_blank",
        },
      ];
    },
    onShortCut(ev) {
      switch (ev.code) {
        case "KeyR":
          this.onReload();
          return true;
        case "KeyF":
          this.$view.focus(this.$refs?.form, ".input-search input", true);
          return true;
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

      const query = {};

      Object.assign(query, this.filter);

      for (let key in query) {
        if (query[key] === undefined || !query[key]) {
          delete query[key];
        }
      }

      if (JSON.stringify(this.$route.query) === JSON.stringify(query)) {
        return false;
      }

      this.$router.replace({ query });

      return true;
    },
    showDetails(err) {
      this.details.err = err;
      this.details.visible = true;
    },
    onDelete() {
      if (this.loading) {
        return;
      }

      this.dialog.delete = true;
    },
    onConfirmDelete() {
      this.dialog.delete = false;

      if (this.loading) {
        return;
      }

      this.loading = true;
      this.scrollDisabled = true;

      // Delete error logs.
      $api
        .delete("errors")
        .then((resp) => {
          if (resp && resp.data.code && resp.data.code === 200) {
            this.errors = [];
            this.dirty = false;
            this.page = 0;
            this.offset = 0;
          }
        })
        .finally(() => {
          this.scrollDisabled = false;
          this.loading = false;
        });
    },
    onReload() {
      if (this.loading) {
        return;
      }

      this.page = 0;
      this.offset = 0;
      this.scrollDisabled = false;

      this.loadMore();
    },
    loadMore() {
      if (this.scrollDisabled) return;

      if (this.offset === 0) {
        this.loading = true;
      }

      this.scrollDisabled = true;

      const count = this.dirty ? (this.page + 2) * this.batchSize : this.batchSize;
      const offset = this.dirty ? 0 : this.offset;
      const q = this.filter.q;

      const params = { count, offset, q };

      // Fetch error logs.
      $api
        .get("errors", { params })
        .then((resp) => {
          if (!resp.data) {
            resp.data = [];
          }

          if (offset === 0) {
            this.errors = resp.data;
          } else {
            this.errors = this.errors.concat(resp.data);
          }

          this.scrollDisabled = resp.data.length < count;

          if (!this.scrollDisabled) {
            this.offset = offset + count;
            this.page++;
          }
        })
        .finally(() => {
          this.loading = false;
          this.dirty = false;
        });
    },
    level(s) {
      return s.substring(0, 4).toUpperCase();
    },
    formatLevel(level) {
      switch (level) {
        case "error":
          return this.$gettext("Error");
        case "warning":
          return this.$gettext("Warning");
      }

      return level;
    },
    localTime(s) {
      if (!s) {
        return this.$gettext("Unknown");
      }

      return DateTime.fromISO(s, { zone: this.timeZone }).toLocaleString(formats.TIMESTAMP_LONG_TZ);
    },
    utcTime(s) {
      if (!s) {
        return this.$gettext("Unknown");
      }

      return DateTime.fromISO(s, { zone: "UTC" }).toLocaleString(formats.TIMESTAMP_LONG_TZ);
    },
  },
};
</script>
