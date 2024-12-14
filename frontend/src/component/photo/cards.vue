<template>
  <v-container grid-list-xs fluid class="pa-1 p-photos p-photo-cards">
    <template v-if="photos.length === 0">
      <v-alert color="surface-variant" :icon="isSharedView ? 'mdi-image-off' : 'mdi-lightbulb-outline'" class="no-results ma-2 opacity-70" variant="outlined">
        <h3 v-if="filter.order === 'edited'" class="text-subtitle-2 ma-0 pa-0">
          <translate>No recently edited pictures</translate>
        </h3>
        <h3 v-else class="text-subtitle-2 ma-0 pa-0">
          <translate>No pictures found</translate>
        </h3>
        <p class="mt-2 mb-0 pa-0">
          <translate>Try again using other filters or keywords.</translate>
          <template v-if="!isSharedView">
            <translate>In case pictures you expect are missing, please rescan your library and wait until indexing has been completed.</translate>
            <template v-if="$config.feature('review')">
              <translate>Non-photographic and low-quality images require a review before they appear in search results.</translate>
            </template>
          </template>
        </p>
      </v-alert>
    </template>
    <div class="v-row search-results photo-results cards-view ma-0" :class="{ 'select-results': selectMode }">
      <div v-for="(photo, index) in photos" ref="items" :key="photo.ID" :data-index="index" class="v-col-12 v-col-sm-6 v-col-md-4 v-col-lg-3 v-col-xl-2 v-col-xxl-1 pa-1">
        <div v-if="index < firstVisibleElementIndex || index > lastVisibileElementIndex" :data-uid="photo.UID" class="result card bg-card placeholder">
          <div class="card image" />
          <div v-if="photo.Quality < 3 && context === 'review'" style="width: 100%; height: 34px" />
          <div class="pa-6 card-details">
            <div>
              <h3 class="text-subtitle-2 mb-2" :title="photo.Title">
                <!-- TODO: change this filter -->
                <!-- {{ photo.Title | truncate(80) }} -->
                {{ photo.Title }}
              </h3>
              <div v-if="photo.Description" class="text-caption mb-2">
                {{ photo.Description }}
              </div>
              <div class="text-caption">
                <i />
                {{ photo.getDateString(true) }}
                <br />
                <i />
                <template v-if="photo.Type === 'video' || photo.Type === 'animated'">
                  {{ photo.getVideoInfo() }}
                </template>
                <template v-else>
                  {{ photo.getPhotoInfo() }}
                </template>
                <template v-if="filter.order === 'name' && $config.feature('download')">
                  <br />
                  <i />
                  {{ photo.baseName() }}
                </template>
                <template v-if="featPlaces && photo.Country !== 'zz'">
                  <br />
                  <i />
                  {{ photo.locationInfo() }}
                </template>
              </div>
            </div>
          </div>
        </div>
        <div v-else :data-id="photo.ID" :data-uid="photo.UID" class="result card bg-card" :class="photo.classes()" @contextmenu.stop="onContextMenu($event, index)">
          <div
            :key="photo.Hash"
            :title="photo.Title"
            class="card clickable image"
            :style="`background-image: url(${photo.thumbnailUrl('tile_500')})`"
            @touchstart.passive="input.touchStart($event, index)"
            @touchend.stop.prevent="onClick($event, index)"
            @mousedown.stop.prevent="input.mouseDown($event, index)"
            @click.stop.prevent="onClick($event, index)"
            @mouseover="playLive(photo)"
            @mouseleave="pauseLive(photo)"
          >
            <v-row v-if="photo.Type === 'live' || photo.Type === 'animated'" class="live-player">
              <video :id="'live-player-' + photo.ID" :key="photo.ID" width="500" height="500" preload="none" loop muted playsinline>
                <source :src="photo.videoUrl()" />
              </video>
            </v-row>

            <button
              v-if="photo.Type !== 'image' || photo.isStack()"
              class="input-open"
              @touchstart.stop.prevent="input.touchStart($event, index)"
              @touchend.stop.prevent="onOpen($event, index, !isSharedView, photo.Type === 'live')"
              @touchmove.stop.prevent
              @click.stop.prevent="onOpen($event, index, !isSharedView, photo.Type === 'live')"
            >
              <i v-if="photo.Type === 'raw'" class="action-raw mdi mdi-raw" :title="$gettext('RAW')" />
              <i v-if="photo.Type === 'live'" class="action-live" :title="$gettext('Live')"><icon-live-photo /></i>
              <i v-if="photo.Type === 'video'" class="mdi mdi-play" :title="$gettext('Video')" />
              <i v-if="photo.Type === 'animated'" class="mdi mdi-file-gif-box" :title="$gettext('Animated')" />
              <i v-if="photo.Type === 'vector'" class="action-vector mdi mdi-vector-polyline" :title="$gettext('Vector')"></i>
              <i v-if="photo.Type === 'image'" class="mdi mdi-camera-burst" :title="$gettext('Stack')" />
            </button>

            <button v-if="photo.Type === 'image' && selectMode" class="input-view" :title="$gettext('View')" @touchstart.stop.prevent="input.touchStart($event, index)" @touchend.stop.prevent="onOpen($event, index)" @touchmove.stop.prevent @click.stop.prevent="onOpen($event, index)">
              <i class="mdi mdi-magnify-plus-outline" />
            </button>

            <button v-if="!isSharedView && featPrivate && photo.Private" class="input-private">
              <i class="mdi mdi-lock" />
            </button>

            <!--
              We'd usually use v-if here to only render the button if needed.
              Because the button is supposed to be visible when the result is
              being hovered over, implementing the v-if would require the use of
              a <v-hover> element around the result.

              Because rendering the plain HTML-Button is faster than rendering
              the v-hover component we instead hide the button by default and
              use css to show it when it is being hovered.
            -->
            <button class="input-select" @touchstart.stop.prevent="input.touchStart($event, index)" @touchend.stop.prevent="onSelect($event, index)" @touchmove.stop.prevent @click.stop.prevent="onSelect($event, index)">
              <i class="mdi mdi-check-circle select-on" />
              <i class="mdi mdi-circle-outline select-off" />
            </button>

            <button v-if="!isSharedView" class="input-favorite" @touchstart.stop.prevent="input.touchStart($event, index)" @touchend.stop.prevent="toggleLike($event, index)" @touchmove.stop.prevent @click.stop.prevent="toggleLike($event, index)">
              <i v-if="photo.Favorite" class="mdi mdi-star text-favorite" />
              <i v-else class="mdi mdi-star-outline" />
            </button>
          </div>

          <v-card-actions v-if="!isSharedView && photo.Quality < 3 && context === 'review'" class="card-details pa-0">
            <v-row align="center">
              <v-col cols="6" class="text-center pa-1">
                <v-btn color="card darken-1" density="comfortable" variant="flat" block :rounded="false" class="action-archive text-center" :title="$gettext('Archive')" @click.stop="photo.archive()">
                  <v-icon>mdi-close</v-icon>
                </v-btn>
              </v-col>
              <v-col cols="6" class="text-center pa-1">
                <v-btn color="card darken-1" density="comfortable" variant="flat" block :rounded="false" class="action-approve text-center" :title="$gettext('Approve')" @click.stop="photo.approve()">
                  <v-icon>mdi-check</v-icon>
                </v-btn>
              </v-col>
            </v-row>
          </v-card-actions>

          <div class="pa-6 card-details">
            <div>
              <h3 class="text-body-2 mb-1" :title="photo.Title">
                <button class="action-title-edit" :data-uid="photo.UID" @click.exact="isSharedView ? openPhoto(index) : editPhoto(index)">
                  <!-- TODO: change this filter -->
                  <!-- {{ photo.Title | truncate(80) }} -->
                  {{ photo.Title }}
                </button>
              </h3>
              <div v-if="photo.Description" class="text-caption mb-1" :title="$gettext('Description')">
                <button @click.exact="editPhoto(index)">
                  {{ photo.Description }}
                </button>
              </div>
              <div v-if="filter.order === 'name' && $config.feature('download')" class="text-caption">
                <button :title="$gettext('Name')" @click.exact="downloadFile(index)">
                  <i class="mdi mdi-file" />
                  {{ photo.baseName() }}
                </button>
              </div>
              <div class="text-caption">
                <button class="action-open-date" :data-uid="photo.UID" @click.exact="openDate(index)">
                  <i :title="$gettext('Taken')" class="mdi mdi-calendar-range" />
                  {{ photo.getDateString(true) }}
                </button>
                <br />
                <button v-if="photo.Type === 'video'" :title="$gettext('Video')" @click.exact="openPhoto(index)">
                  <i class="mdi mdi-movie" />
                  {{ photo.getVideoInfo() }}
                </button>
                <button v-else-if="photo.Type === 'live'" :title="$gettext('Live')" @click.exact="openPhoto(index)">
                  <i class="mdi mdi-play-circle" />
                  {{ photo.getVideoInfo() }}
                </button>
                <button v-else-if="photo.Type === 'animated'" :title="$gettext('Animated') + ' GIF'" @click.exact="openPhoto(index)">
                  <i class="mdi mdi-file-gif-box" />
                  {{ photo.getVideoInfo() }}
                </button>
                <button v-else-if="photo.Type === 'vector'" :title="$gettext('Vector')" @click.exact="openPhoto(index)">
                  <i class="mdi mdi-vector-polyline" />
                  {{ photo.getVectorInfo() }}
                </button>
                <button v-else :title="$gettext('Camera')" class="action-camera-edit" :data-uid="photo.UID" @click.exact="editPhoto(index)">
                  <i class="mdi mdi-camera" />
                  {{ photo.getPhotoInfo() }}
                </button>
                <button v-if="photo.LensID > 1 || photo.FocalLength" :title="$gettext('Lens')" class="action-lens-edit" :data-uid="photo.UID" @click.exact="editPhoto(index)">
                  <i class="mdi mdi-camera-iris" />
                  {{ photo.getLensInfo() }}
                </button>
                <template v-if="featPlaces && photo.Country !== 'zz'">
                  <br />
                  <button :title="$gettext('Location')" class="action-location" :data-uid="photo.UID" @click.exact="openLocation(index)">
                    <i class="mdi mdi-map-marker" />
                    {{ photo.locationInfo() }}
                  </button>
                </template>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  </v-container>
