<template>
  <div class="p-photo-preview ra-6 pa-0 ma-0 elevation-0 v-card v-sheet v-sheet--tile no-transition" :title="title">
    <div class="v-responsive v-image card elevation-0 clickable" @click.prevent.stop="openPhoto">
      <div class="v-responsive__sizer" style="padding-bottom: 100%"></div>
      <div class="v-image__image v-image__image--cover w-100" :style="cover"></div>
    </div>
  </div>
</template>
<script>
import Thumb from "model/thumb";

export default {
  name: "PPhotoPreview",
  props: {
    uid: {
      type: String,
      default: "",
    },
  },
  data() {
    const view = this.$view.data();
    return {
      view,
      url: view.model.thumbnailUrl("tile_500"),
      title: view.model.Title ? view.model.Title : "",
    };
  },
  computed: {
    cover() {
      return `background-image: url('${this.url}'); background-position: center center;background-size: cover;`;
    },
  },
  watch: {
    uid() {
      this.url = this.view.model.thumbnailUrl("tile_500");
      this.title = this.view.model.Title ? this.view.model.view.Title : "";
    },
  },
  methods: {
    openPhoto() {
      if (!this.$lightbox || !this.view.model) {
        return;
      }

      this.$root.$refs.lightbox.showThumbs(Thumb.fromFiles([this.view.model]), 0);
    },
  },
};
</script>
