<template>
  <div>
    <v-card v-if="photos.length === 0" class="p-photos-empty secondary-light lighten-1 ma-1" flat>
      <v-card-title primary-title>
        <div>
          <h3 class="title ma-0 pa-0" v-if="filter.order === 'edited'">
            <translate>Couldn't find recently edited</translate>
          </h3>
          <h3 class="title ma-0 pa-0" v-else>
            <translate>Couldn't find anything</translate>
          </h3>
          <p class="mt-4 mb-0 pa-0">
            <translate>Try again using other filters or keywords.</translate>
          </p>
        </div>
      </v-card-title>
    </v-card>
    <v-data-table v-else
                  :headers="listColumns"
                  :items="photos"
                  hide-actions
                  class="elevation-0 p-photos p-photo-list p-results"
                  disable-initial-sort
                  item-key="ID"
                  v-model="selected"
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
            <v-layout
                    slot="placeholder"
                    fill-height
                    align-center
                    justify-center
                    ma-0
            >
              <v-progress-circular indeterminate
                                   color="accent lighten-5"></v-progress-circular>
            </v-layout>

            <v-btn v-if="selection.length && $clipboard.has(props.item)" :ripple="false"
                   flat icon large absolute class="p-photo-select">
              <v-icon color="white" class="t-select t-on">check_circle</v-icon>
            </v-btn>
            <v-btn v-else-if="!selection.length && props.item.Type === 'video' && props.item.isPlayable()"
                   :ripple="false"
                   flat icon large absolute class="p-photo-play opacity-75"
                   @click.stop.prevent="openPhoto(props.index, true)">
              <v-icon color="white" class="action-play">play_arrow</v-icon>
            </v-btn>
          </v-img>
        </td>

        <td class="p-photo-desc clickable" :data-uid="props.item.UID"
            @click.stop.prevent="openPhoto(props.index, false)" style="user-select: none;">
          {{ props.item.Title }}
        </td>
        <td class="p-photo-desc hidden-xs-only" :title="props.item.getDateString()"
            @click.stop.prevent="openPhoto(props.index, false)" style="user-select: none;">
          {{ props.item.shortDateString() }}
        </td>
        <td class="p-photo-desc hidden-sm-and-down" style="user-select: none;">
          {{ props.item.CameraMake }} {{ props.item.CameraModel }}
        </td>
        <td class="p-photo-desc hidden-xs-only">
          <button @click.exact="downloadFile(props.index)"
                  title="Name" v-if="filter.order === 'name'">
            {{ props.item.FileName }}
          </button>
          <button v-else-if="props.item.Country !== 'zz' && showLocation"
                  @click.stop.prevent="openPhoto(props.index, false)"
                  style="user-select: none;">
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
        name: 'p-photo-list',
        props: {
            photos: Array,
            selection: Array,
            openPhoto: Function,
            editPhoto: Function,
            openLocation: Function,
            album: Object,
            filter: Object,
        },
        data() {
            let m = this.$gettext("Couldn't find anything.");

            m += " " + this.$gettext("Try again using other filters or keywords.");

            let showName = this.filter.order === 'name'

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

                for (let i = 0; i <= photos.length; i++) {
                    if (this.$clipboard.has(photos[i])) {
                        this.selected.push(photos[i]);
                    }
                }
            },
            selection: function () {
                this.refreshSelection();
            },
        },
        methods: {
            downloadFile(index) {
                const photo = this.photos[index];
                const link = document.createElement('a')
                link.href = `/api/v1/dl/${photo.Hash}?t=${this.$config.downloadToken()}`;
                link.download = photo.FileName;
                link.click()
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

                if (longClick || this.selection.length > 0) {
                    if (longClick || ev.shiftKey) {
                        this.selectRange(index);
                    } else {
                        this.$clipboard.toggle(this.photos[index]);
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
            selectRange(index) {
                this.$clipboard.addRange(index, this.photos);
            },
            refreshSelection() {
                this.selected.splice(0);

                for (let i = 0; i <= this.photos.length; i++) {
                    if (this.$clipboard.has(this.photos[i])) {
                        this.selected.push(this.photos[i]);
                    }
                }
            },
        },
        mounted: function () {
            this.$nextTick(function () {
                this.refreshSelection();
            })
        }
    };
</script>
