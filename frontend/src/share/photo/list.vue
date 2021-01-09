<template>
  <div>
    <div v-if="photos.length === 0" class="pa-2">
      <v-card class="p-photos-empty secondary-light lighten-1 ma-1" flat>
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
                  class="elevation-0 p-photos p-photo-list p-results"
                  disable-initial-sort
                  item-key="ID"
                  :no-data-text="notFoundMessage"
    >
      <template slot="items" slot-scope="props">
        <td style="user-select: none;" :data-uid="props.item.UID">
          <v-img class="accent lighten-2 clickable" aspect-ratio="1"
                 :src="props.item.thumbnailUrl('tile_50')"
                 @mousedown="onMouseDown($event, props.index)"
                 @contextmenu="onContextMenu($event, props.index)"
                 @click.stop.prevent="onClick($event, props.index)"
          >
            <v-btn v-if="props.item.Selected" :ripple="false"
                   flat icon large absolute class="p-photo-select">
              <v-icon color="white" class="t-select t-on">check_circle</v-icon>
            </v-btn>
            <v-btn v-else-if="!selectMode && (props.item.Type === 'video' || props.item.Type === 'live')"
                   :ripple="false"
                   flat icon large absolute class="p-photo-play opacity-75"
                   @click.stop.prevent="openPhoto(props.index, true)">
              <v-icon color="white" class="action-play">play_arrow</v-icon>
            </v-btn>
          </v-img>
        </td>

        <td class="p-photo-desc clickable" :data-uid="props.item.UID"
            style="user-select: none;" @click.stop.prevent="openPhoto(props.index, false)">
          {{ props.item.Title }}
        </td>
        <td class="p-photo-desc hidden-xs-only" :title="props.item.getDateString()"
            style="user-select: none;" @click.stop.prevent="openPhoto(props.index, false)">
          {{ props.item.shortDateString() }}
        </td>
        <td class="p-photo-desc hidden-sm-and-down" style="user-select: none;">
          {{ props.item.CameraMake }} {{ props.item.CameraModel }}
        </td>
        <td class="p-photo-desc hidden-xs-only">
          <button v-if="filter.order === 'name'"
                  title="Name" @click.exact="downloadFile(props.index)">
            {{ props.item.FileName }}
          </button>
          <button v-else-if="props.item.Country !== 'zz' && showLocation"
                  style="user-select: none;"
                  @click.stop.prevent="openPhoto(props.index, false)">
            {{ props.item.locationInfo() }}
          </button>
          <span v-else>
                    {{ props.item.locationInfo() }}
                </span>
        </td>
      </template>
    </v-data-table>
  </div>
</template>
<script>
export default {
  name: 'PPhotoList',
  props: {
    photos: Array,
    openPhoto: Function,
    editPhoto: Function,
    openLocation: Function,
    album: Object,
    filter: Object,
    selectMode: Boolean,
  },
  data() {
    let m = this.$gettext("Couldn't find anything.");

    m += " " + this.$gettext("Try again using other filters or keywords.");

    let showName = this.filter.order === 'name';

    return {
      notFoundMessage: m,
      'selected': [],
      'listColumns': [
        {text: '', value: '', align: 'center', class: 'p-col-select', sortable: false},
        {text: this.$gettext('Title'), value: 'Title', sortable: false},
        {text: this.$gettext('Taken'), class: 'hidden-xs-only', value: 'TakenAt', sortable: false},
        {text: this.$gettext('Camera'), class: 'hidden-sm-and-down', value: 'CameraModel', sortable: false},
        {
          text: showName ? this.$gettext('Name') : this.$gettext('Location'),
          class: 'hidden-xs-only',
          value: showName ? 'FileName' : 'PlaceLabel',
          sortable: false
        },
      ],
      showName: showName,
      showLocation: this.$config.settings().features.places,
      hidePrivate: this.$config.settings().features.private,
      mouseDown: {
        index: -1,
        timeStamp: -1,
      },
    };
  },
  watch: {
    photos: function (photos) {
      this.selected.splice(0);

      for (let i = 0; i < photos.length; i++) {
        if (this.$clipboard.has(photos[i])) {
          this.selected.push(photos[i]);
        }
      }
    },
    selection: function () {
      this.refreshSelection();
    },
  },
  mounted: function () {
    this.$nextTick(function () {
      this.refreshSelection();
    });
  },
  methods: {
    downloadFile(index) {
      const photo = this.photos[index];
      const link = document.createElement('a');
      link.href = `/api/v1/dl/${photo.Hash}?t=${this.$config.downloadToken()}`;
      link.download = photo.FileName;
      link.click();
    },
    onSelect(ev, index) {
      if (ev.shiftKey) {
        this.selectRange(index);
      } else {
        this.$clipboard.toggle(this.photos[index]);
      }
    },
    onMouseDown(ev, index) {
      this.mouseDown.index = index;
      this.mouseDown.timeStamp = ev.timeStamp;
    },
    onClick(ev, index) {
      let longClick = (this.mouseDown.index === index && ev.timeStamp - this.mouseDown.timeStamp > 400);

      if (longClick || this.selectMode) {
        if (longClick || ev.shiftKey) {
          this.selectRange(index);
        } else {
          this.$clipboard.toggle(this.photos[index]);
        }
      } else if (this.photos[index]) {
        let photo = this.photos[index];

        if (photo.Type === 'video' || photo.Type === 'live') {
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
    selectRange(index) {
      this.$clipboard.addRange(index, this.photos);
    },
    refreshSelection() {
      this.selected.splice(0);

      for (let i = 0; i < this.photos.length; i++) {
        if (this.$clipboard.has(this.photos[i])) {
          this.selected.push(this.photos[i]);
        }
      }
    },
  }
};
</script>
