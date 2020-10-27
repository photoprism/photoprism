<template>
  <v-container grid-list-xs fluid class="pa-2 p-photos p-photo-mosaic">
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
            <translate>If a file you expect is missing, please re-index your library and wait until indexing has been completed.</translate>
            <template v-if="$config.feature('review')" class="mt-2 mb-0 pa-0">
              <translate>Non-photographic and low-quality images require a review before they appear in search results.</translate>
            </template>
          </p>
        </div>
      </v-card-title>
    </v-card>
    <v-layout row wrap class="p-results">
      <p-photo-tile
        v-for="(photo, index) in photos"
        :key="index"
        :index="index"
        :photo="photo"
        :isSelected="(selection.length > 0) && $clipboard.has(photo)"
        :openPhoto="openPhoto"
        :onSelect="onSelect"
        :onClick="onClick"
        :onMouseDown="onMouseDown"
        :onContextMenu="onContextMenu"
      >
      </p-photo-tile>
    </v-layout>
  </v-container>
</template>
<script>
export default {
  name: 'p-photo-mosaic',
  props: {
    photos: Array,
    selection: Array,
    openPhoto: Function,
    filter: Object,
  },
  data() {
    return {
      mouseDown: {
        index: -1,
        timeStamp: -1,
      },
    };
  },
  methods: {
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
      } else {
        this.openPhoto(index, false);
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
    }
  },
};
</script>
