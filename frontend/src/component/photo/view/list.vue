<template>
  <div>
    <div v-if="photos.length === 0" class="pa-3">
      <v-alert
        color="surface-variant"
        :icon="isSharedView ? 'mdi-image-off' : 'mdi-lightbulb-outline'"
        class="no-results"
        variant="outlined"
      >
        <div v-if="filter.order === 'edited'" class="font-weight-bold">
          {{ $gettext(`No recently edited pictures`) }}
        </div>
        <div v-else class="font-weight-bold">
          {{ $gettext(`No pictures found`) }}
        </div>
        <div class="mt-2">
          {{ $gettext(`Try again using other filters or keywords.`) }}
          <template v-if="!isSharedView">
            {{
              $gettext(
                `In case pictures you expect are missing, please rescan your library and wait until indexing has been completed.`
              )
            }}
            <template v-if="$config.feature('review')">
              {{
                $gettext(
                  `Non-photographic and low-quality images require a review before they appear in search results.`
                )
              }}
            </template>
          </template>
        </div>
      </v-alert>
    </div>
    <div v-else class="search-results photo-results list-view">
      <div
        :class="$vuetify.display.smAndDown ? 'v-table--density-compact' : 'v-table--density-default'"
        class="v-table v-table--density-default v-table v-table--hover v-datatable"
      >
        <div class="v-table__wrapper">
          <table>
            <thead class="hidden-xs">
              <tr>
                <th class="col-select"></th>
                <th class="col-preview"></th>
                <th class="col-auto text-start">
                  {{ showTitles ? $gettext("Title") : $gettext("File Name") }}
                </th>
                <th class="col-taken text-start hidden-xs">
                  {{ $gettext("Taken") }}
                </th>
                <th class="col-md text-start hidden-sm-and-down">
                  {{ $gettext("Camera") }}
                </th>
                <th class="col-lg text-start hidden-md-and-down">
                  {{ showName ? $gettext("Name") : $gettext("Location") }}
                </th>
                <th v-if="!isSharedView" class="col-xs text-center"></th>
              </tr>
            </thead>
            <tbody>
              <tr v-for="(m, index) in photos" :key="m.ID" ref="items" :data-index="index">
                <td :data-id="m.ID" :data-uid="m.UID" class="col-select" :class="{ 'is-selected': isSelected(m) }">
                  <button
                    class="input-select"
                    @touchstart.passive="onMouseDown($event, index)"
                    @touchend.stop="onClick($event, index, true)"
                    @mousedown="onMouseDown($event, index)"
                    @contextmenu.stop="onContextMenu($event, index)"
                    @click.stop.prevent="onClick($event, index, true)"
                  >
                    <i class="mdi mdi-checkbox-marked select-on" />
                    <i class="mdi mdi-checkbox-blank-outline select-off" />
                  </button>
                </td>
                <td :data-id="m.ID" :data-uid="m.UID" class="media result col-preview" :class="m.classes()">
                  <div v-if="index < firstVisibleElementIndex || index > lastVisibleElementIndex" class="preview"></div>
                  <div
                    v-else
                    :style="`background-image: url(${m.thumbnailUrl('tile_224')})`"
                    class="preview"
                    @touchstart.passive="onMouseDown($event, index)"
                    @touchend.stop="onClick($event, index, false)"
                    @mousedown="onMouseDown($event, index)"
                    @contextmenu.stop="onContextMenu($event, index)"
                    @click.stop.prevent="onClick($event, index, false)"
                  >
                    <div class="preview__overlay"></div>
                    <button
                      v-if="m.Type === 'video' || m.Type === 'live' || m.Type === 'animated'"
                      class="input-open"
                      @click.stop.prevent="openPhoto(index, false)"
                    >
                      <i v-if="m.Type === 'live'" class="action-live" :title="$gettext('Live')"><icon-live-photo /></i>
                      <i v-else-if="m.Type === 'animated'" class="mdi mdi-file-gif-box" :title="$gettext('Animated')" />
                      <i v-else-if="m.Type === 'video'" class="mdi mdi-play" :title="$gettext('Video')" />
                    </button>
                  </div>
                </td>
                <td
                  class="meta-data meta-title col-auto text-start clickable"
                  :title="m.Title"
                  @click.exact="isSharedView ? openPhoto(index) : editPhoto(index)"
                >
                  {{ showTitles && m.Title ? m.Title : m.getOriginalName() }}
                </td>
                <td class="meta-data meta-date hidden-xs text-start col-taken" :title="m.getDateString()">
                  <span class="text-truncate clickable" @click.stop.prevent="openDate(index)">
                    {{ m.shortDateString() }}
                  </span>
                </td>
                <td class="meta-data hidden-sm-and-down text-start col-md">
                  <span class="text-truncate clickable" @click.stop.prevent="editPhoto(index)">
                    {{ m.CameraMake }} {{ m.CameraModel }}
                  </span>
                </td>
                <td class="meta-data hidden-md-and-down text-start col-lg">
                  <span
                    v-if="m.Country !== 'zz' && showLocation"
                    class="text-truncate clickable"
                    @click.stop.prevent="openLocation(index)"
                  >
                    {{ m.locationInfo() }}
                  </span>
                  <span v-else class="text-truncate">
                    {{ m.locationInfo() }}
                  </span>
                </td>
                <td v-if="!isSharedView" class="text-center col-xs">
                  <div class="table-actions">
                    <template v-if="index < firstVisibleElementIndex || index > lastVisibleElementIndex">
                      <div class="v-btn v-btn--icon v-btn--small" />
                    </template>

                    <template v-else>
                      <v-btn
                        icon
                        density="comfortable"
                        variant="text"
                        :ripple="false"
                        class="input-favorite"
                        @click.stop.prevent="m.toggleLike()"
                      >
                        <v-icon v-if="m.Favorite" icon="mdi-star" color="favorite" class="favorite-on"></v-icon>
                        <v-icon v-else icon="mdi-star-outline" color="surface" class="favorite-off"></v-icon>
                      </v-btn>
                    </template>
                  </div>
                </td>
              </tr>
            </tbody>
          </table>
        </div>
      </div>
    </div>
  </div>
