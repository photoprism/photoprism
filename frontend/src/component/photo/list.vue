<template>
  <div>
    <div v-if="photos.length === 0" class="pa-2">
      <v-alert
        :value="true"
        color="secondary-dark"
        :icon="isSharedView ? 'image_not_supported' : 'lightbulb_outline'"
        class="no-results ma-2 opacity-70"
        outline>
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
            <template v-if="config.settings.features.review">
              <translate>Non-photographic and low-quality images require a review before they appear in search results.</translate>
            </template>
          </template>
        </p>
      </v-alert>
    </div>
    <div v-else class="search-results photo-results list-view">
      <div class="v-table__overflow">
        <table class="v-datatable v-table theme--light">
          <thead>
            <tr>
              <th class="p-col-select" />
              <th class="text-xs-left">
                {{$gettext('Title')}}
              </th>
              <th class="text-xs-left hidden-xs-only">
                {{$gettext('Taken')}}
              </th>
              <th class="text-xs-left hidden-sm-and-down">
                {{$gettext('Camera')}}
              </th>
              <th class="text-xs-left hidden-xs-only">
                {{showName ? $gettext('Name') : $gettext('Location')}}
              </th>
              <th class="text-xs-center hidden-xs-only" />
            </tr>
          </thead>
          <tbody>
            <tr v-for="(photo, index) in photos" :key="photo.ID" ref="items" :data-index="index">
              <td :data-uid="photo.UID" class="result" :class="photo.classes()">
                <div
                  v-if="index < firstVisibleElementIndex || index > lastVisibileElementIndex"
                  :key="photo.Hash"
                  class="image card darken-1">
                </div>
                <div
                  v-else
                  :key="photo.Hash"
                  :style="`background-image: url(${photo.thumbnailUrl('tile_50')})`"
                  class="card darken-1 clickable image"
                  @touchstart="onMouseDown($event, index)"
                  @touchend.stop.prevent="onClick($event, index)"
                  @mousedown="onMouseDown($event, index)"
                  @contextmenu.stop="onContextMenu($event, index)"
                  @click.stop.prevent="onClick($event, index)">
                  <button v-if="selectMode" class="input-select">
                    <i class="select-on">check_circle</i>
                    <i class="select-off">radio_button_off</i>
                  </button>
                  <button v-else-if="photo.Type === 'video' || photo.Type === 'live' || photo.Type === 'animated'"
                    class="input-open"
                    @click.stop.prevent="openPhoto(index, false, photo.Type === 'live')">
                    <i v-if="photo.Type === 'live'" class="action-live" :title="$gettext('Live')"><icon-live-photo/></i>
                    <i v-if="photo.Type === 'animated'" class="action-animated" :title="$gettext('Animated')">gif</i>
                    <i v-if="photo.Type === 'vector'" class="action-vector" :title="$gettext('Vector')">font_download</i>
                    <i v-if="photo.Type === 'video'" class="action-play" :title="$gettext('Video')">play_arrow</i>
                  </button>
                </div>
              </td>

              <td class="p-photo-desc clickable" :data-uid="photo.UID"
                @click.exact="isSharedView ? openPhoto(index) : editPhoto(index)">
                {{ photo.Title }}
              </td>
              <td class="p-photo-desc hidden-xs-only" :title="photo.getDateString()">
                <button @click.stop.prevent="editPhoto(index)">
                  {{ photo.shortDateString() }}
                </button>
              </td>
              <td class="p-photo-desc hidden-sm-and-down">
                <button @click.stop.prevent="editPhoto(index)">
                  {{ photo.CameraMake }} {{ photo.CameraModel }}
                </button>
              </td>
              <td class="p-photo-desc hidden-xs-only">
                <button v-if="filter.order === 'name'"
                        :title="$gettext('Name')" @click.exact="downloadFile(index)">
                  {{ photo.FileName }}
                </button>
                <button v-else-if="photo.Country !== 'zz' && showLocation"
                        @click.stop.prevent="openLocation(index)">
                  {{ photo.locationInfo() }}
                </button>
                <span v-else>
                  {{ photo.locationInfo() }}
                </span>
              </td>
              <template v-if="!isSharedView">
                <td class="text-xs-center">
                  <template v-if="index < firstVisibleElementIndex || index > lastVisibileElementIndex">
                    <div v-if="hidePrivate" class="v-btn v-btn--icon v-btn--small" />
                    <div class="v-btn v-btn--icon v-btn--small" />
                  </template>

                  <template v-else>
                    <v-btn v-if="hidePrivate" class="input-private" icon small flat :ripple="false"
                          :data-uid="photo.UID" @click.stop.prevent="photo.togglePrivate()">
                      <v-icon v-if="photo.Private" color="secondary-dark" class="select-on">lock</v-icon>
                      <v-icon v-else color="secondary" class="select-off">lock_open</v-icon>
                    </v-btn>
                    <v-btn class="input-like" icon small flat :ripple="false"
                          :data-uid="photo.UID" @click.stop.prevent="photo.toggleLike()">
                      <v-icon v-if="photo.Favorite" color="pink lighten-3" :data-uid="photo.UID" class="select-on">
                        favorite
                      </v-icon>
                      <v-icon v-else color="secondary" :data-uid="photo.UID" class="select-off">favorite_border</v-icon>
                    </v-btn>
                  </template>
                </td>
              </template>
            </tr>
          </tbody>
        </table>
      </div>
    </div>
  </div>
</template>
<script>
import download from "common/download";
import Notify from "common/notify";
import {virtualizationTools} from 'common/virtualization-tools';
import IconLivePhoto from "component/icon/live-photo.vue";

export default {
  name: 'PPhotoList',
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
      default:() => {},
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
      showName: this.filter.order === 'name',
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
    }
  },
  beforeCreate() {
    this.intersectionObserver = new IntersectionObserver((entries) => {
      this.visibilitiesChanged(entries);
    }, {
      rootMargin: "100% 0px",
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
      return parseInt(entry.target.getAttribute('data-index'));
    },
    visibilitiesChanged(entries) {
      const [smallestIndex, largestIndex] = virtualizationTools.updateVisibleElementIndices(
        this.visibleElementIndices,
        entries,
        this.elementIndexFromIntersectionObserverEntry,
      );

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
      const longClick = (this.mouseDown.index === index && (ev.timeStamp - this.mouseDown.timeStamp) > 400);
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
  }
};
</script>
