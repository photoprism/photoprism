<template>
  <div class="p-page p-page-errors" v-infinite-scroll="loadMore" :infinite-scroll-disabled="scrollDisabled"
       :infinite-scroll-distance="10" :infinite-scroll-listen-for-event="'scrollRefresh'">
    <v-toolbar flat color="secondary">
      <v-toolbar-title>
        <translate>Event Log</translate>
      </v-toolbar-title>
      <v-spacer></v-spacer>

      <v-btn icon @click.stop="reload" class="action-reload">
        <v-icon>refresh</v-icon>
      </v-btn>
    </v-toolbar>

    <v-list two-line>
      <v-list-tile
              v-for="(err, index) in errors" :key="index"
              avatar
              @click=""
      >
        <v-list-tile-avatar>
          <v-icon>{{ err.Level }}</v-icon>
        </v-list-tile-avatar>

        <v-list-tile-content>
          <v-list-tile-title>{{ err.Message }}</v-list-tile-title>
          <v-list-tile-sub-title>{{ formatTime(err.Time) }}</v-list-tile-sub-title>
        </v-list-tile-content>
      </v-list-tile>
    </v-list>
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
                loading: false,
            };
        },
        methods: {
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
            formatTime(s) {
                return DateTime.fromISO(s).toFormat("yyyy-LL-dd HH:mm:ss");
            },
        },
        created() {
            this.loadMore();
        },
    };
</script>
