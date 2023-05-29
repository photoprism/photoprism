<template>
  <v-container grid-list-xs fluid class="pa-2 p-photos p-photo-mosaic">
    <template v-if="visiblePhotos.length === 0">
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
    <v-layout ref="container" row wrap class="search-results photo-results mosaic-view" :class="{'select-results': selectMode}" :style="`position: relative; height: ${scrollHeight}px`">
      <div
          v-for="(photo, index) in visiblePhotos"
          ref="items"
          :key="photo.ID"
          class="flex xs4 sm3 md2 lg1 image-container"
          :data-index="index"
          :style="getVirtualizedElementStyle(index)"
      >
        <div :key="photo.Hash"
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
      firstElementToRender: 1,
      lastElementToRender: 0,
      visiblePhotos: {},
      elementSize: 100, // no need to differentiate width and height, because the images are square.
      containerWidth: 100,
      containerHeight: window.innerHeight,
      containerTop: 0,
      scrollHeight: 100,
      scrollPos: 0,
      columnCount: 1,
      visibleElementIndices: new Set(),
    };
  },
  watch: {
    photos: {
      handler() {
        this.$nextTick(() => {
          this.observeItems();
        });
        if (this.visiblePhotos[0] === undefined) {
          this.visiblePhotos[0] = this.photos[0];
        }
      },
      immediate: true,
    }
  },
  beforeCreate() {
    this.elementObserver = new ResizeObserver((entries) => {
      this.elementSize = entries[0].borderBoxSize[0].inlineSize;
      this.updateGeometry();
    });
    this.containerObserver = new ResizeObserver((entries) => {
      this.containerWidth = entries[0].contentRect.width;
      this.updateGeometry();
    });
  },
  created() {
    window.addEventListener('scroll', this.handleScroll);
    window.addEventListener('resize', this.handleResize);
  },
  mounted() {
    this.containerTop = this.$refs.container.getBoundingClientRect().top;
    this.updateGeometry();
  },
  beforeDestroy() {
    window.removeEventListener('scroll', this.handleScroll);
    window.removeEventListener('resize', this.handleResize);
    this.elementObserver.disconnect();
    this.containerObserver.disconnect();
  },
  methods: {
    observeItems() {
      if (this.$refs.items === undefined) {
        return;
      }

      this.elementObserver.observe(this.$refs.items[0]);
      if (this.$refs.container !== undefined) {
        this.containerObserver.observe(this.$refs.container);
      }
    },
    handleScroll(event) {
      this.scrollPos = document.scrollingElement.scrollTop;
      this.updateGeometry();
    },
    handleResize(event) {
      this.containerHeight = window.innerHeight;
      this.updateGeometry();
    },
    updateGeometry() {
      const {
        visibleColumnCount,
        firstElementToRender,
        lastElementToRender,
        totalScrollHeight,
      } = virtualizationTools.getVisibleRange(
        this.photos,
        this.containerWidth,
        this.containerHeight,
        this.elementSize,
        this.elementSize,
        this.scrollPos - this.containerTop,
        2,
      );

      this.columnCount = visibleColumnCount;
      this.scrollHeight = totalScrollHeight;

      if (this.firstElementToRender !== firstElementToRender || this.lastElementToRender !== lastElementToRender) {
        this.firstElementToRender = firstElementToRender;
        this.lastElementToRender = lastElementToRender;
        this.visiblePhotos = {0: this.photos[0]};
        for (let i = firstElementToRender; i <= lastElementToRender; i++) {
          this.visiblePhotos[i] = this.photos[i];
        }
      }
    },
    getVirtualizedElementStyle(index) {
      // the very fist element is not actually virtualized
      if (index <= 0) {
        return '';
      }

      return virtualizationTools.getVirtualizedElementStyle(index, this.columnCount, this.elementSize, this.elementSize);
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
