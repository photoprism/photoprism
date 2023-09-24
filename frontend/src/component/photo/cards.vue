<template>
  <v-container grid-list-xs fluid class="pa-2 p-photos p-photo-cards">
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
    <v-layout row wrap class="search-results photo-results cards-view" :class="{'select-results': selectMode}">
      <div
          v-for="(photo, index) in photos"
          ref="items"
          :key="photo.ID"
          :data-index="index"
          class="flex xs12 sm6 md4 lg3 xlg2 xxxl1 d-flex"
      >
        <div v-if="index < firstVisibleElementIndex || index > lastVisibileElementIndex"
             :data-uid="photo.UID"
             class="card result placeholder"
        >
          <div class="card darken-1 image"/>
          <div v-if="photo.Quality < 3 && context === 'review'" style="width: 100%; height: 34px"/>
          <div class="pa-3 card-details">
            <div>
              <h3 class="body-2 mb-2" :title="photo.Title">
                {{ photo.Title | truncate(80) }}
              </h3>
              <div v-if="photo.Description" class="caption mb-2">
                {{ photo.Description }}
              </div>
              <div class="caption">
                  <i/>
                  {{ photo.getDateString(true) }}
                <br>
                <i/>
                <template v-if="photo.Type === 'video' || photo.Type === 'animated'">
                  {{ photo.getVideoInfo() }}
                </template>
                <template v-else>
                  {{ photo.getPhotoInfo() }}
                </template>
                <template v-if="filter.order === 'name' && $config.feature('download')">
                  <br>
                  <i/>
                  {{ photo.baseName() }}
                </template>
                <template v-if="featPlaces && photo.Country !== 'zz'">
                  <br>
                  <i/>
                  {{ photo.locationInfo() }}
                </template>
              </div>
            </div>
          </div>
        </div>
        <div v-else
              :data-id="photo.ID"
              :data-uid="photo.UID"
              class="result card"
              :class="photo.classes()"
              @contextmenu.stop="onContextMenu($event, index)">
          <div class="card-background card"></div>
          <div :key="photo.Hash"
                :title="photo.Title"
                class="card darken-1 clickable image"
                :style="`background-image: url(${photo.thumbnailUrl('tile_500')})`"
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
              <i class="action-fullscreen">zoom_in</i>
            </button>

            <button v-if="!isSharedView && featPrivate && photo.Private" class="input-private">
              <i class="select-on">lock</i>
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
                  @touchstart.stop.prevent="input.touchStart($event, index)"
                  @touchend.stop.prevent="onSelect($event, index)"
                  @touchmove.stop.prevent
                  @click.stop.prevent="onSelect($event, index)">
              <i class="select-on">check_circle</i>
              <i class="select-off">radio_button_off</i>
            </button>

            <button
                  v-if="!isSharedView"
                  class="input-favorite"
                  @touchstart.stop.prevent="input.touchStart($event, index)"
                  @touchend.stop.prevent="toggleLike($event, index)"
                  @touchmove.stop.prevent
                  @click.stop.prevent="toggleLike($event, index)">
              <i v-if="photo.Favorite">favorite</i>
              <i v-else>favorite_border</i>
            </button>
          </div>

          <v-card-actions v-if="!isSharedView && photo.Quality < 3 && context === 'review'" class="card-details pa-0">
            <v-layout row wrap align-center>
              <v-flex xs6 class="text-xs-center pa-1">
                <v-btn color="card darken-1"
                      small depressed dark block :round="false"
                      class="action-archive text-xs-center"
                      :title="$gettext('Archive')" @click.stop="photo.archive()">
                  <v-icon dark>clear</v-icon>
                </v-btn>
              </v-flex>
              <v-flex xs6 class="text-xs-center pa-1">
                <v-btn color="card darken-1"
                      small depressed dark block :round="false"
                      class="action-approve text-xs-center"
                      :title="$gettext('Approve')" @click.stop="photo.approve()">
                  <v-icon dark>check</v-icon>
                </v-btn>
              </v-flex>
            </v-layout>
          </v-card-actions>

          <div class="pa-3 card-details">
            <div>
              <h3 class="body-2 mb-2" :title="photo.Title">
                <button class="action-title-edit" :data-uid="photo.UID"
                        @click.exact="isSharedView ? openPhoto(index) : editPhoto(index)">
                  {{ photo.Title | truncate(80) }}
                </button>
              </h3>
              <div v-if="photo.Description" class="caption mb-2" :title="$gettext('Description')">
                <button @click.exact="editPhoto(index)">
                  {{ photo.Description }}
                </button>
              </div>
              <div class="caption">
                <button class="action-date-edit" :data-uid="photo.UID"
                        @click.exact="editPhoto(index)">
                  <i :title="$gettext('Taken')">date_range</i>
                  {{ photo.getDateString(true) }}
                </button>
                <br>
                <button v-if="photo.Type === 'video'" :title="$gettext('Video')"
                        @click.exact="openPhoto(index)">
                  <i>movie</i>
                  {{ photo.getVideoInfo() }}
                </button>
                <button v-else-if="photo.Type === 'live'" :title="$gettext('Live')"
                        @click.exact="openPhoto(index)">
                  <i>play_circle</i>
                  {{ photo.getVideoInfo() }}
                </button>
                <button v-else-if="photo.Type === 'animated'" :title="$gettext('Animated')+' GIF'"
                        @click.exact="openPhoto(index)">
                  <i>gif_box</i>
                  {{ photo.getVideoInfo() }}
                </button>
                <button v-else-if="photo.Type === 'vector'" :title="$gettext('Vector')"
                        @click.exact="openPhoto(index)">
                  <i>font_download</i>
                  {{ photo.getVectorInfo() }}
                </button>
                <button v-else :title="$gettext('Camera')" class="action-camera-edit"
                        :data-uid="photo.UID" @click.exact="editPhoto(index)">
                  <i>photo_camera</i>
                  {{ photo.getPhotoInfo() }}
                </button>
                <button v-if="photo.LensID > 1" :title="$gettext('Lens')" class="action-lens-edit"
                        :data-uid="photo.UID" @click.exact="editPhoto(index)">
                  <i>camera</i>
                  {{ photo.LensModel }}
                </button>
                <template v-if="filter.order === 'name' && $config.feature('download')">
                  <br>
                  <button :title="$gettext('Name')"
                          @click.exact="downloadFile(index)">
                    <i>insert_drive_file</i>
                    {{ photo.baseName() }}
                  </button>
                </template>
                <template v-if="featPlaces && photo.Country !== 'zz'">
                  <br>
                  <button :title="$gettext('Location')" class="action-location"
                          :data-uid="photo.UID" @click.exact="openLocation(index)">
                    <i>location_on</i>
                    {{ photo.locationInfo() }}
                  </button>
                </template>
              </div>
            </div>
          </div>
        </div>
      </div>
    </v-layout>
  </v-container>
</template>
<script>
import download from "common/download";
import Notify from "common/notify";
import {Input, InputInvalid, ClickShort, ClickLong} from "common/input";
import {virtualizationTools} from 'common/virtualization-tools';
import IconLivePhoto from "component/icon/live-photo.vue";

export default {
  name: 'PPhotoCards',
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
    const debug = this.$config.get('debug');

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
       * observing only every 5th item doesn't work here, because on small
       * screens there aren't >= 5 elements in the viewport at all times.
       * observing every second element should work.
       */
      for (let i = 0; i < this.$refs.items.length; i += 2) {
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
  }
};
</script>
