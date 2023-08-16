<template>
  <v-container grid-list-xs fluid class="pa-2 p-photos p-photo-mosaic">
    <div v-if="photos.length === 0" class="pa-0">
      <v-alert
          :value="true"
          color="secondary-dark"
          :icon="isSharedView ? 'image_not_supported' : 'lightbulb_outline'"
          class="no-results ma-2 opacity-70"
          outline
      >
        <h3 v-if="filter.order === 'edited'" class="body-2 ma-0 pa-0">
          <translate>No recently edited pictures</translate>
        </h3>
        <h3 v-else class="body-2 ma-0 pa-0">
          <translate>No pictures found</translate>
        </h3>
        <p class="body-1 mt-2 mb-0 pa-0">
          <translate>Try again using other filters or keywords.</translate>
          <template v-if="!isSharedView">
            <translate>In case pictures you expect are missing, please rescan your library and wait until indexing has been completed.</translate>
            <template v-if="$config.feature('review')">
              <translate>Non-photographic and low-quality images require a review before they appear in search results.</translate>
            </template>
          </template>
        </p>
      </v-alert>
    </div>
    <v-layout row wrap class="search-results photo-results mosaic-view" :class="{'select-results': selectMode}">
      <div
          v-for="(photo, index) in photos"
          ref="items"
          :key="photo.ID"
          class="flex xs4 sm3 md2 lg1"
          :data-index="index"
      >
       <!--
         The following div is the layout + size container. It makes the browser not
         re-layout all elements in the list when the children of one of them changes
        -->
        <div class="image-container">
          <div v-if="index < firstVisibleElementIndex || index > lastVisibileElementIndex"
               :data-uid="photo.UID"
               class="card darken-1 result image"
          />
          <div  v-else
                :key="photo.Hash"
                tile
                :data-id="photo.ID"
                :data-uid="photo.UID"
                :style="`background-image: url(${photo.thumbnailUrl('tile_224')})`"
                :class="photo.classes().join(' ') + ' card darken-1 result clickable image'"
                :alt="photo.Title"
                :title="photo.Title"
                @contextmenu.stop="onContextMenu($event, index)"
                @touchstart.passive="input.touchStart($event, index)"
                @touchend.stop.prevent="onClick($event, index)"
                @mousedown.stop.prevent="input.mouseDown($event, index)"
                @click.stop.prevent="onClick($event, index)"
                @mouseover="playLive(photo)"
                @mouseleave="pauseLive(photo)">
            <v-layout v-if="photo.Type === 'live' || photo.Type === 'animated'" class="live-player">
              <video :id="'live-player-' + photo.ID" :key="photo.ID" width="224" height="224" preload="none"
                     loop muted playsinline>
                <source :src="photo.videoUrl()">
              </video>
            </v-layout>

            <button v-if="photo.Type !== 'image' || photo.Files.length > 1"
                  class="input-open"
                  @touchstart.stop.prevent="input.touchStart($event, index)"
                  @touchend.stop.prevent="onOpen($event, index, !isSharedView, photo.Type === 'live')"
                  @touchmove.stop.prevent
                  @click.stop.prevent="onOpen($event, index, !isSharedView, photo.Type === 'live')">
              <i v-if="photo.Type === 'raw'" class="action-raw" :title="$gettext('RAW')">raw_on</i>
              <i v-if="photo.Type === 'live'" class="action-live" :title="$gettext('Live')"><icon-live-photo/></i>
              <i v-if="photo.Type === 'video'" class="action-play" :title="$gettext('Video')">play_arrow</i>
              <i v-if="photo.Type === 'animated'" class="action-animated" :title="$gettext('Animated')">gif</i>
              <i v-if="photo.Type === 'vector'" class="action-vector" :title="$gettext('Vector')">font_download</i>
              <i v-if="photo.Type === 'image'" class="action-stack" :title="$gettext('Stack')">burst_mode</i>
            </button>

            <button v-if="photo.Type === 'image' && selectMode"
                  class="input-view"
                  :title="$gettext('View')"
                  @touchstart.stop.prevent="input.touchStart($event, index)"
                  @touchend.stop.prevent="onOpen($event, index)"
                  @touchmove.stop.prevent
                  @click.stop.prevent="onOpen($event, index)">
              <i color="white" class="action-fullscreen">zoom_in</i>
            </button>

            <button v-if="!isSharedView && hidePrivate && photo.Private" class="input-private">
              <i color="white" class="select-on">lock</i>
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
            <button
                  class="input-select"
                  @mousedown.stop.prevent="input.mouseDown($event, index)"
                  @touchstart.stop.prevent="input.touchStart($event, index)"
                  @touchend.stop.prevent="onSelect($event, index)"
                  @touchmove.stop.prevent
                  @click.stop.prevent="onSelect($event, index)">
              <i color="white" class="select-on">check_circle</i>
              <i color="white" class="select-off">radio_button_off</i>
            </button>

            <button v-if="!isSharedView"
                class="input-favorite"
                @touchstart.stop.prevent="input.touchStart($event, index)"
                @touchend.stop.prevent="toggleLike($event, index)"
                @touchmove.stop.prevent
                @click.stop.prevent="toggleLike($event, index)"
            >
              <i v-if="photo.Favorite">favorite</i>
              <i v-else>favorite_border</i>
            </button>
          </div>
        </div>
      </div>
    </v-layout>
  </v-container>
</template>
<script>
import {Input, InputInvalid, ClickShort, ClickLong} from "common/input";
import {virtualizationTools} from 'common/virtualization-tools';
import IconLivePhoto from "component/icon/live-photo.vue";

export default {
  name: 'PPhotoMosaic',
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
      default:() => {},
    },
    editPhoto: {
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
    return {
      hidePrivate: this.$config.settings().features.private,
      input: new Input(),
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
    }
  },
  beforeCreate() {
    this.intersectionObserver = new IntersectionObserver((entries) => {
      this.visibilitiesChanged(entries);
    }, {
      rootMargin: "50% 0px",
    });
  },
  beforeDestroy() {
    this.intersectionObserver.disconnect();
  },
  methods: {
    observeItems() {
      if (this.$refs.items === undefined) {
        return;
      }

      /**
       * observing only every 5th item reduces the amount of time
       * spent computing intersection by 80%. me might render up to
       * 8 items more than required, but the time saved computing
       * intersections is far greater than the time lost rendering
       * a couple more items
       */
      for (let i = 0; i < this.$refs.items.length; i += 5) {
        this.intersectionObserver.observe(this.$refs.items[i]);
      }
    },
    elementIndexFromIntersectionObserverEntry(entry) {
      return parseInt(entry.target.getAttribute('data-index'));
    },
    visibilitiesChanged(entries) {
      const [smallestIndex, largestIndex] = virtualizationTools.updateVisibleElementIndices(
        this.visibleElementIndices,
        entries,
        this.elementIndexFromIntersectionObserverEntry,
      );

      // we observe only every 5th item, so we increase the rendered
      // range here by 4 items in every directio just to be safe
      this.firstVisibleElementIndex = smallestIndex - 4;
      this.lastVisibileElementIndex = largestIndex + 4;
    },
    livePlayer(photo) {
      return document.querySelector("#live-player-" + photo.ID);
    },
    playLive(photo) {
      const player = this.livePlayer(photo);
      try { if (player) player.play(); }
      catch (e) {
        // Ignore.
      }
    },
    pauseLive(photo) {
      const player = this.livePlayer(photo);
      try { if (player) player.pause(); }
      catch (e) {
        // Ignore.
      }
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
      this.$forceUpdate();
    }
  },
};
</script>
