<template>
  <div>
    <div v-if="photos.length === 0" class="pa-2">
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
            <template v-if="config.settings.features.review">
              <translate>Non-photographic and low-quality images require a review before they appear in search results.</translate>
            </template>
          </template>
        </p>
      </v-alert>
    </div>
    <v-data-table v-else
                  ref="dataTable"
                  v-model="selected"
                  :headers="listColumns"
                  :items="photos"
                  hide-actions
                  class="search-results photo-results list-view"
                  :class="{'select-results': selectMode}"
                  disable-initial-sort
                  item-key="ID"
                  :no-data-text="notFoundMessage"

    >
      <template #items="props">
        <td style="user-select: none;" :data-uid="props.item.UID" class="result" :class="props.item.classes()">
          <div
              v-if="props.index < firstVisibleElementIndex || props.index > lastVisibileElementIndex"
              class="v-image accent lighten-2"
              style="aspect-ratio: 1"
          />
          <div
                v-if="props.index >= firstVisibleElementIndex && props.index <= lastVisibileElementIndex"
                :key="props.item.Hash"
                :alt="props.item.Title"
                :style="`background-image: url(${props.item.thumbnailUrl('tile_50')})`"
                class="accent lighten-2 clickable image"
                @touchstart="onMouseDown($event, props.index)"
                @touchend.stop.prevent="onClick($event, props.index)"
                @mousedown="onMouseDown($event, props.index)"
                @contextmenu.stop="onContextMenu($event, props.index)"
                @click.stop.prevent="onClick($event, props.index)"
          >
            <button v-if="selectMode" class="input-select">
              <i class="select-on">check_circle</i>
              <i class="select-off">radio_button_off</i>
            </button>
            <button v-else-if="props.item.Type === 'video' || props.item.Type === 'live' || props.item.Type === 'animated'"
                  class="input-open"
                  @click.stop.prevent="openPhoto(props.index, true)">
              <i v-if="props.item.Type === 'live'" class="action-live" :title="$gettext('Live')">$vuetify.icons.live_photo</i>
              <i v-if="props.item.Type === 'animated'" class="action-animated" :title="$gettext('Animated')">gif</i>
              <i v-if="props.item.Type === 'video'" class="action-play" :title="$gettext('Video')">play_arrow</i>
            </button>
          </div>
        </td>

        <td class="p-photo-desc clickable" :data-uid="props.item.UID" style="user-select: none;"
            @click.exact="isSharedView ? openPhoto(props.index, false) : editPhoto(props.index)">
          {{ props.item.Title }}
        </td>
        <td class="p-photo-desc hidden-xs-only" :title="props.item.getDateString()">
          <button style="user-select: none;" @click.stop.prevent="editPhoto(props.index)">
            {{ props.item.shortDateString() }}
          </button>
        </td>
        <td class="p-photo-desc hidden-sm-and-down" style="user-select: none;">
          <button @click.stop.prevent="editPhoto(props.index)">
            {{ props.item.CameraMake }} {{ props.item.CameraModel }}
          </button>
        </td>
        <td class="p-photo-desc hidden-xs-only">
          <button v-if="filter.order === 'name'"
                  :title="$gettext('Name')" @click.exact="downloadFile(props.index)">
            {{ props.item.FileName }}
          </button>
          <button v-else-if="props.item.Country !== 'zz' && showLocation"
                  style="user-select: none;"
                  @click.stop.prevent="openLocation(props.index)">
            {{ props.item.locationInfo() }}
          </button>
          <span v-else>
                    {{ props.item.locationInfo() }}
                </span>
        </td>
        <template v-if="!isSharedView">
          <td class="text-xs-center">
            <template v-if="props.index < firstVisibleElementIndex || props.index > lastVisibileElementIndex">
              <div v-if="hidePrivate" class="v-btn v-btn--icon v-btn--small" />
              <div class="v-btn v-btn--icon v-btn--small" />
            </template>

            <template v-else>
              <v-btn v-if="hidePrivate" class="input-private" icon small flat :ripple="false"
                    :data-uid="props.item.UID" @click.stop.prevent="props.item.togglePrivate()">
                <v-icon v-if="props.item.Private" color="secondary-dark" class="select-on">lock</v-icon>
                <v-icon v-else color="secondary" class="select-off">lock_open</v-icon>
              </v-btn>
              <v-btn class="input-like" icon small flat :ripple="false"
                    :data-uid="props.item.UID" @click.stop.prevent="props.item.toggleLike()">
                <v-icon v-if="props.item.Favorite" color="pink lighten-3" :data-uid="props.item.UID" class="select-on">
                  favorite
                </v-icon>
                <v-icon v-else color="secondary" :data-uid="props.item.UID" class="select-off">favorite_border</v-icon>
              </v-btn>
            </template>
          </td>
        </template>
      </template>
    </v-data-table>
  </div>
