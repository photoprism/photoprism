<template>
  <v-container grid-list-xs fluid class="pa-2 p-photos p-photo-mosaic">
    <template v-if="photos.length === 0">
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
    </template>
    <v-layout row wrap class="search-results photo-results mosaic-view" :class="{'select-results': selectMode}">
      <v-flex
          v-for="(photo, index) in photos"
          ref="items"
          :key="photo.ID"
          :data-index="index"
          xs4 sm3 md2 lg1 d-flex
      >

        <div v-if="index < firstVisibleElementIndex || index > lastVisibileElementIndex" 
                style="user-select: none; aspect-ratio: 1"
                class="accent lighten-2 result"/>
        <v-hover v-if="index >= firstVisibleElementIndex && index <= lastVisibileElementIndex">
          <v-card slot-scope="{ hover }"
                  tile
                  :data-id="photo.ID"
                  :data-uid="photo.UID"
                  style="user-select: none; aspect-ratio: 1"
                  class="accent lighten-2 result"
                  :class="photo.classes()"
                  @contextmenu.stop="onContextMenu($event, index)">
            <v-img
                  :key="photo.Hash"
                  :src="photo.thumbnailUrl('tile_224')"
                  :alt="photo.Title"
                  :title="photo.Title"
                  :transition="false"
                  aspect-ratio="1"
                  class="clickable"
                  @touchstart.passive="input.touchStart($event, index)"
                  @touchend.stop.prevent="onClick($event, index)"
                  @mousedown.stop.prevent="input.mouseDown($event, index)"
                  @click.stop.prevent="onClick($event, index)"
                  @mouseover="playLive(photo)"
                  @mouseleave="pauseLive(photo)"
            >
              <v-layout v-if="photo.Type === 'live' || photo.Type === 'animated'" class="live-player">
                <video :id="'live-player-' + photo.ID" :key="photo.ID" width="224" height="224" preload="none"
                      loop muted playsinline>
                  <source :src="photo.videoUrl()">
                </video>
              </v-layout>

              <v-btn v-if="photo.Type !== 'image' || photo.Files.length > 1"
                    :ripple="false" :depressed="false" class="input-open"
                    icon flat small absolute
                    @touchstart.stop.prevent="input.touchStart($event, index)"
                    @touchend.stop.prevent="onOpen($event, index, true)"
                    @touchmove.stop.prevent
                    @click.stop.prevent="onOpen($event, index, true)">
                <v-icon v-if="photo.Type === 'raw'" color="white" class="action-raw" :title="$gettext('RAW')">photo_camera</v-icon>
                <v-icon v-if="photo.Type === 'live'" color="white" class="action-live" :title="$gettext('Live')">$vuetify.icons.live_photo</v-icon>
                <v-icon v-if="photo.Type === 'animated'" color="white" class="action-animated" :title="$gettext('Animated')">gif</v-icon>
                <v-icon v-if="photo.Type === 'video'" color="white" class="action-play" :title="$gettext('Video')">play_arrow</v-icon>
                <v-icon v-if="photo.Type === 'image'" color="white" class="action-stack" :title="$gettext('Stack')">burst_mode</v-icon>
              </v-btn>

              <v-btn v-if="photo.Type === 'image' && selectMode"
                    :ripple="false" :depressed="false" class="input-view"
                    icon flat small absolute :title="$gettext('View')"
                    @touchstart.stop.prevent="input.touchStart($event, index)"
                    @touchend.stop.prevent="onOpen($event, index, false)"
                    @touchmove.stop.prevent
                    @click.stop.prevent="onOpen($event, index, false)">
                <v-icon color="white" class="action-fullscreen">zoom_in</v-icon>
              </v-btn>

              <v-btn v-if="!isSharedView && hidePrivate && photo.Private" :ripple="false"
                    icon flat small absolute
                    class="input-private">
                <v-icon color="white" class="select-on">lock</v-icon>
              </v-btn>

              <v-btn v-if="hover || $clipboard.has(photo)"
                    :ripple="false"
                    icon flat small absolute
                    class="input-select"
                    @touchstart.stop.prevent="input.touchStart($event, index)"
                    @touchend.stop.prevent="onSelect($event, index)"
                    @touchmove.stop.prevent
                    @click.stop.prevent="onSelect($event, index)">
                <v-icon v-if="$clipboard.has(photo)" color="white" class="select-on">check_circle</v-icon>
                <v-icon v-else color="white" class="select-off">radio_button_off</v-icon>
              </v-btn>

              <v-btn v-if="!isSharedView"
                    :ripple="false"
                    icon flat small absolute
                    class="input-favorite"
                    @touchstart.stop.prevent="input.touchStart($event, index)"
                    @touchend.stop.prevent="toggleLike($event, index)"
                    @touchmove.stop.prevent
                    @click.stop.prevent="toggleLike($event, index)">
                <v-icon v-if="photo.Favorite" color="white" class="select-on">favorite</v-icon>
                <v-icon v-else color="white" class="select-off">favorite_border</v-icon>
              </v-btn>
            </v-img>
          </v-card>
        </v-hover>
      </v-flex>
    </v-layout>
  </v-container>
</template>
<script>
import {Input, InputInvalid, ClickShort, ClickLong} from "common/input";
import {virtualizationTools} from 'common/virtualization-tools';

export default {
  name: 'PPhotoMosaic',
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
  beforeCreate() {
    this.intersectionObserver = new IntersectionObserver((entries) => {
      this.visibilitiesChanged(entries);
    }, {
      rootMargin: "50% 0px",
    });
  },
  mounted() {
    this.observeItems();
  },
  updated() {
    this.observeItems();
  },
  beforeDestroy() {
    this.intersectionObserver.disconnect();
  },
  methods: {
    observeItems() {
      if (this.$refs.items === undefined) {
        return;
      }
      for (const item of this.$refs.items) {
        this.intersectionObserver.observe(item);
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

      this.firstVisibleElementIndex = smallestIndex;
      this.lastVisibileElementIndex = largestIndex;
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
    },
    onOpen(ev, index, showMerged) {
      const inputType = this.input.eval(ev, index);

      if (inputType !== ClickShort) {
        return;
      }

      this.openPhoto(index, showMerged);
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
        this.openPhoto(index, false);
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
    }
  },
};
</script>
