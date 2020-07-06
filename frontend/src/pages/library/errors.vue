<template>
  <div class="p-page p-page-errors" v-infinite-scroll="loadMore" :infinite-scroll-disabled="scrollDisabled"
       :infinite-scroll-distance="10" :infinite-scroll-listen-for-event="'scrollRefresh'">
    <v-toolbar flat color="secondary">
      <v-toolbar-title v-if="errors.length > 0">
        <translate>Event Log</translate>
      </v-toolbar-title>
      <v-toolbar-title v-else>
        <translate>No warnings or errors</translate>
      </v-toolbar-title>

      <v-spacer></v-spacer>

      <v-btn icon @click.stop="reload" class="action-reload">
        <v-icon>refresh</v-icon>
      </v-btn>
    </v-toolbar>

    <v-list dense two-line v-if="errors.length > 0">
      <v-list-tile
              v-for="(err, index) in errors" :key="index"
              avatar
              @click="showDetails(err)"
      >
        <v-list-tile-avatar>
          <v-icon :color="err.Level">{{ err.Level }}</v-icon>
        </v-list-tile-avatar>

        <v-list-tile-content>
          <v-list-tile-title>{{ err.Message }}</v-list-tile-title>
          <v-list-tile-sub-title>{{ formatTime(err.Time) }}</v-list-tile-sub-title>
        </v-list-tile-content>
      </v-list-tile>
    </v-list>
    <v-card v-else class="errors-empty secondary-light lighten-1 ma-0 pa-2" flat>
      <v-card-title primary-title>
        <translate>When PhotoPrism found broken files or there are other potential issues, you'll see a short message on this page.</translate>
      </v-card-title>
    </v-card>

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
          <v-btn color="secondary-light" depressed @click="details.show = false"
                 class="action-close">
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
        name: 'p-page-errors',
        data() {
            return {
                errors: [],
                dirty: false,
                results: [],
                scrollDisabled: false,
                pageSize: 100,
                offset: 0,
                page: 0,
                details: {
                    show: false,
                    err: {"Level": "", "Message": "", "Time": ""},
                },
                loading: false,
            };
        },
        methods: {
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
                if (this.loading || this.scrollDisabled) return;

                this.loading = true;
                this.scrollDisabled = true;

                const count = this.dirty ? (this.page + 2) * this.pageSize : this.pageSize;
                const offset = this.dirty ? 0 : this.offset;

                const params = {
                    count: count,
                    offset: offset,
                };

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
                }).finally(() => this.loading = false)
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
        created() {
            this.loadMore();
        },
    };
</script>
