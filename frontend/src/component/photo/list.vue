<template>
  <div>
    <div v-if="photos.length === 0" class="pa-2">
      <!-- TODO: change this icon -->
      <v-alert color="secondary-dark" :icon="isSharedView ? 'image_not_supported' : 'mdi-lightbulb-outline'" class="no-results ma-2 opacity-70" variant="outlined">
        <h3 v-if="filter.order === 'edited'" class="text-body-2 ma-0 pa-0">
          <translate>No recently edited pictures</translate>
        </h3>
        <h3 v-else class="text-body-2 ma-0 pa-0">
          <translate>No pictures found</translate>
        </h3>
        <p class="text-body-1 mt-2 mb-0 pa-0">
          <translate>Try again using other filters or keywords.</translate>
          <template v-if="!isSharedView">
            <translate>In case pictures you expect are missing, please rescan your library and wait until indexing has been completed.</translate>
            <template v-if="config.settings.features.review">
              <translate>Non-photographic and low-quality images require a review before they appear in search results.</translate>
            </template>
          </template>
        </p>
      </v-alert>
    </div>
    <div v-else class="search-results photo-results list-view">
      <div class="v-table__overflow">
        <v-table class="v-datatable v-table">
          <thead>
            <tr>
              <th class="p-col-select" />
              <th class="text-left">
                {{ $gettext("Title") }}
              </th>
              <th class="text-left hidden-xs">
                {{ $gettext("Taken") }}
              </th>
              <th class="text-left hidden-sm-and-down">
                {{ $gettext("Camera") }}
              </th>
              <th class="text-left hidden-xs">
                {{ showName ? $gettext("Name") : $gettext("Location") }}
              </th>
              <th class="text-center hidden-xs" />
            </tr>
          </thead>
          <tbody>
            <tr v-for="(photo, index) in photos" :key="photo.ID" ref="items" :data-index="index">
              <td :data-uid="photo.UID" class="result" :class="photo.classes()">
                <div v-if="index < firstVisibleElementIndex || index > lastVisibileElementIndex" class="image card"></div>
                <div
                  v-else
                  :key="photo.Hash"
                  :style="`background-image: url(${photo.thumbnailUrl('tile_50')})`"
                  class="card clickable image"
                  @touchstart="onMouseDown($event, index)"
                  @touchend.stop.prevent="onClick($event, index)"
                  @mousedown="onMouseDown($event, index)"
                  @contextmenu.stop="onContextMenu($event, index)"
                  @click.stop.prevent="onClick($event, index)"
                >
                  <button v-if="selectMode" class="input-select">
                    <i class="mdi mdi-circle-outline" />
                    <i class="mdi mdi-radiobox-blank" />
                  </button>
                  <button v-else-if="photo.Type === 'video' || photo.Type === 'live' || photo.Type === 'animated'" class="input-open" @click.stop.prevent="openPhoto(index, false, photo.Type === 'live')">
                    <i v-if="photo.Type === 'live'" class="action-live" :title="$gettext('Live')"><icon-live-photo /></i>
                    <i v-if="photo.Type === 'animated'" class="mdi mdi-file-gif-box" :title="$gettext('Animated')" />
                    <i v-if="photo.Type === 'vector'" class="action-vector mdi mdi-vector-polyline" :title="$gettext('Vector')"></i>
                    <i v-if="photo.Type === 'video'" class="mdi mdi-play" :title="$gettext('Video')" />
                  </button>
                </div>
              </td>

              <td class="p-photo-desc clickable" :data-uid="photo.UID" @click.exact="isSharedView ? openPhoto(index) : editPhoto(index)">
                {{ photo.Title }}
              </td>
              <td class="p-photo-desc hidden-xs" :title="photo.getDateString()">
                <button @click.stop.prevent="openDate(index)">
                  {{ photo.shortDateString() }}
                </button>
              </td>
              <td class="p-photo-desc hidden-sm-and-down">
                <button @click.stop.prevent="editPhoto(index)">{{ photo.CameraMake }} {{ photo.CameraModel }}</button>
              </td>
              <td class="p-photo-desc hidden-xs">
                <button v-if="filter.order === 'name'" :title="$gettext('Name')" @click.exact="downloadFile(index)">
                  {{ photo.FileName }}
                </button>
                <button v-else-if="photo.Country !== 'zz' && showLocation" @click.stop.prevent="openLocation(index)">
                  {{ photo.locationInfo() }}
                </button>
                <span v-else>
                  {{ photo.locationInfo() }}
                </span>
              </td>
              <template v-if="!isSharedView">
                <td class="text-center">
                  <template v-if="index < firstVisibleElementIndex || index > lastVisibileElementIndex">
                    <div v-if="hidePrivate" class="v-btn v-btn--icon v-btn--small" />
                    <div class="v-btn v-btn--icon v-btn--small" />
                  </template>

                  <template v-else>
                    <v-btn v-if="hidePrivate" class="input-private" icon size="small" variant="text" :ripple="false" :data-uid="photo.UID" @click.stop.prevent="photo.togglePrivate()">
                      <v-icon v-if="photo.Private" color="secondary-dark" class="select-on">mdi-lock</v-icon>
                      <v-icon v-else color="secondary" class="select-off">mdi-lock-open</v-icon>
                    </v-btn>
                    <v-btn class="input-favorite" icon size="small" variant="text" :ripple="false" :data-uid="photo.UID" @click.stop.prevent="photo.toggleLike()">
                      <v-icon v-if="photo.Favorite" color="secondary-dark" :data-uid="photo.UID" class="select-on">mdi-heart</v-icon>
                      <v-icon v-else color="secondary" :data-uid="photo.UID" class="select-off">mdi-heart-outline</v-icon>
                    </v-btn>
                  </template>
                </td>
              </template>
            </tr>
          </tbody>
        </v-table>
      </div>
    </div>
  </div>
</template>
<script>
import download from "common/download";
import Notify from "common/notify";
import { virtualizationTools } from "common/virtualization-tools";
import IconLivePhoto from "component/icon/live-photo.vue";

export default {
  name: "PPhotoList",
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
      m += " " + this.$gettext("Non-photographic and low-quality images require a review before they appear in search results.");
    }

    return {
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
      const [smallestIndex, largestIndex] = virtualizationTools.updateVisibleElementIndices(this.visibleElementIndices, entries, this.elementIndexFromIntersectionObserverEntry);

      // we observe only every 5th item, so we increase the rendered
      // range here by 4 items in every directio just to be safe
      this.firstVisibleElementIndex = smallestIndex - 4;
      this.lastVisibileElementIndex = largestIndex + 4;
    },
    downloadFile(index) {
      Notify.success(this.$gettext("Downloading…"));

      const photo = this.photos[index];
      download(`${this.$config.apiUri}/dl/${photo.Hash}?t=${this.$config.downloadToken}`, photo.FileName);
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
    onClick(ev, index) {
      const longClick = this.mouseDown.index === index && ev.timeStamp - this.mouseDown.timeStamp > 400;
      const scrolled = this.mouseDown.scrollY - window.scrollY !== 0;

      if (scrolled) {
        return;
      }

      if (longClick || this.selectMode) {
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
