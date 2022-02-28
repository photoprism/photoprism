<template>
  <v-container grid-list-xs fluid class="pa-2 p-photos p-photo-mosaic">
    <v-alert
        :value="photos.length === 0"
        color="secondary-dark" icon="lightbulb_outline" class="no-results ma-2 opacity-70" outline
    >
      <h3 v-if="filter.order === 'edited'" class="body-2 ma-0 pa-0">
        <translate>No recently edited pictures</translate>
      </h3>
      <h3 v-else class="body-2 ma-0 pa-0">
        <translate>No pictures found</translate>
      </h3>
      <p class="body-1 mt-2 mb-0 pa-0">
        <translate>Try again using other filters or keywords.</translate>
        <translate>In case pictures you expect are missing, please rescan your library and wait until indexing has been completed.</translate>
        <template v-if="$config.feature('review')" class="mt-2 mb-0 pa-0">
          <translate>Non-photographic and low-quality images require a review before they appear in search results.</translate>
        </template>
      </p>
    </v-alert>
    <v-layout row wrap class="search-results photo-results mosaic-view" :class="{'select-results': selectMode}">
      <v-flex
          v-for="(photo, index) in photos"
          :key="photo.ID"
          xs4 sm3 md2 lg1 d-flex
      >
        <v-card tile
                :data-id="photo.ID"
                :data-uid="photo.UID"
                style="user-select: none"
                class="result"
                :class="photo.classes()"
                @contextmenu.stop="onContextMenu($event, index)">
          <v-img :key="photo.Hash"
                 :src="photo.thumbnailUrl('tile_224')"
                 :alt="photo.Title"
                 :title="photo.Title"
                 :transition="false"
                 aspect-ratio="1"
                 class="accent lighten-2 clickable"
                 @touchstart.passive="input.touchStart($event, index)"
                 @touchend.stop.prevent="onClick($event, index)"
                 @mousedown.stop.prevent="input.mouseDown($event, index)"
                 @click.stop.prevent="onClick($event, index)"
                 @mouseover="playLive(photo)"
                 @mouseleave="pauseLive(photo)"
          >
            <v-layout v-if="photo.Type === 'live'" class="live-player">
              <video :id="'live-player-' + photo.ID" :key="photo.ID" width="224" height="224" preload="none"
                     loop muted playsinline>
                <source :src="photo.videoUrl()">
              </video>
            </v-layout>

            <v-btn :ripple="false" :depressed="false" class="input-open"
                   icon flat small absolute
                   @touchstart.stop.prevent="input.touchStart($event, index)"
                   @touchend.stop.prevent="onOpen($event, index, true)"
                   @touchmove.stop.prevent
                   @click.stop.prevent="onOpen($event, index, true)">
              <v-icon color="white" class="default-hidden action-raw" :title="$gettext('RAW')">photo_camera</v-icon>
              <v-icon color="white" class="default-hidden action-live" :title="$gettext('Live')">$vuetify.icons.live_photo</v-icon>
              <v-icon color="white" class="default-hidden action-play" :title="$gettext('Video')">play_arrow</v-icon>
              <v-icon color="white" class="default-hidden action-stack" :title="$gettext('Stack')">burst_mode</v-icon>
            </v-btn>

            <v-btn :ripple="false" :depressed="false" class="input-view"
                   icon flat small absolute :title="$gettext('View')"
                   @touchstart.stop.prevent="input.touchStart($event, index)"
                   @touchend.stop.prevent="onOpen($event, index, false)"
                   @touchmove.stop.prevent
                   @click.stop.prevent="onOpen($event, index, false)">
              <v-icon color="white" class="action-fullscreen">zoom_in</v-icon>
            </v-btn>

            <v-btn :ripple="false" :depressed="false" color="white" class="input-play"
                   icon flat small absolute :title="$gettext('Play')"
                   @touchstart.stop.prevent="input.touchStart($event, index)"
                   @touchend.stop.prevent="onOpen($event, index, true)"
                   @touchmove.stop.prevent
                   @click.stop.prevent="onOpen($event, index, true)">
              <v-icon color="white" class="action-play">play_arrow</v-icon>
            </v-btn>

            <v-btn v-if="hidePrivate" :ripple="false"
                   icon flat small absolute
                   class="input-private">
              <v-icon color="white" class="select-on">lock</v-icon>
            </v-btn>

            <v-btn :ripple="false"
                   icon flat small absolute
                   class="input-select"
                   @touchstart.stop.prevent="input.touchStart($event, index)"
                   @touchend.stop.prevent="onSelect($event, index)"
                   @touchmove.stop.prevent
                   @click.stop.prevent="onSelect($event, index)">
              <v-icon color="white" class="select-on">check_circle</v-icon>
              <v-icon color="white" class="select-off">radio_button_off</v-icon>
            </v-btn>

            <v-btn :ripple="false"
                   icon flat small absolute
                   class="input-favorite"
                   @touchstart.stop.prevent="input.touchStart($event, index)"
                   @touchend.stop.prevent="toggleLike($event, index)"
                   @touchmove.stop.prevent
                   @click.stop.prevent="toggleLike($event, index)">
              <v-icon color="white" class="select-on">favorite</v-icon>
              <v-icon color="white" class="select-off">favorite_border</v-icon>
            </v-btn>
          </v-img>
        </v-card>
      </v-flex>
    </v-layout>
  </v-container>
</template>
<script>
import {Input, InputInvalid, ClickShort, ClickLong} from "common/input";

export default {
  name: 'PPhotoMosaic',
  props: {
    photos: {
      type: Array,
      default: () => [],
    },
    openPhoto: Function,
    editPhoto: Function,
    album: {
      type: Object,
      default: () => {},
    },
    filter: {
      type: Object,
      default: () => {},
    },
    context: String,
    selectMode: Boolean,
  },
  data() {
    return {
      hidePrivate: this.$config.settings().features.private,
      input: new Input(),
    };
  },
  methods: {
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
