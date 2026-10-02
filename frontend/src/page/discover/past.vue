<template>
  <div class="p-tab p-tab-discover-past" :class="$config.aclClasses('photos')">
    <div v-if="loading && results.length === 0" class="p-page__loading">
      <p-loading></p-loading>
    </div>
    <div v-else class="p-page__content">
      <p-scroll :load-more="loadMore" :load-disabled="scrollDisabled" :load-distance="scrollDistance" :loading="loading"></p-scroll>
      <p-photo-clipboard :context="context" :refresh="refresh"></p-photo-clipboard>

      <div v-if="yearGroups.length === 0" class="pa-3">
        <v-alert color="surface-variant" icon="mdi-lightbulb-outline" class="no-results" variant="outlined">
          <div class="font-weight-bold">
            {{ emptyTitle }}
          </div>
          <div class="mt-2">
            {{ emptyBody }}
          </div>
        </v-alert>
      </div>

      <section v-for="group in yearGroups" :id="yearSectionId(group.year)" :key="group.year" class="p-discover-year">
        <header class="p-discover-year-header bg-background">
          <h2 class="p-discover-year-title">{{ group.year }}</h2>
          <p class="p-discover-year-meta">
            {{ yearsAgoText(group.yearsAgo) }}
            <span v-if="complete" class="p-discover-year-count">{{ pictureCountText(group.count) }}</span>
          </p>
        </header>
        <p-photo-view-mosaic
          v-if="view === 'mosaic'"
          :context="context"
          :photos="group.photos"
          :select-mode="selectMode"
          :filter="filter"
          :edit-photo="(index, tab) => editPhotoInYear(group.year, index, tab)"
          :open-photo="(index, merged) => openPhotoInYear(group.year, index, merged)"
          :is-shared-view="isShared"
        ></p-photo-view-mosaic>
        <p-photo-view-cards
          v-else
          :context="context"
          :photos="group.photos"
          :select-mode="selectMode"
          :filter="filter"
          :open-photo="(index, merged) => openPhotoInYear(group.year, index, merged)"
          :edit-photo="(index, tab) => editPhotoInYear(group.year, index, tab)"
          :open-date="(index) => openDateInYear(group.year, index)"
          :open-location="(index) => openLocationInYear(group.year, index)"
          :is-shared-view="isShared"
        ></p-photo-view-cards>
      </section>
    </div>
  </div>
</template>

<script>
import { Photo } from "model/photo";
import Thumb from "model/thumb";
import { $gettext } from "common/gettext";
import * as contexts from "options/contexts";
import PPhotoClipboard from "component/photo/clipboard.vue";
import PPhotoViewCards from "component/photo/view/cards.vue";
import PPhotoViewMosaic from "component/photo/view/mosaic.vue";
import PLoading from "component/loading.vue";
import PScroll from "component/scroll.vue";
import { groupPhotosByYear, isPastYearPhoto, discoverSearchParams, pageAdvance, viewerToday } from "common/discover";

