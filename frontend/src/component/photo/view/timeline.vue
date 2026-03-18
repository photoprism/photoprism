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
    <div v-else class="search-results photo-results timeline-view" :class="{ 'select-results': selectMode }">
      <template v-for="group in groups" :key="group.key">
        <div class="timeline-group-header">
          <span class="timeline-group-title">{{ group.title }}</span>
          <span class="timeline-group-count">{{ group.photos.length }}</span>
        </div>
        <div class="v-row timeline-group-photos">
          <div
            v-for="(m, groupIndex) in group.photos"
            :key="m.ID"
            ref="items"
            :data-index="m.globalIndex"
            class="v-col-4 v-col-sm-3 v-col-md-2 v-col-lg-1"
          >
            <div class="result-container">
              <div
                v-if="m.globalIndex < firstVisibleElementIndex || m.globalIndex > lastVisibleElementIndex"
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
                @contextmenu.stop="onContextMenu($event, m.globalIndex)"
                @touchstart.passive="input.touchStart($event, m.globalIndex)"
                @touchend.stop="onClick($event, m.globalIndex)"
                @mousedown.stop="input.mouseDown($event, m.globalIndex)"
                @click.stop.prevent="onClick($event, m.globalIndex)"
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
                  @touchstart.stop="input.touchStart($event, m.globalIndex)"
                  @touchend.stop="onOpen($event, m.globalIndex, !isSharedView)"
                  @touchmove.stop
                  @click.stop.prevent="onOpen($event, m.globalIndex, !isSharedView)"
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
                  @mousedown.stop="input.mouseDown($event, m.globalIndex)"
                  @touchstart.stop="input.touchStart($event, m.globalIndex)"
                  @touchend.stop="onSelect($event, m.globalIndex)"
                  @touchmove.stop.prevent
                  @click.stop.prevent="onSelect($event, m.globalIndex)"
                >
                  <i class="mdi mdi-check-circle select-on" />
                  <i class="mdi mdi-circle-outline select-off" />
                </button>

                <button
                  v-if="!isSharedView"
                  class="input-favorite"
                  @touchstart.stop="input.touchStart($event, m.globalIndex)"
                  @touchend.stop="toggleLike($event, m.globalIndex)"
                  @touchmove.stop
                  @click.stop.prevent="toggleLike($event, m.globalIndex)"
                >
                  <i v-if="m.Favorite" class="mdi mdi-star text-favorite favorite-on" />
                  <i v-else class="mdi mdi-star-outline favorite-off" />
                </button>
              </div>
            </div>
          </div>
        </div>
      </template>
    </div>
  </div>
</template>
<script>
import { Input, InputInvalid, ClickShort, ClickLong } from "common/input";
import { virtualizationTools } from "common/virtualization-tools";
import IconLivePhoto from "component/icon/live-photo.vue";

const monthNames = ["January", "February", "March", "April", "May", "June", "July", "August", "September", "October", "November", "December"];

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
    const debug = this.$config.get("debug");
    const trace = this.$config.get("trace");
    const settings = this.$config.getSettings();
    const showTitles = settings.search.showTitles;
    const showCaptions = settings.search.showCaptions;

    return {
      input,
      debug,
      trace,
      showTitles,
      showCaptions,
      hidePrivate: settings.features.private,
      firstVisibleElementIndex: 0,
      lastVisibleElementIndex: 0,
      visibleElementIndices: new Set(),
    };
  },
  computed: {
    groups() {
      const buckets = new Map();

      for (let i = 0; i < this.photos.length; i++) {
        const m = this.photos[i];
        const year = m.Year > 0 ? m.Year : 0;
        const month = m.Month > 0 && m.Month <= 12 ? m.Month : 0;
        const key = year > 0 && month > 0 ? `${year}-${String(month).padStart(2, "0")}` : year > 0 ? `${year}` : "unknown";

        if (!buckets.has(key)) {
          let title;
          if (year > 0 && month > 0) {
            title = `${monthNames[month - 1]} ${year}`;
          } else if (year > 0) {
            title = String(year);
          } else {
            title = this.$gettext("Unknown Date");
          }
          buckets.set(key, { key, title, photos: [] });
        }

        buckets.get(key).photos.push(Object.assign(m, { globalIndex: i }));
      }

      return Array.from(buckets.values());
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
      return parseInt(entry.target.getAttribute("data-index"));
    },
    visibilitiesChanged(entries) {
      const [smallestIndex, largestIndex] = virtualizationTools.updateVisibleElementIndices(
        this.visibleElementIndices,
        entries,
        this.elementIndexFromIntersectionObserverEntry
      );

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
