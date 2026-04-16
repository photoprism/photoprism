<template>
  <div class="p-photos p-photo-view-cards">
    <div v-if="photos.length === 0" class="pa-3">
      <v-alert color="surface-variant" :icon="isSharedView ? 'mdi-image-off' : 'mdi-lightbulb-outline'" class="no-results" variant="outlined">
        <div v-if="filter.order === 'edited'" class="font-weight-bold">
          {{ $gettext(`No recently edited pictures`) }}
        </div>
        <div v-else class="font-weight-bold">
          {{ $gettext(`No pictures found`) }}
        </div>
        <div class="mt-2">
          {{ $gettext(`Try again using other filters or keywords.`) }}
          <template v-if="!isSharedView">
            {{ $gettext(`In case pictures you expect are missing, please rescan your library and wait until indexing has been completed.`) }}
            <template v-if="$config.feature('review')">
              {{ $gettext(`Non-photographic and low-quality images require a review before they appear in search results.`) }}
            </template>
          </template>
        </div>
      </v-alert>
    </div>
    <div v-else class="v-row search-results photo-results cards-view" :class="{ 'select-results': selectMode }">
      <div v-for="(m, index) in photos" :key="m.ID" ref="items" :data-index="index" class="v-col-12 v-col-sm-6 v-col-md-4 v-col-lg-3 v-col-xl-2">
        <div v-if="index < firstVisibleElementIndex || index > lastVisibleElementIndex" :data-id="m.ID" :data-uid="m.UID" class="media result placeholder">
          <div class="preview" />
          <div v-if="!isSharedView && m.Quality < 3 && context === contexts.Review" class="review" />
          <div class="meta">
            <div class="meta-details meta-fields">
              <div v-for="item in cardMetadataItems(m, false)" :key="`${m.ID}-${item.key}`" :title="item.label" :class="item.className">
                <i v-if="item.showIcon" class="mdi" :class="item.icon" />
                {{ item.text }}
              </div>
            </div>
          </div>
        </div>
        <div v-else :data-id="m.ID" :data-uid="m.UID" class="media result" :class="m.classes()" @contextmenu.stop="onContextMenu($event, index)">
          <div
            :title="m.Title"
            :style="`background-image: url(${m.thumbnailUrl(tileSize)})`"
            class="preview"
            @touchstart.passive="input.touchStart($event, index)"
            @touchend.stop="onClick($event, index)"
            @mousedown.stop="input.mouseDown($event, index)"
            @click.stop.prevent="onClick($event, index)"
            @mouseover="playLive(m)"
            @mouseleave="pauseLive(m)"
          >
            <div class="preview__overlay"></div>
            <div v-if="m.Type === 'live' || m.Type === 'animated'" class="live-player">
              <video :id="'live-player-' + m.ID" width="500" height="500" preload="none" loop muted playsinline>
                <source :type="m.videoContentType()" :src="m.videoUrl()" />
              </video>
            </div>

            <button
              v-if="(m.Type !== 'image' && m.Type !== 'video') || selectMode || m.isStack()"
              class="input-open"
              @touchstart.stop="input.touchStart($event, index)"
              @touchend.stop="onOpen($event, index, !isSharedView)"
              @touchmove.stop
              @click.stop.prevent="onOpen($event, index, !isSharedView)"
            >
              <i v-if="m.Type === 'raw'" class="action-raw mdi mdi-raw" :title="$gettext('RAW')" />
              <i v-else-if="m.Type === 'live'" class="action-live" :title="$gettext('Live')"><icon-live-photo /></i>
              <i v-else-if="m.Type === 'animated'" class="mdi mdi-file-gif-box" :title="$gettext('Animated')" />
              <i v-else-if="m.Type === 'vector'" class="action-vector mdi mdi-vector-polyline" :title="$gettext('Vector')"></i>
              <i v-else-if="m.Type === 'document'" class="action-document mdi mdi-file-pdf-box" :title="$gettext('Document')" />
              <i v-else-if="m.Type === 'image' && !selectMode" class="mdi mdi-camera-burst" :title="$gettext('Stack')" />
              <i v-else class="mdi mdi-magnify-plus-outline" :title="$gettext('View')" />
            </button>

            <div class="preview-details">
              <div v-if="!isSharedView && featPrivate && m.Private" class="info-icon"><i class="mdi mdi-lock" /></div>
              <div v-if="m.Type === 'video'" :title="$gettext('Video')" class="info-text">
                {{ m.getDurationInfo() }}
              </div>
            </div>

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
              @touchstart.stop="input.touchStart($event, index)"
              @touchend.stop="onSelect($event, index)"
              @touchmove.stop
              @click.stop.prevent="onSelect($event, index)"
            >
              <i class="mdi mdi-check-circle select-on" />
              <i class="mdi mdi-circle-outline select-off" />
            </button>

            <button
              v-if="!isSharedView"
              class="input-favorite"
              @touchstart.stop="input.touchStart($event, index)"
              @touchend.stop="toggleLike($event, index)"
              @touchmove.stop
              @click.stop.prevent="toggleLike($event, index)"
            >
              <i v-if="m.Favorite" class="mdi mdi-star text-favorite favorite-on" />
              <i v-else class="mdi mdi-star-outline favorite-off" />
            </button>
          </div>

          <div v-if="!isSharedView && m.Quality < 3 && context === contexts.Review" class="review">
            <button
              type="button"
              class="v-btn v-btn--flat bg-button v-btn--variant-tonal action-archive text-center"
              :title="$pgettext('Verb', 'Archive')"
              @click.stop="m.archive()"
            >
              <span class="v-btn__overlay"></span>
              <span class="v-btn__underlay"></span>
              <span class="v-btn__content" data-no-activator=""
                ><i class="mdi-close mdi v-icon notranslate v-theme--default v-icon--size-default" aria-hidden="true"></i
              ></span>
            </button>
            <button
              type="button"
              class="v-btn v-btn--flat bg-button v-btn--variant-tonal action-approve text-center"
              :title="$gettext('Approve')"
              @click.stop="m.approve()"
            >
              <span class="v-btn__overlay"></span>
              <span class="v-btn__underlay"></span>
              <span class="v-btn__content" data-no-activator=""><i class="mdi-check mdi v-icon notranslate v-icon--size-default" aria-hidden="true"></i></span>
            </button>
          </div>
          <div class="meta">
            <div class="meta-details meta-fields">
              <component
                :is="item.clickable ? 'button' : 'div'"
                v-for="item in cardMetadataItems(m)"
                :key="`${m.ID}-${item.key}`"
                :title="item.label"
                :class="item.className"
                @click.exact="onMetadataAction(item.action, index)"
              >
                <i v-if="item.showIcon" class="mdi" :class="item.icon" />
                {{ item.text }}
              </component>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>
