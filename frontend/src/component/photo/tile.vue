<template>
  <v-flex
      :data-uid="photo.UID"
      v-bind:class="{ selected: isSelected }"
      class="p-photo"
      xs4 sm3 md2 lg1 d-flex
  >
    <v-hover>
      <v-card tile slot-scope="{ hover }"
              @contextmenu="onContextMenu($event, index)"
              :class="isSelected ? 'elevation-10 ma-0' : 'elevation-0 ma-1'"
              :title="photo.Title">
        <v-img :src="photo.thumbnailUrl('tile_224')"
               aspect-ratio="1"
               class="accent lighten-2 clickable"
               @mousedown="onMouseDown($event, index)"
               @click.stop.prevent="onClick($event, index)"
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

          <v-layout
              fill-height
              align-center
              justify-center
              ma-0
              class="p-photo-live"
              style="overflow: hidden;"
              v-if="photo.Type === 'live'"
              v-show="hover"
          >
            <video width="224" height="224" autoplay loop muted playsinline :key="photo.videoUrl()">
              <source :src="photo.videoUrl()" type="video/mp4">
            </video>
          </v-layout>

          <v-btn v-if="hidePrivate && photo.Private" :ripple="false"
                 icon flat small absolute
                 class="p-photo-private opacity-75">
            <v-icon color="white">lock</v-icon>
          </v-btn>

          <v-btn v-if="hover || isSelected" :ripple="false"
                 icon flat small absolute
                 :class="isSelected ? 'p-photo-select' : 'p-photo-select opacity-50'"
                 @click.stop.prevent="onSelect($event, index)">
            <v-icon v-if="isSelected" color="red"
                    class="t-select t-on">check_circle
            </v-icon>
            <v-icon v-else color="accent lighten-3" class="t-select t-off">radio_button_off</v-icon>
          </v-btn>

          <v-btn icon flat small absolute :ripple="false"
                 :class="photo.Favorite ? 'p-photo-like opacity-75' : 'p-photo-like opacity-50'"
                 @click.stop.prevent="photo.toggleLike()">
            <v-icon v-if="photo.Favorite" color="white" class="t-like t-on" :data-uid="photo.UID">favorite</v-icon>
            <v-icon v-else color="accent lighten-3" class="t-like t-off" :data-uid="photo.UID">favorite_border
            </v-icon>
          </v-btn>

          <template v-if="photo.isPlayable()">
            <v-btn v-if="photo.Type === 'live'" color="white"
                   icon flat small absolute class="p-photo-live opacity-75" :depressed="false" :ripple="false"
                   @click.stop.prevent="openPhoto(index, true)" title="Live Photo">
              <v-icon color="white" class="action-play">adjust</v-icon>
            </v-btn>
            <v-btn v-else color="white"
                   outline fab absolute class="p-photo-play opacity-75" :depressed="false" :ripple="false"
                   @click.stop.prevent="openPhoto(index, true)" title="Play">
              <v-icon color="white" class="action-play">play_arrow</v-icon>
            </v-btn>
          </template>
          <v-btn v-else-if="photo.Type === 'image' && photo.Files.length > 1" :ripple="false"
                 icon flat small absolute class="p-photo-merged opacity-75"
                 @click.stop.prevent="openPhoto(index, true)">
            <v-icon color="white" class="action-burst">burst_mode</v-icon>
          </v-btn>
          <v-btn v-else-if="photo.Type === 'image' && hover && ($clipboard.selection.length > 0)" :ripple="false"
                 icon flat small absolute class="p-photo-merged opacity-75"
                 @click.stop.prevent="openPhoto(index, false)">
            <v-icon color="white" class="action-open">zoom_in</v-icon>
          </v-btn>
          <v-btn v-else-if="photo.Type === 'raw'" :ripple="false"
                 icon flat small absolute class="p-photo-merged opacity-75"
                 @click.stop.prevent="openPhoto(index, true)" title="RAW">
            <v-icon color="white" class="action-burst">photo_camera</v-icon>
          </v-btn>
        </v-img>
      </v-card>
    </v-hover>
  </v-flex>
</template>
<script>
export default {
  name: 'p-photo-tile',
  props: {
    photo: Object,
    index: Number,
    isSelected: Boolean,
    openPhoto: Function,
    onSelect: Function,
    onClick: Function,
    onMouseDown: Function,
    onContextMenu: Function,
  },
  data() {
    return {
      hidePrivate: this.$config.settings().features.private,
    };
  },
};
</script>