export default {
  name: "PTabDiscoverPast",
  components: {
    PPhotoClipboard,
    PPhotoViewCards,
    PPhotoViewMosaic,
    PLoading,
    PScroll,
  },
  props: {
    mode: {
      type: String,
      default: "day",
    },
    view: {
      type: String,
      default: "cards",
    },
  },
  data() {
    const settings = this.$config.getSettings();
    const features = settings?.features || {};
    const batchSize = Photo.batchSize();

    return {
      isShared: this.$config.deny("photos", "manage"),
      canEdit: this.$config.allow("photos", "update") && features.edit,
      listen: false,
      dirty: false,
      complete: false,
      results: [],
      scrollDisabled: true,
      scrollDistance: window.innerHeight * 4,
      batchSize,
      offset: 0,
      page: 0,
      selection: this.$clipboard.selection,
      filter: { order: "newest" },
      loading: true,
      lightbox: { open: false, loading: false },
      subscriptions: [],
    };
  },
  computed: {
    selectMode() {
      return this.selection.length > 0;
    },
    context() {
      return contexts.Photos;
    },
    today() {
      return viewerToday(this.$config.getTimeZone());
    },
    yearGroups() {
      return groupPhotosByYear(this.results, this.today.year);
    },
    emptyTitle() {
      return this.mode === "month" ? $gettext("No pictures from this month in past years") : $gettext("No pictures from this day in past years");
    },
    emptyBody() {
      return $gettext("Photos taken in previous years will appear here when they match today's date.");
    },
  },
  mounted() {
    this.subscriptions.push(
      this.$event.subscribe("lightbox.opened", () => {
        this.lightbox.open = true;
      })
    );
    this.subscriptions.push(
      this.$event.subscribe("lightbox.closed", () => {
        this.lightbox.open = false;
      })
    );
    this.search();
  },
  unmounted() {
    this.subscriptions.forEach((id) => this.$event.unsubscribe(id));
    this.subscriptions = [];
  },
  methods: {
    // visibilityParams copies Browse's private/review filters so Discover matches the library.
    visibilityParams() {
      const settings = this.$config.getSettings();
      const features = settings?.features || {};
      const params = {};

      if (features.private) {
        params.public = "true";
      }

      if (features.review) {
        params.quality = "3";
      }

      return params;
    },
    // searchParams builds the past-day or past-month photo query for the active tab.
    searchParams(offset = 0) {
      return discoverSearchParams({
        mode: this.mode,
        timeZone: this.$config.getTimeZone(),
        years: this.$config.values?.years || [],
        count: this.batchSize,
        offset,
        ...this.visibilityParams(),
      });
    },
    // search loads the first page of past photos and replaces the current results.
    search() {
      if (this.lightbox.open) {
        return;
      }

      this.scrollDisabled = true;
      this.offset = 0;
      this.page = 0;
      this.loading = true;
      this.listen = false;
      this.complete = false;

      Photo.search(this.searchParams(0))
        .then((response) => {
          this.results = (response.models || []).filter((photo) => this.keepPhoto(photo));
          this.applyPage(response, 0, this.batchSize);
          this.maybeFillViewport();
        })
        .catch(() => {
          this.results = [];
          this.scrollDisabled = false;
        })
        .finally(() => {
          this.loading = false;
          this.listen = true;
        });
    },
    // keepPhoto drops current-year and undated results left in by the before fallback.
    keepPhoto(photo) {
      return isPastYearPhoto(photo, this.today.year);
    },
    // applyPage stores the next offset from the page the server actually returned.
    applyPage(response, requestedOffset, requestedCount) {
      const page = pageAdvance(response, requestedOffset, requestedCount);
      this.complete = page.complete || this.results.length >= Photo.limit();
      this.scrollDisabled = this.complete;
      this.offset = page.offset;
    },
    // maybeFillViewport loads another page when the first results do not fill the screen.
    maybeFillViewport() {
      this.$nextTick(() => {
        const height = this.$el?.clientHeight || 0;

        if (height > 0 && !this.scrollDisabled && height <= window.document.documentElement.clientHeight + 300) {
          this.loadMore();
        }
      });
    },
    // loadMore appends the next page of past photos for infinite scroll.
    loadMore(force) {
      if (!force && (this.scrollDisabled || this.$view.isHidden("PPageDiscover"))) {
        return;
      }

      this.scrollDisabled = true;
      this.loading = true;
      this.listen = false;

      const count = this.dirty ? (this.page + 2) * this.batchSize : this.batchSize;
      const offset = this.dirty ? 0 : this.offset;
      const params = this.searchParams(offset);
      params.count = count;

      Photo.search(params)
        .then((response) => {
          const models = (response.models || []).filter((photo) => this.keepPhoto(photo));
          this.results = this.dirty ? models : Photo.mergeResponse(this.results, { ...response, models });
          this.applyPage(response, offset, count);

          if (!this.complete) {
            this.page++;
            this.maybeFillViewport();
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
    // refresh reloads the visible past photos after a library change.
    refresh() {
      if (this.loading || !this.listen) {
        return;
      }

      this.dirty = true;
      this.complete = false;
      this.scrollDisabled = false;
      this.loadMore(true);
    },
    yearSectionId(year) {
      return `discover-year-${year}`;
    },
    // yearsAgoText formats the relative year label.
    yearsAgoText(yearsAgo) {
      if (yearsAgo === 1) {
        return $gettext("1 year ago");
      }

      return $gettext("%{n} years ago", { n: yearsAgo });
    },
    // pictureCountText formats the loaded photo count for a year.
    pictureCountText(count) {
      if (count === 1) {
        return $gettext("1 picture");
      }

      return $gettext("%{n} pictures", { n: count });
    },
    // indexInResults maps a photo inside one year section onto the flat result list.
    indexInResults(year, localIndex) {
      const group = this.yearGroups.find((item) => item.year === year);
      const photo = group?.photos?.[localIndex];

      if (!photo) {
        return -1;
      }

      return this.results.findIndex((item) => item.UID === photo.UID);
    },
    // openPhotoInYear opens the photo at a year-section index.
    openPhotoInYear(year, localIndex, showMerged = false) {
      return this.openPhoto(this.indexInResults(year, localIndex), showMerged);
    },
    // editPhotoInYear opens the editor for a photo at a year-section index.
    editPhotoInYear(year, localIndex, tab) {
      return this.editPhoto(this.indexInResults(year, localIndex), tab);
    },
    // openDateInYear opens the taken date for a photo at a year-section index.
    openDateInYear(year, localIndex) {
      return this.openDate(this.indexInResults(year, localIndex));
    },
    // openLocationInYear opens the place for a photo at a year-section index.
    openLocationInYear(year, localIndex) {
      return this.openLocation(this.indexInResults(year, localIndex));
    },
    // openPhoto opens one result in the lightbox.
    openPhoto(index, showMerged = false) {
      if (this.loading || !this.listen || this.lightbox.loading || !this.results[index]) {
        return false;
      }

      const selected = this.results[index];

      if (this.selection.length > 0 || selected.jpegFiles().length < 2) {
        showMerged = false;
      }

      if (showMerged) {
        this.$lightbox.openModels(Thumb.fromFiles([selected]), 0);
      } else {
        this.$lightbox.openView(this, index);
      }

      return true;
    },
    // getLightboxContext supplies the flat result list so the lightbox can page across years.
    getLightboxContext(index = 0) {
      return {
        models: Thumb.fromPhotos(this.results),
        index,
        context: this.context,
        allowEdit: this.canEdit,
        allowSelect: true,
      };
    },
    // editPhoto opens the edit dialog for one result.
    editPhoto(index, tab) {
      if (!this.canEdit) {
        return this.openPhoto(index);
      }

      const selection = this.results.map((photo) => photo.getId());
      this.$event.publish("dialog.edit", { selection, album: null, index, tab });
    },
    // openDate opens Browse for the photo's taken date.
    openDate(index) {
      const photo = this.results[index];

      if (!photo) {
        return;
      } else if (!photo.TakenAt || photo.TakenAt.length < 10) {
        this.editPhoto(index);
        return;
      }

      const takenDate = photo.TakenAt.substring(0, 10);

      if (this.$isMobile) {
        this.$router.push({ name: "browse", query: { q: "taken:" + takenDate } });
      } else {
        const routeUrl = this.$router.resolve({ name: "all", query: { q: "taken:" + takenDate } }).href;
        if (routeUrl) {
          this.$util.openUrl(routeUrl);
        }
      }
    },
    // openLocation opens Places for the photo's cell or country.
    openLocation(index) {
      const photo = this.results[index];

      if (!photo) {
        return;
      }

      if (photo.CellID && photo.CellID !== "zz") {
        this.$router.push({ name: "places", query: { q: photo.CellID } });
      } else if (photo.Country && photo.Country !== "zz") {
        this.$router.push({ name: "places", query: { q: "country:" + photo.Country } });
      } else {
        this.$notify.warn($gettext("Unknown location"));
      }
    },
  },
};
</script>