</template>
<script>
import download from "common/download";
import $notify from "common/notify";
import { virtualizationTools } from "common/virtualization-tools";
import IconLivePhoto from "component/icon/live-photo.vue";
import { PhotoClipboard } from "common/clipboard";

export default {
  name: "PPhotoViewList",
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
    let m = this.$gettext("Couldn't find anything.");

    m += " " + this.$gettext("Try again using other filters or keywords.");

    if (!this.isSharedView && this.$config.feature("review")) {
      m +=
        " " +
        this.$gettext("Non-photographic and low-quality images require a review before they appear in search results.");
    }
    const settings = this.$config.getSettings();
    const showTitles = settings.search.showTitles;
    const showCaptions = settings.search.showCaptions;

    return {
      showTitles,
      showCaptions,
      config: this.$config.values,
      notFoundMessage: m,
      showName: this.filter.order === "name",
      showLocation: this.$config.values.settings.features.places,
      hidePrivate: this.$config.values.settings.features.private,
      mouseDown: {
        index: -1,
        scrollY: window.scrollY,
        timeStamp: -1,
      },
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
  beforeCreate() {
    this.intersectionObserver = new IntersectionObserver(
      (entries) => {
        this.visibilitiesChanged(entries);
      },
      {
        rootMargin: "100% 0px",
      }
    );
  },
  beforeUnmount() {
    this.intersectionObserver.disconnect();
  },
  methods: {
    isSelected(m) {
      return PhotoClipboard.has(m);
    },
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
      return parseInt(entry.target.getAttribute("data-index"));
    },
    visibilitiesChanged(entries) {
      const [smallestIndex, largestIndex] = virtualizationTools.updateVisibleElementIndices(
        this.visibleElementIndices,
        entries,
        this.elementIndexFromIntersectionObserverEntry
      );

      // we observe only every 5th item, so we increase the rendered
      // range here by 4 items in every directio just to be safe
      this.firstVisibleElementIndex = smallestIndex - 4;
      this.lastVisibleElementIndex = largestIndex + 4;
    },
    downloadFile(index) {
      $notify.success(this.$gettext("Downloading…"));

      const photo = this.photos[index];
      download(`${this.$config.apiUri}/dl/${photo.Hash}?t=${this.$config.downloadToken}`, photo.FileName);
    },
    onMouseDown(ev, index) {
      this.mouseDown.index = index;
      this.mouseDown.scrollY = window.scrollY;
      this.mouseDown.timeStamp = ev.timeStamp;
    },
    onClick(ev, index, select) {
      const longClick = this.mouseDown.index === index && ev.timeStamp - this.mouseDown.timeStamp > 400;
      const scrolled = this.mouseDown.scrollY - window.scrollY !== 0;

      if (!select && scrolled) {
        return;
      }

      ev.preventDefault();
      ev.stopPropagation();

      if (select !== false && (select || longClick || this.selectMode)) {
        if (longClick || ev.shiftKey) {
          this.selectRange(index);
        } else {
          this.toggle(this.photos[index]);
        }
      } else if (this.photos[index]) {
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
    toggle(photo) {
      this.$clipboard.toggle(photo);
    },
    selectRange(index) {
      this.$clipboard.addRange(index, this.photos);
    },
  },
};
</script>
