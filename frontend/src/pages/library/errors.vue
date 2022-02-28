<template>
  <div v-infinite-scroll="loadMore" class="p-page p-page-errors" :infinite-scroll-disabled="scrollDisabled"
       :infinite-scroll-distance="1200" :infinite-scroll-listen-for-event="'scrollRefresh'">
    <v-toolbar flat :dense="$vuetify.breakpoint.smAndDown" class="page-toolbar" color="secondary">
      <v-text-field v-model="filter.q"
                    class="input-search background-inherit elevation-0"
                    browser-autocomplete="off"
                    autocorrect="off"
                    autocapitalize="none"
                    solo hide-details clearable overflow
                    :label="$gettext('Search')"
                    prepend-inner-icon="search"
                    color="secondary-dark"
                    @click:clear="clearQuery"
                    @keyup.enter.native="updateQuery"
      ></v-text-field>

      <v-btn icon class="action-reload" :title="$gettext('Reload')" @click.stop="reload">
        <v-icon>refresh</v-icon>
      </v-btn>

      <v-btn icon href="https://github.com/photoprism/photoprism/issues" target="_blank" class="action-bug-report"
             :title="$gettext('Report Bug')">
        <v-icon>bug_report</v-icon>
      </v-btn>
    </v-toolbar>
    <v-container v-if="loading" fluid class="pa-4">
      <v-progress-linear color="secondary-dark" :indeterminate="true"></v-progress-linear>
    </v-container>
    <v-list v-else-if="errors.length > 0" dense two-line class="transparent">
      <v-list-tile
          v-for="err in errors" :key="err.ID"
          avatar
          @click="showDetails(err)"
      >
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
      <v-alert
          :value="true"
          color="secondary-dark" icon="check_circle_outline" class="no-results ma-2 opacity-70" outline
      >
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

    <v-dialog
        v-model="details.show"
        max-width="500"
    >
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
          <v-btn color="secondary-light" depressed class="action-close"
                 @click="details.show = false">
            <translate>Close</translate>
          </v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>
  </div>
</template>

<script>
import {DateTime} from "luxon";
import Api from "common/api";

export default {
  name: 'PPageErrors',
  data() {
    const query = this.$route.query;
    const q = query["q"] ? query["q"] : "";

    return {
      dirty: false,
      loading: false,
      scrollDisabled: false,
      filter: {q},
      batchSize: 100,
      offset: 0,
      page: 0,
      errors: [],
      results: [],
      details: {
        show: false,
        err: {"Level": "", "Message": "", "Time": ""},
      },
    };
  },
  watch: {
    '$route'() {
      const query = this.$route.query;
      this.filter.q = query['q'] ? query['q'] : '';
      this.reload();
    }
  },
  created() {
    this.loadMore();
  },
  methods: {
    updateQuery() {
      this.filter.q = this.filter.q.trim();

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

      this.$router.replace({query});
    },
    clearQuery() {
      this.filter.q = "";
      this.updateQuery();
    },
    showDetails(err) {
      this.details.err = err;
      this.details.show = true;
    },
    reload() {
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

      const params = {count, offset, q};

      Api.get("errors", {params}).then((resp) => {
        if (!resp.data) {
          resp.data = [];
        }

        if (offset === 0) {
          this.errors = resp.data;
        } else {
          this.errors = this.errors.concat(resp.data);
        }

        this.scrollDisabled = (resp.data.length < count);

        if (!this.scrollDisabled) {
          this.offset = offset + count;
          this.page++;
        }
      }).finally(() => {
        this.loading = false;
        this.dirty = false;
      });
    },
    level(s) {
      return s.substr(0, 4).toUpperCase();
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