</template>
<script>
import download from "common/download";
import Notify from "common/notify";
import { Input, InputInvalid, ClickShort, ClickLong } from "common/input";
import { virtualizationTools } from "common/virtualization-tools";
import IconLivePhoto from "component/icon/live-photo.vue";

export default {
  name: "PPhotoCards",
  components: {
    IconLivePhoto,
  },
  props: {
    photos: {
      type: Array,
      default: () => [],
    },
    openPhoto: {
      type: Function,
      default: () => {},
    },
    editPhoto: {
      type: Function,
      default: () => {},
    },
    openDate: {
      type: Function,
      default: () => {},
    },
    openLocation: {
      type: Function,
      default: () => {},
    },
    album: {
      type: Object,
      default: () => {},
    },
    filter: {
      type: Object,
      default: () => {},
    },
    context: {
      type: String,
      default: "",
    },
    selectMode: Boolean,
    isSharedView: {
      type: Boolean,
      default: false,
    },
  },
  data() {
    const featPlaces = this.$config.settings().features.places;
    const featPrivate = this.$config.settings().features.private;
    const input = new Input();
    const debug = this.$config.get("debug");

    return {
      featPlaces,
      featPrivate,
      debug,
      input,
      firstVisibleElementIndex: 0,
      lastVisibileElementIndex: 0,
      visibleElementIndices: new Set(),
    };
  },
  watch: {
    photos: {
      handler() {
        this.$nextTick(() => {
          this.observeItems();
        });
      },
      immediate: true,
    },
  },
  beforeCreate() {
    this.intersectionObserver = new IntersectionObserver(
      (entries) => {
        this.visibilitiesChanged(entries);
      },
      {
        rootMargin: "50% 0px",
      }
    );
  },
  beforeUnmount() {
    this.intersectionObserver.disconnect();
  },
  methods: {
    observeItems() {
      if (this.$refs.items === undefined) {
        return;
      }

      /**
       * observing only every 5th item doesn't work here, because on small
       * screens there aren't >= 5 elements in the viewport at all times.
       * observing every second element should work.
       */
      for (let i = 0; i < this.$refs.items.length; i += 2) {
        this.intersectionObserver.observe(this.$refs.items[i]);
      }
    },
    elementIndexFromIntersectionObserverEntry(entry) {
      return parseInt(entry.target.getAttribute("data-index"));
    },
    visibilitiesChanged(entries) {
      const [smallestIndex, largestIndex] = virtualizationTools.updateVisibleElementIndices(this.visibleElementIndices, entries, this.elementIndexFromIntersectionObserverEntry);

      // we observe only every 5th item, so we increase the rendered
      // range here by 4 items in every direction just to be safe
      this.firstVisibleElementIndex = smallestIndex - 4;
      this.lastVisibileElementIndex = largestIndex + 4;
    },
    livePlayer(photo) {
      return document.querySelector("#live-player-" + photo.ID);
    },
    playLive(photo) {
      const player = this.livePlayer(photo);
      try {
        if (player) player.play();
      } catch (e) {
        // Ignore.
      }
    },
    pauseLive(photo) {
      const player = this.livePlayer(photo);
      try {
        if (player) player.pause();
      } catch (e) {
        // Ignore.
      }
    },
    downloadFile(index) {
      Notify.success(this.$gettext("Downloading…"));

      const photo = this.photos[index];
      download(`${this.$config.apiUri}/dl/${photo.Hash}?t=${this.$config.downloadToken}`, photo.FileName);
    },
    toggleLike(ev, index) {
      const inputType = this.input.eval(ev, index);

      if (inputType !== ClickShort) {
        return;
      }

      const photo = this.photos[index];

      if (!photo) {
        return;
      }

      photo.toggleLike();
    },
    onSelect(ev, index) {
      const inputType = this.input.eval(ev, index);

      if (inputType !== ClickShort) {
        return;
      }

      if (ev.shiftKey) {
        this.selectRange(index);
      } else {
        this.toggle(this.photos[index]);
      }
    },
    toggle(photo) {
      this.$clipboard.toggle(photo);
      /**
       * updating the clipboard does not rerender this component. Because of that
       * there can be scenarios where the select-icon is missing after a change,
       * for example when using touch and no hover-state changes.We therefore
       * force an update to fix that.
       */
      this.$forceUpdate();
    },
    onOpen(ev, index, showMerged, preferVideo) {
      const inputType = this.input.eval(ev, index);

      if (inputType !== ClickShort) {
        return;
      }

      this.openPhoto(index, showMerged, preferVideo);
    },
    onClick(ev, index) {
      const inputType = this.input.eval(ev, index);
      const longClick = inputType === ClickLong;

      if (inputType === InputInvalid) {
        return;
      }

      if (longClick || this.selectMode) {
        if (longClick || ev.shiftKey) {
          this.selectRange(index);
        } else {
          this.toggle(this.photos[index]);
        }
      } else {
        this.openPhoto(index);
      }
    },
    onContextMenu(ev, index) {
      if (this.$isMobile) {
        ev.preventDefault();
        ev.stopPropagation();
        this.selectRange(index);
      }
    },
    selectRange(index) {
      this.$clipboard.addRange(index, this.photos);
      /**
       * updating the clipboard does not rerender this component. Because of that
       * there can be scenarios where the select-icon is missing after a change,
       * for example when selecting multiple elements at once. We therefore
       * force an update to fix that.
       */
      this.$forceUpdate();
    },
  },
};
</script>
