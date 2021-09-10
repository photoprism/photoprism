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
                      :disable-selection="true"
                      :filter="filter"
                      :open-photo="openPhotoByYear(year)"></p-photo-cards>
      </div>
    </v-container>
  </div>
</template>

<script>
import {Photo, TypeLive, TypeRaw, TypeVideo} from "model/photo";
import Thumb from "model/thumb";

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
      viewer: {
        results: [],
        loading: false,
      },
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
    photosByYear(year) {
      const photos = [];

      for (let item of this.results) {
        if (item.Year == year) {
          photos.push(item);
        }
      }

      return photos;
    },
    photosOffsetByYear(selectedYear) {
      let offset = 0;

      for (let year of this.years) {
        if (year > selectedYear) {
          offset += this.photosByYear(year).length;
        }
      }

      return offset;
    },
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
    openPhotoByYear(year) {
      const this_ = this;

      function openPhoto(index, showMerged) {
        const offset = this_.photosOffsetByYear(year);
        index += offset;

        if (this_.loading || !this_.results[index]) {
          return false;
        }

        const selected = this_.results[index];

        // Don't open as stack when a RAW has only one JPEG.
        if (selected.Type === TypeRaw && selected.jpegFiles().length < 2) {
          showMerged = false;
        }

        if (showMerged && selected.Type === TypeLive || selected.Type === TypeVideo) {
          if (selected.isPlayable()) {
            this_.$viewer.play({video: selected});
          } else {
            this_.$viewer.show(Thumb.fromPhotos(this_.results), index);
          }
        } else if (showMerged) {
          this_.$viewer.show(Thumb.fromFiles([selected]), 0);
        } else {
          this_.viewerResults().then((results) => {
            const thumbsIndex = results.findIndex(result => result.UID === selected.UID);

            if (thumbsIndex < 0) {
              this_.$viewer.show(Thumb.fromPhotos(this_.results), index);
            } else {
              this_.$viewer.show(Thumb.fromPhotos(results), thumbsIndex);
            }
          });
        }
      }

      return openPhoto;
    },
    viewerResults() {
      if (this.loading || this.viewer.loading) {
        return Promise.resolve(this.results);
        }

      if (this.viewer.results.length > (this.results.length + this.batchSize)) {
        return Promise.resolve(this.viewer.results);
      }

      this.viewer.loading = true;

      const count = this.batchSize * (this.page + 6);
      const offset = 0;

      const params = {
        count: count,
        offset: offset,
        merged: true,
      };

      Object.assign(params, this.filter);

      return Photo.search(params).then((resp) => {
        // Success.
        this.viewer.loading = false;
        this.viewer.results = resp.models;
        return Promise.resolve(this.viewer.results);
      }, () => {
        // Error.
        this.viewer.loading = false;
        return Promise.resolve(this.results);
    }
      );
    },
  }
};
</script>