<script>
import download from "common/download";
import { hasMetadataText, metadataIcon, metadataLabel, metadataLayout, metadataText, MetadataView } from "common/metadata";
import $notify from "common/notify";
import { Input, InputInvalid, ClickShort, ClickLong } from "common/input";
import { virtualizationTools } from "common/virtualization-tools";
import * as contexts from "options/contexts";
import IconLivePhoto from "component/icon/live-photo.vue";

export default {
  name: "PPhotoViewCards",
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
    const input = new Input();
    const debug = this.$config.get("debug");
    const trace = this.$config.get("trace");
    const settings = this.$config.getSettings();
    const featPlaces = settings.features.places;
    const featPrivate = settings.features.private;
    const featDownload = settings.features.download;
    const tileSize = settings.display?.retinaThumbnails ? "tile_1080" : "tile_500";

    return {
      featPlaces,
      featPrivate,
      featDownload,
      tileSize,
      input,
      debug,
      trace,
      contexts,
      firstVisibleElementIndex: 0,
      lastVisibleElementIndex: 0,
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
  computed: {
    cardLayout() {
      return metadataLayout(this.$config.getSettings(), MetadataView.Cards);
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
      const [smallestIndex, largestIndex] = virtualizationTools.updateVisibleElementIndices(
        this.visibleElementIndices,
        entries,
        this.elementIndexFromIntersectionObserverEntry
      );

      // we observe only every 5th item, so we increase the rendered
      // range here by 4 items in every direction just to be safe
      this.firstVisibleElementIndex = smallestIndex - 4;
      this.lastVisibleElementIndex = largestIndex + 4;
    },
    livePlayer(photo) {
      return document.querySelector("#live-player-" + photo.ID);
    },
    playLive(photo) {
      const player = this.livePlayer(photo);
      if (player) {
        try {
          // Calling pause() before a play promise has been resolved may result in an error,
          // see https://developer.chrome.com/blog/play-request-was-interrupted.
          const playPromise = player.play();
          if (playPromise !== undefined) {
            playPromise.catch((err) => {
              if (this.trace && err && err.message) {
                // Ignore this error, or uncomment the following line to log it.
                // console.debug(err.message);
              }
            });
          }
        } catch (e) {
          // Ignore.
        }
      }
    },
    pauseLive(photo) {
      const player = this.livePlayer(photo);
      if (player) {
        try {
          // Calling pause() before a play promise has been resolved may result in an error,
          // see https://developer.chrome.com/blog/play-request-was-interrupted.
          if (!player.paused) {
            player.pause();
          }
        } catch (e) {
          if (this.trace) {
            console.log(e);
          }
        }
      }
    },
    downloadFile(index) {
      $notify.success(this.$gettext("Downloading…"));

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
    cardMetadataItems(photo, interactive = true) {
      return this.cardLayout
        .map((fieldId, index) => {
          if (fieldId === "location" && (!this.featPlaces || photo?.Country === "zz")) {
            return null;
          } else if (!hasMetadataText(photo, fieldId)) {
            return null;
          }

          const action = this.metadataAction(fieldId);

          return {
            key: `${fieldId}-${index}`,
            label: metadataLabel(fieldId),
            text: metadataText(photo, fieldId),
            icon: metadataIcon(fieldId, photo),
            showIcon: !["title", "caption"].includes(fieldId),
            clickable: interactive && action !== "",
            className: this.metadataClass(fieldId, interactive && action !== ""),
            action,
          };
        })
        .filter(Boolean);
    },
    metadataAction(fieldId) {
      switch (fieldId) {
        case "date":
          return "date";
        case "location":
          return "location";
        case "filename":
          return this.isSharedView ? "open" : "files";
        case "camera":
        case "lens":
        case "exposure":
        case "fileInfo":
          return this.isSharedView ? "open" : "details";
        case "title":
        case "caption":
        case "keywords":
        default:
          return this.isSharedView ? "open" : "edit";
      }
    },
    metadataClass(fieldId, clickable) {
      const classes = ["meta-field", `meta-field--${fieldId}`, `meta-${fieldId}`];

      if (clickable) {
        classes.push("clickable");
      }

      return classes.join(" ");
    },
    onMetadataAction(action, index) {
      switch (action) {
        case "date":
          this.openDate(index);
          break;
        case "location":
          this.openLocation(index);
          break;
        case "files":
          this.editPhoto(index, "files");
          break;
        case "details":
          this.editPhoto(index, "details");
          break;
        case "edit":
          this.editPhoto(index);
          break;
        case "open":
          this.openPhoto(index);
          break;
        default:
          break;
      }
    },
  },
};
</script>
