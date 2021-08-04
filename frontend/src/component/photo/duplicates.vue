<template>
  <div>
    <div v-if="photos.length === 0" class="pa-2">
      <v-card class="no-results secondary-light lighten-1 ma-1" flat>
        <v-card-title primary-title>
          <div>
            <h3 v-if="filter.order === 'edited'" class="title ma-0 pa-0">
              <translate>Couldn't find recently edited</translate>
            </h3>
            <h3 v-else class="title ma-0 pa-0">
              <translate>Couldn't find anything</translate>
            </h3>
            <p class="mt-4 mb-0 pa-0">
              <translate>Try again using other filters or keywords.</translate>
              <translate>If a file you expect is missing, please re-index your library and wait until indexing has been completed.</translate>
              <template v-if="config.settings.features.review" class="mt-2 mb-0 pa-0">
                <translate>Non-photographic and low-quality images require a review before they appear in search results.</translate>
              </template>
            </p>
          </div>
        </v-card-title>
      </v-card>
    </div>
    <v-data-table v-else
                  v-model="selected"
                  :headers="listColumns"
                  :items="photos"
                  hide-actions
                  class="search-results photo-results list-view"
                  disable-initial-sort
                  item-key="ID"
                  :no-data-text="notFoundMessage"
    >
      <template #items="props">
        <td style="user-select: none;" :data-uid="props.item.UID" class="result" :class="props.item.classes()">
          <v-img :key="props.item.Hash"
                 :src="props.item.thumbnailUrl('tile_50')"
                 :alt="props.item.Title"
                 :transition="false"
                 aspect-ratio="1"
                 style="user-select: none"
                 class="accent lighten-2 clickable"
                 @touchstart="onMouseDown($event, props.index)"
                 @touchend.stop.prevent="onClick($event, props.index)"
                 @mousedown="onMouseDown($event, props.index)"
                 @contextmenu.stop="onContextMenu($event, props.index)"
                 @click.stop.prevent="onClick($event, props.index)"
          >
            <v-btn v-if="props.item.Type === 'video' || props.item.Type === 'live'"
                   :ripple="false"
                   flat icon large absolute class="input-open"
                   @click.stop.prevent="openPhoto(props.index, true)">
              <v-icon color="white" class="default-hidden action-live" :title="$gettext('Live')">$vuetify.icons.live_photo</v-icon>
              <v-icon color="white" class="default-hidden action-play" :title="$gettext('Video')">play_arrow</v-icon>
            </v-btn>
          </v-img>
        </td>

        <td class="p-photo-desc clickable" :data-uid="props.item.UID" style="user-select: none;"
            @click.exact="editPhoto(props.index)">
          {{ props.item.Title }}
        </td>
        
        <td class="text-xs-left">
          <div>
            <span>{{ $gettext('Original Path') }}:&nbsp; {{ props.item.FileName }}</span>
          </div>
          <div v-for="(dup, index) in props.item.Duplicates"
          :key="index">
            <span>{{ $gettext('Duplicate Path') }}:&nbsp; {{ dup.Name }}</span>
            <v-btn small icon flat color="remove" class="ma-0 action-delete" :title="$gettext('Delete')" @click.stop.exact="props.item.deleteDuplicate(dup.Name)">
              <v-icon>delete</v-icon>
            </v-btn>
          </div>
          <div>
          <br>
          </div>
          Items: {{ props.item }}
        </td>
      </template>
    </v-data-table>
  </div>
</template>
<script>
import download from "common/download";
import Notify from "../../common/notify";

export default {
  name: 'PPhotoDuplicates',
  props: {
    photos: Array,
    openPhoto: Function,
    editPhoto: Function,
    album: Object,
    filter: Object,
    context: String,
    selectMode: Boolean,
  },
  data() {
    let m = this.$gettext("Couldn't find anything.");

    m += " " + this.$gettext("Try again using other filters or keywords.");

    if (this.$config.feature("review")) {
      m += " " + this.$gettext("Non-photographic and low-quality images require a review before they appear in search results.");
    }

    let showName = this.filter.order === 'name';

    const align = !this.$rtl ? 'left' : 'right';
    return {
      config: this.$config.values,
      notFoundMessage: m,
      'selected': [],
      'listColumns': [
        {text: '', value: '', align: 'center', class: 'p-col-select', sortable: false},
        {text: this.$gettext('Title'), align, value: 'Title', sortable: true},
        {text: this.$gettext('Paths'), value: 'Paths', align: 'left', sortable: false},
      ],
      showName: showName,
      showLocation: this.$config.values.settings.features.places,
      hidePrivate: this.$config.values.settings.features.private,
      mouseDown: {
        index: -1,
        scrollY: window.scrollY,
        timeStamp: -1,
      },
    };
  },
  methods: {
    downloadFile(index) {
      Notify.success(this.$gettext("Downloading…"));

      const photo = this.photos[index];
      download(`/api/v1/dl/${photo.Hash}?t=${this.$config.downloadToken()}`, photo.FileName);
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

        if (photo.Type === 'video' && photo.isPlayable()) {
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
