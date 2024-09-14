<template>
  <div v-infinite-scroll="loadMore" class="p-page p-page-errors" :infinite-scroll-disabled="scrollDisabled" :infinite-scroll-distance="scrollDistance" :infinite-scroll-listen-for-event="'scrollRefresh'">
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
        browser-autocomplete="off"
        autocorrect="off"
        autocapitalize="none"
        :label="$gettext('Search')"
        prepend-inner-icon="search"
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
      <v-spacer></v-spacer>
      <v-btn icon class="action-reload" :title="$gettext('Reload')" @click.stop="onReload()">
        <v-icon>refresh</v-icon>
      </v-btn>
      <v-btn v-if="!isPublic" icon class="action-delete" :title="$gettext('Delete')" @click.stop="onDelete()">
        <v-icon>delete</v-icon>
      </v-btn>
      <v-btn icon href="https://docs.photoprism.app/getting-started/troubleshooting/" target="_blank" class="action-bug-report" :title="$gettext('Troubleshooting Checklists')">
        <v-icon>bug_report</v-icon>
      </v-btn>
    </v-toolbar>
    <v-container v-if="loading" fluid class="pa-4">
      <v-progress-linear color="secondary-dark" :indeterminate="true"></v-progress-linear>
    </v-container>
    <v-list v-else-if="errors.length > 0" dense two-line class="transparent pa-1">
      <v-list-tile v-for="err in errors" :key="err.ID" avatar class="rounded-4" @click="showDetails(err)">
        <v-list-tile-avatar>
          <v-icon :color="err.Level">{{ err.Level }}</v-icon>
        </v-list-tile-avatar>

        <v-list-tile-content class="text-selectable">
          <v-list-tile-title>{{ err.Message }}</v-list-tile-title>
          <v-list-tile-sub-title>{{ formatTime(err.Time) }}</v-list-tile-sub-title>
        </v-list-tile-content>
      </v-list-tile>
    </v-list>
    <div v-else class="pa-2">
      <v-alert :value="true" color="secondary-dark" icon="check_circle_outline" class="no-results ma-2 opacity-70" outline>
        <p class="body-1 mt-0 mb-0 pa-0">
          <template v-if="filter.q !== ''">
            <translate>No warnings or error containing this keyword. Note that search is case-sensitive.</translate>
          </template>
          <template>
            <translate>Log messages appear here whenever PhotoPrism comes across broken files, or there are other potential issues.</translate>
          </template>
        </p>
      </v-alert>
    </div>
    <p-confirm-dialog :show="dialog.delete" icon="delete_outline" @cancel="dialog.delete = false" @confirm="onConfirmDelete"></p-confirm-dialog>
    <v-dialog v-model="details.show" max-width="500">
      <v-card class="pa-2">
        <v-card-title class="headline pa-2">
          {{ details.err.Level | capitalize }}
        </v-card-title>

        <v-card-text class="pa-2 body-2">
          {{ localTime(details.err.Time) }}
        </v-card-text>

        <v-card-text class="pa-2 body-1">
          {{ details.err.Message }}
        </v-card-text>

        <v-card-actions class="pa-2">
          <v-spacer></v-spacer>
          <v-btn color="secondary-light" depressed class="action-close" @click="details.show = false">
            <translate>Close</translate>
          </v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>
  </div>
</template>

<script>
import { DateTime } from "luxon";
import Api from "common/api";

export default {
  name: "PPageErrors",
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
      batchSize: 100,
      offset: 0,
      page: 0,
      errors: [],
      dialog: {
        delete: false,
      },
      details: {
        show: false,
        err: { Level: "", Message: "", Time: "" },
      },
    };
  },
  watch: {
    $route() {
      const query = this.$route.query;
      this.filter.q = query["q"] ? query["q"] : "";
      this.onReload();
    },
  },
  created() {
    if (this.$config.deny("logs", "view")) {
      this.$router.push({ name: "albums" });
      return;
    }

    this.loadMore();
  },
  methods: {
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

      const query = {};

      Object.assign(query, this.filter);

      for (let key in query) {
        if (query[key] === undefined || !query[key]) {
          delete query[key];
        }
      }

      if (JSON.stringify(this.$route.query) === JSON.stringify(query)) {
        return;
      }

      this.$router.replace({ query });
    },
    showDetails(err) {
      this.details.err = err;
      this.details.show = true;
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
      Api.delete("errors")
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
      Api.get("errors", { params })
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
    localTime(s) {
      if (!s) {
        return this.$gettext("Unknown");
      }

      return DateTime.fromISO(s).toLocaleString(DateTime.DATETIME_FULL_WITH_SECONDS);
    },
    formatTime(s) {
      if (!s) {
        return this.$gettext("Unknown");
      }

      return DateTime.fromISO(s).toFormat("yyyy-LL-dd HH:mm:ss");
    },
  },
};
</script>
