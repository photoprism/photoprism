<template>
  <div class="p-tab p-tab-discover-random" :class="$config.aclClasses('photos')">
    <div v-if="loading && results.length === 0" class="p-page__loading">
      <p-loading></p-loading>
    </div>
    <div v-else class="p-page__content">
      <p-photo-clipboard :context="context" :refresh="shuffle"></p-photo-clipboard>

      <p-photo-view-mosaic
        v-if="view === 'mosaic'"
        :context="context"
        :photos="results"
        :select-mode="selectMode"
        :filter="filter"
        :edit-photo="editPhoto"
        :open-photo="openPhoto"
        :is-shared-view="isShared"
      ></p-photo-view-mosaic>
      <p-photo-view-cards
        v-else
        :context="context"
        :photos="results"
        :select-mode="selectMode"
        :filter="filter"
        :open-photo="openPhoto"
        :edit-photo="editPhoto"
        :open-date="openDate"
        :open-location="openLocation"
        :is-shared-view="isShared"
      ></p-photo-view-cards>
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
import { randomSearchParams } from "common/discover";

export default {
  name: "PTabDiscoverRandom",
  components: {
    PPhotoClipboard,
    PPhotoViewCards,
    PPhotoViewMosaic,
    PLoading,
  },
  props: {
    view: {
      type: String,
      default: "cards",
    },
  },
  emits: ["loading"],
  data() {
    const settings = this.$config.getSettings();
    const features = settings?.features || {};

    return {
      isShared: this.$config.deny("photos", "manage"),
      canEdit: this.$config.allow("photos", "update") && features.edit,
      listen: false,
      results: [],
      selection: this.$clipboard.selection,
      filter: { order: "random" },
      loading: true,
      fetching: false,
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
    this.shuffle();
  },
  unmounted() {
    this.subscriptions.forEach((id) => this.$event.unsubscribe(id));
    this.subscriptions = [];
  },
  methods: {
    // visibilityParams copies Browse's private/review filters so random results match the library.
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
    // shuffle fetches a new finite random set and reports loading to the toolbar.
    shuffle() {
      if (this.lightbox.open || this.fetching) {
        return;
      }

      this.fetching = true;
      this.loading = true;
      this.listen = false;
      this.$emit("loading", true);

      Photo.search(randomSearchParams(this.visibilityParams()))
        .then((response) => {
          this.results = response.models || [];
        })
        .catch(() => {
          this.results = [];
        })
        .finally(() => {
          this.fetching = false;
          this.loading = false;
          this.listen = true;
          this.$emit("loading", false);
        });
    },
    // openPhoto opens one random result in the lightbox.
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
        this.$lightbox.openModels(Thumb.fromPhotos(this.results), index);
      }

      return true;
    },
    // getLightboxContext supplies the loaded random set so the lightbox can page through it.
    getLightboxContext(index = 0) {
      return {
        models: Thumb.fromPhotos(this.results),
        index,
        context: this.context,
        allowEdit: this.canEdit,
        allowSelect: true,
      };
    },
    // editPhoto opens the edit dialog for one random result.
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