</template>
<script>
import download from "common/download";
import Notify from "common/notify";
import {virtualizationTools} from 'common/virtualization-tools';

export default {
  name: 'PPhotoList',
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

    if (this.$config.feature("review") && !this.isSharedView) {
      m += " " + this.$gettext("Non-photographic and low-quality images require a review before they appear in search results.");
    }

    let showName = this.filter.order === 'name';

    const align = !this.$rtl ? 'left' : 'right';
    const listColumns = [
      {text: '', value: '', align: 'center', class: 'p-col-select', sortable: false},
      {text: this.$gettext('Title'), align, value: 'Title', sortable: false},
      {text: this.$gettext('Taken'), align, class: 'hidden-xs-only', value: 'TakenAt', sortable: false},
      {text: this.$gettext('Camera'), align, class: 'hidden-sm-and-down', value: 'CameraModel', sortable: false},
      {
        text: showName ? this.$gettext('Name') : this.$gettext('Location'),
        align,
        class: 'hidden-xs-only',
        value: showName ? 'FileName' : 'PlaceLabel',
        sortable: false
      },
    ];

    if (!this.isSharedView) {
      listColumns.push({text: '', value: '', align: 'center', sortable: false});
    }

    return {
      config: this.$config.values,
      notFoundMessage: m,
      'selected': [],
      'listColumns': listColumns,
      showName: showName,
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
  beforeCreate() {
    this.intersectionObserver = new IntersectionObserver((entries) => {
      this.visibilitiesChanged(entries);
    }, {
      rootMargin: "100% 0px",
    });
  },
  mounted() {
    this.observeItems();
  },
  updated() {
    this.observeItems();
  },
  beforeDestroy() {
    this.intersectionObserver.disconnect();
  },
  methods: {
    observeItems() {
      if (this.$refs.dataTable === undefined) {
        return;
      }
      const rows = this.$refs.dataTable.$el.getElementsByTagName('tbody')[0].children;
      for (const row of rows) {
        this.intersectionObserver.observe(row);
      }
    },
    elementIndexFromIntersectionObserverEntry(entry) {
      return entry.target.rowIndex - 2;
    },
    visibilitiesChanged(entries) {
      const [smallestIndex, largestIndex] = virtualizationTools.updateVisibleElementIndices(
        this.visibleElementIndices,
        entries,
        this.elementIndexFromIntersectionObserverEntry,
      );

      this.firstVisibleElementIndex = smallestIndex;
      this.lastVisibileElementIndex = largestIndex;
    },
    downloadFile(index) {
      Notify.success(this.$gettext("Downloading…"));

      const photo = this.photos[index];
      download(`${this.$config.apiUri}/dl/${photo.Hash}?t=${this.$config.downloadToken()}`, photo.FileName);
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
        let photo = this.photos[index];

        if ((photo.Type === 'video' || photo.Type === 'animated') && photo.isPlayable()) {
          this.openPhoto(index, true);
        } else {
          this.openPhoto(index, false);
        }
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
