<template>
  <div class="p-photo-preview pa-0 ma-0 elevation-0 v-card v-sheet v-sheet--tile no-transition" :title="title">
    <div class="v-responsive v-image card darken-1 elevation-0 clickable">
      <div class="v-responsive__sizer" style="padding-bottom: 100%;"></div>
      <div class="v-image__image v-image__image--cover" :style="cover"></div>
      <div class="v-responsive__content"></div>
    </div>
  </div>
</template>
<script>
import Photo from "model/photo";
import Thumb from "model/thumb";

export default {
  name: 'PPhotoPreview',
  props: {
    model: {
      type: Object,
      default: () => new Photo(false),
    },
  },
  data() {
    return {
      url: this.model.thumbnailUrl('tile_500'),
      title: this.model.Title ? this.model.Title : '',
    };
  },
  computed: {
    cover() {
      return `background-image: url('${this.url}'); background-position: center center;`;
    },
  },
  watch: {
    model() {
      this.url = this.model.thumbnailUrl('tile_500');
      this.title = this.model.Title ? this.model.Title : '';
    },
  },
  methods: {
    openPhoto() {
      this.$viewer.show(Thumb.fromFiles([this.model]), 0);
    },
  },
};
</script>