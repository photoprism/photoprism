<template>
  <v-container grid-list-xs fluid class="pa-2 p-photos p-photo-cards">
    <v-alert
        :value="photos.length === 0"
        color="secondary-dark" icon="image_not_supported" class="no-results ma-2 opacity-70" outline
    >
      <h3 v-if="filter.order === 'edited'" class="body-2 ma-0 pa-0">
        <translate>No recently edited pictures</translate>
      </h3>
      <h3 v-else class="body-2 ma-0 pa-0">
        <translate>No pictures found</translate>
      </h3>
      <p class="body-1 mt-2 mb-0 pa-0">
        <translate>Try again using other filters or keywords.</translate>
      </p>
    </v-alert>
    <v-layout row wrap class="search-results photo-results cards-view" :class="{'select-results': selectMode}">
      <v-flex
          v-for="(photo, index) in photos"
          :key="photo.ID"
          xs12 sm6 md4 lg3 xlg2 xxxl1 d-flex
      >
        <v-card tile
                :data-id="photo.ID"
                :data-uid="photo.UID"
                style="user-select: none"
                class="result accent lighten-3"
                :class="photo.classes()"
                @contextmenu.stop="onContextMenu($event, index)">
          <div class="card-background accent lighten-3"></div>
          <v-img :key="photo.Hash"
                 :src="photo.thumbnailUrl('tile_500')"
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
            <v-layout v-if="photo.Type === 'live' || photo.Type === 'animated'" class="live-player">
              <video :id="'live-player-' + photo.ID" :key="photo.ID" width="500" height="500" preload="none"
                     loop muted playsinline>
                <source :src="photo.videoUrl()">
              </video>
            </v-layout>

            <v-btn :ripple="false" :depressed="false" class="input-open"
                   icon flat absolute
                   @touchstart.stop.prevent="input.touchStart($event, index)"
                   @touchend.stop.prevent="onOpen($event, index, true)"
                   @touchmove.stop.prevent
                   @click.stop.prevent="onOpen($event, index, true)">
              <v-icon color="white" class="default-hidden action-raw" :title="$gettext('RAW')">photo_camera</v-icon>
              <v-icon color="white" class="default-hidden action-live" :title="$gettext('Live')">$vuetify.icons.live_photo</v-icon>
              <v-icon color="white" class="default-hidden action-animated" :title="$gettext('Animated')">gif</v-icon>
              <v-icon color="white" class="default-hidden action-play" :title="$gettext('Video')">play_arrow</v-icon>
              <v-icon color="white" class="default-hidden action-stack" :title="$gettext('Stack')">burst_mode</v-icon>
            </v-btn>

            <v-btn :ripple="false" :depressed="false" class="input-view"
                   icon flat absolute :title="$gettext('View')"
                   @touchstart.stop.prevent="input.touchStart($event, index)"
                   @touchend.stop.prevent="onOpen($event, index, false)"
                   @touchmove.stop.prevent
                   @click.stop.prevent="onOpen($event, index, true)">
              <v-icon color="white" class="action-fullscreen">zoom_in</v-icon>
            </v-btn>

            <v-btn :ripple="false" :depressed="false" color="white" class="input-play"
                   outline fab large absolute :title="$gettext('Play')"
                   @touchstart.stop.prevent="input.touchStart($event, index)"
                   @touchend.stop.prevent="onOpen($event, index, true)"
                   @touchmove.stop.prevent
                   @click.stop.prevent="onOpen($event, index, true)">
              <v-icon color="white" class="action-play">play_arrow</v-icon>
            </v-btn>

            <v-btn :ripple="false"
                   icon flat absolute
                   class="input-select"
                   @touchstart.stop.prevent="input.touchStart($event, index)"
                   @touchend.stop.prevent="onSelect($event, index)"
                   @touchmove.stop.prevent
                   @click.stop.prevent="onSelect($event, index)">
              <v-icon color="white" class="select-on">check_circle</v-icon>
              <v-icon color="white" class="select-off">radio_button_off</v-icon>
            </v-btn>
          </v-img>

          <v-card-title primary-title class="pa-3 card-details" style="user-select: none;">
            <div>
              <h3 class="body-2 mb-2" :title="photo.Title">
                <div @click.stop.prevent="openPhoto(index, false)">
                  {{ photo.Title | truncate(80) }}
                </div>
              </h3>
              <div v-if="photo.Description" class="caption mb-2" :title="$gettext('Description')">
                <div>
                  {{ photo.Description }}
                </div>
              </div>
              <div class="caption">
                <div>
                  <v-icon size="14" :title="$gettext('Taken')">date_range</v-icon>
                  {{ photo.getDateString(true) }}
                </div>
                <template v-if="!photo.Description">
                  <div v-if="photo.Type === 'video'" :title="$gettext('Video')">
                    <v-icon size="14">movie</v-icon>
                    {{ photo.getVideoInfo() }}
                  </div>
                  <div v-else-if="photo.Type === 'animated'" :title="$gettext('Animated')+' GIF'">
                    <v-icon size="14">gif_box</v-icon>
                    {{ photo.getVideoInfo() }}
                  </div>
                  <div v-else :title="$gettext('Camera')">
                    <v-icon size="14">photo_camera</v-icon>
                    {{ photo.getPhotoInfo() }}
                  </div>
                </template>
                <template v-if="filter.order === 'name' && $config.feature('download')">
                  <div :title="$gettext('Name')">
                    <v-icon size="14">insert_drive_file</v-icon>
                    {{ photo.baseName() }}
                  </div>
                </template>
                <template v-if="featPlaces && photo.Country !== 'zz'">
                  <div :title="$gettext('Location')">
                    <v-icon size="14">location_on</v-icon>
                    {{ photo.locationInfo() }}
                  </div>
                </template>
              </div>
            </div>
          </v-card-title>
        </v-card>
      </v-flex>
    </v-layout>
  </v-container>
</template>
<script>
import download from "common/download";
import Notify from "common/notify";
import {Input, InputInvalid, ClickShort, ClickLong} from "common/input";

export default {
  name: 'PPhotoCards',
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
    openLocation: {
      type: Function,
      default: () => {},
    },
    album: {
      type: Object,
      default: () => {
      },
    },
    filter: {
      type: Object,
      default: () => {
      },
    },
    context: {
      type: String,
      default: "",
    },
    selectMode: Boolean,
  },
  data() {
    const featPlaces = this.$config.settings().features.places;
    const input = new Input();
    const debug = this.$config.get('debug');

    return {
      featPlaces,
      debug,
      input,
    };
  },
  methods: {
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
      download(`${this.$config.apiUri}/dl/${photo.Hash}?t=${this.$config.downloadToken()}`, photo.FileName);
    },
    onSelect(ev, index) {
      const inputType = this.input.eval(ev, index);

      if (inputType !== ClickShort) {
        return;
      }

      if (ev.shiftKey) {
        this.selectRange(index);
      } else {
        this.$clipboard.toggle(this.photos[index]);
      }
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
          this.$clipboard.toggle(this.photos[index]);
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
    },
  }
};
</script>
