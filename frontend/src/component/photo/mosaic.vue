<template>
  <v-container grid-list-xs fluid class="pa-2 p-photos p-photo-mosaic">
    <v-card v-if="photos.length === 0" class="no-results secondary-light lighten-1 ma-1" flat>
      <v-card-title primary-title>
        <div>
          <h3 v-if="filter.order === 'edited'" class="title ma-0 pa-0">
            <translate>Couldn't find recently edited</translate>
          </h3>
          <h3 v-else class="title ma-0 pa-0">
            <translate>Couldn't find anything</translate>
          </h3>
          <p class="mt-4 mb-0 pa-0">
            <translate>Try again using other filters or keywords.</translate>
            <translate>If a file you expect is missing, please re-index your library and wait until indexing has been completed.</translate>
            <template v-if="$config.feature('review')" class="mt-2 mb-0 pa-0">
              <translate>Non-photographic and low-quality images require a review before they appear in search results.</translate>
            </template>
          </p>
        </div>
      </v-card-title>
    </v-card>
    <v-layout row wrap class="search-results photo-results mosaic-view" :class="{'select-results': selectMode}">
      <v-flex
          v-for="(photo, index) in photos"
          :key="index"
          xs4 sm3 md2 lg1 d-flex
      >
        <v-card tile
                :data-id="photo.ID"
                :data-uid="photo.UID"
                style="user-select: none"
                class="result"
                :class="photo.classes()"
                @contextmenu="onContextMenu($event, index)">
          <v-img :key="photo.Hash"
                 :src="photo.thumbnailUrl('tile_224')"
                 :alt="photo.Title"
                 :title="photo.Title"
                 :transition="false"
                 aspect-ratio="1"
                 class="accent lighten-2 clickable"
                 @touchstart="onMouseDown($event, index)"
                 @touchend.stop.prevent="onClick($event, index)"
                 @mousedown="onMouseDown($event, index)"
                 @click.stop.prevent="onClick($event, index)"
                 @mouseover="playLive(photo)"
                 @mouseleave="pauseLive(photo)"
          >
            <v-layout v-if="photo.Type === 'live'" class="live-player">
              <video :id="'live-player-' + photo.ID" :key="photo.ID" width="224" height="224" preload="none"
                     loop muted playsinline>
                <source :src="photo.videoUrl()" type="video/mp4">
              </video>
            </v-layout>

            <v-btn :ripple="false" :depressed="false" class="input-open"
                   icon flat small absolute
                   @touchstart.stop.prevent="openPhoto(index, true)"
                   @touchend.stop.prevent
                   @touchmove.stop.prevent
                   @click.stop.prevent="openPhoto(index, true)">
              <v-icon color="white" class="default-hidden action-raw" :title="$gettext('RAW')">photo_camera</v-icon>
              <v-icon color="white" class="default-hidden action-live" :title="$gettext('Live')">$vuetify.icons.live_photo</v-icon>
              <v-icon color="white" class="default-hidden action-play" :title="$gettext('Video')">play_circle_fill</v-icon>
              <v-icon color="white" class="default-hidden action-stack" :title="$gettext('Stack')">burst_mode</v-icon>
            </v-btn>

            <v-btn :ripple="false" :depressed="false" class="input-view"
                   icon flat small absolute :title="$gettext('View')"
                   @touchstart.stop.prevent="openPhoto(index, false)"
                   @touchend.stop.prevent
                   @touchmove.stop.prevent
                   @click.stop.prevent="openPhoto(index, false)">
              <v-icon color="white" class="action-fullscreen">zoom_in</v-icon>
            </v-btn>

            <v-btn :ripple="false" :depressed="false" color="white" class="input-play"
                   icon flat small absolute :title="$gettext('Play')"
                   @touchstart.stop.prevent="openPhoto(index, true)"
                   @touchend.stop.prevent
                   @touchmove.stop.prevent
                   @click.stop.prevent="openPhoto(index, true)">
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
                   @touchstart.stop.prevent="onSelect($event, index)"
                   @touchend.stop.prevent
                   @touchmove.stop.prevent
                   @click.stop.prevent="onSelect($event, index)">
              <v-icon color="white" class="select-on">check_circle</v-icon>
              <v-icon color="white" class="select-off">radio_button_off</v-icon>
            </v-btn>

            <v-btn :ripple="false"
                   icon flat small absolute
                   class="input-favorite"
                   @touchstart.stop.prevent="photo.toggleLike()"
                   @touchend.stop.prevent
                   @touchmove.stop.prevent
                   @click.stop.prevent="photo.toggleLike()">
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
export default {
  name: 'PPhotoMosaic',
  props: {
    photos: Array,
    openPhoto: Function,
    editPhoto: Function,
    album: Object,
    filter: Object,
    context: String,
    selectMode: Boolean,
  },
  data() {
    return {
      hidePrivate: this.$config.settings().features.private,
      mouseDown: {
        index: -1,
        scrollY: window.scrollY,
        timeStamp: -1,
      },
    };
  },
  methods: {
    livePlayer(photo) {
      return document.querySelector("#live-player-" + photo.ID);
    },
    playLive(photo) {
      const player = this.livePlayer(photo);
      if (player) player.play();
    },
    pauseLive(photo) {
      const player = this.livePlayer(photo);
      if (player) player.pause();
    },
    onSelect(ev, index) {
      if (ev.shiftKey) {
        this.selectRange(index);
      } else {
        this.toggle(this.photos[index]);
      }
    },
    onMouseDown(ev, index) {
      this.mouseDown.index = index;
      this.mouseDown.scrollY = window.scrollY;
      this.mouseDown.timeStamp = ev.timeStamp;
    },
    toggle(photo) {
      this.$clipboard.toggle(photo);
    },
    onClick(ev, index) {
      const longClick = (this.mouseDown.index === index && ev.timeStamp - this.mouseDown.timeStamp > 400);
      const scrolled = (this.mouseDown.scrollY - window.scrollY) !== 0;

      if (scrolled) {
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
