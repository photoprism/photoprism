<template>
  <div class="p-photos p-photo-view-timeline">
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
    <div v-else class="search-results photo-results mosaic-view timeline-view" :class="{ 'select-results': selectMode }">
      <section v-for="section in sections" :key="section.key" class="timeline-section">
        <div class="timeline-section__header">
          <h2>{{ section.title }}</h2>
          <span>{{ section.countLabel }}</span>
        </div>
        <div v-for="day in section.days" :key="day.key" class="timeline-day">
          <div class="timeline-day__label">
            <span>{{ day.title }}</span>
          </div>
          <div class="v-row timeline-day__photos">
            <div
              v-for="{ photo: m, index } in day.entries"
              :key="m.ID || m.UID"
              ref="items"
              class="v-col-4 v-col-sm-3 v-col-md-2 v-col-lg-1"
              :data-index="index"
            >
              <div class="result-container">
                <div
                  v-if="index < firstVisibleElementIndex || index > lastVisibleElementIndex"
                  :data-id="m.ID"
                  :data-uid="m.UID"
                  class="media result preview placeholder"
                />
                <div
                  v-else
                  :data-id="m.ID"
                  :data-uid="m.UID"
                  :title="showTitles && m.Title ? m.Title : m.getOriginalName()"
                  :style="`background-image: url(${m.thumbnailUrl('tile_224')})`"
                  :class="m.classes()"
                  class="media result preview"
                  @contextmenu.stop="onContextMenu($event, index)"
                  @touchstart.passive="input.touchStart($event, index)"
                  @touchend.stop="onClick($event, index)"
                  @mousedown.stop="input.mouseDown($event, index)"
                  @click.stop.prevent="onClick($event, index)"
                  @mouseover="playLive(m)"
                  @mouseleave="pauseLive(m)"
                >
                  <div class="preview__overlay"></div>
                  <div v-if="m.Type === 'live' || m.Type === 'animated'" class="live-player">
                    <video :id="'live-player-' + m.ID" width="224" height="224" preload="none" loop muted playsinline>
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
                    <div v-if="!isSharedView && hidePrivate && m.Private" class="info-icon"><i class="mdi mdi-lock" /></div>
                    <div v-if="m.Type === 'video'" :title="$gettext('Video')" class="info-text">
                      {{ m.getDurationInfo() }}
                    </div>
                  </div>

                  <button
                    class="input-select"
                    @mousedown.stop="input.mouseDown($event, index)"
                    @touchstart.stop="input.touchStart($event, index)"
                    @touchend.stop="onSelect($event, index)"
                    @touchmove.stop.prevent
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
              </div>
            </div>
          </div>
        </div>
      </section>
    </div>
  </div>
</template>

<script>
import * as options from "options/options";
import { Input, InputInvalid, ClickShort, ClickLong } from "common/input";
import { virtualizationTools } from "common/virtualization-tools";
import IconLivePhoto from "component/icon/live-photo.vue";
import { buildTimelineSections } from "component/photo/view/timeline";

export default {
  name: "PPhotoViewTimeline",
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
    const trace = this.$config.get("trace");
    const settings = this.$config.getSettings();
    const showTitles = settings.search?.showTitles;

    return {
      input,
      trace,
      showTitles,
      hidePrivate: settings.features?.private,
      firstVisibleElementIndex: 0,
      lastVisibleElementIndex: 0,
      visibleElementIndices: new Set(),
    };
  },
  computed: {
    sections() {
      return buildTimelineSections(this.photos, options.Months(), this.$gettext);
    },
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

      for (let i = 0; i < this.$refs.items.length; i += 5) {
        this.intersectionObserver.observe(this.$refs.items[i]);
      }
    },
    elementIndexFromIntersectionObserverEntry(entry) {
      return parseInt(entry.target.getAttribute("data-index"), 10);
    },
    visibilitiesChanged(entries) {
      const [smallestIndex, largestIndex] = virtualizationTools.updateVisibleElementIndices(
        this.visibleElementIndices,
        entries,
        this.elementIndexFromIntersectionObserverEntry
      );

      if (smallestIndex === undefined || largestIndex === undefined) {
        return;
      }

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
          const playPromise = player.play();
          if (playPromise !== undefined) {
            playPromise.catch((err) => {
              if (this.trace && err && err.message) {
                // Ignore this error, or uncomment the following line to log it.
                // console.debug(err.message);
              }
            });
          }
        } catch {
          // Ignore.
        }
      }
    },
    pauseLive(photo) {
      const player = this.livePlayer(photo);
      if (player) {
        try {
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
      this.$forceUpdate();
    },
  },
};
</script>
