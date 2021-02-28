<template>
  <div v-infinite-scroll="loadMore" class="p-tab p-tab-discover-today" :infinite-scroll-disabled="scrollDisabled"
       :infinite-scroll-distance="1200" :infinite-scroll-listen-for-event="'scrollRefresh'">

    <v-container v-if="loading" fluid class="pa-4">
      <v-progress-linear color="secondary-dark" :indeterminate="true"></v-progress-linear>
    </v-container>
    <v-container v-else-if="!results.length" text-xs-center fluid class="pa-4">
      <p class="subheading pb-3"><translate>No photos on this day.</translate></p>
    </v-container>
    <v-container v-else fluid class="pa-0">
      <p-scroll-top></p-scroll-top>

      <div v-for="year in years" :key="year">
        <h3 class="pl-3 pt-1">{{ year }}</h3>

        <!-- No clipboard, no toolbar only a single, card-based view (no edit-photo, no open-location, no selection) -->
        <p-photo-cards context="photos"
                      :photos="photosByYear(year)"
                      :select-mode="false"
                      :filter="filter"
                      :open-photo="openPhoto"></p-photo-cards>
      </div>
    </v-container>
  </div>
</template>

<script>
import {Photo} from "model/photo";

export default {
  name: 'PTabDiscoverToday',
  data() {
    const today = new Date();

    const filter = {
      month: today.getMonth() + 1,
      day: today.getDate(),
      order: 'newest',
      q: '',
    };

    return {
      results: [],
      filter: filter,
      loading: false,
      scrollDisabled: false,
      batchSize: Photo.batchSize(),
      offset: 0,
      page: 0,
    };
  },
  computed: {
    years: function() {
      const yrs = new Set()

      for (let i = 0; i < this.results.length; i++) {
        let item = this.results[i];
        yrs.add(item.Year);
      }

      return yrs;
    }
  },
  methods: {
    loadMore() {
      if (this.scrollDisabled) return;

      this.scrollDisabled = true;
      this.loading = true;

      const count = this.dirty ? (this.page + 2) * this.batchSize : this.batchSize;
      const offset = this.dirty ? 0 : this.offset;

      const params = {
        count: count,
        offset: offset,
        merged: true,
        month: this.filter.month,
        day: this.filter.day,
      };

      Photo.search(params).then(response => {
        this.results = Photo.mergeResponse(this.results, response);

        this.complete = (response.count < count);
        this.scrollDisabled = this.complete;

        if (this.complete) {
          this.offset = offset;

          if (this.results.length > 1) {
            this.$notify.info(this.$gettextInterpolate(this.$gettext("Showing all %{n} results"), {n: this.results.length}));
          }
        } else if (this.results.length >= Photo.limit()) {
          this.offset = offset;
          this.complete = true;
          this.scrollDisabled = true;
          this.$notify.warn(this.$gettext("Can't load more, limit reached"));
        } else {
          this.offset = offset + count;
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
      });
    },
    openPhoto() {
      // TODO
    },
    photosByYear(year) {
      const photos = [];

      for (let i = 0; i < this.results.length; i++) {
        let item = this.results[i];
        if (item.Year == year) {
          photos.push(item);
        }
      }

      return photos;
    }
  }
};
</script>
