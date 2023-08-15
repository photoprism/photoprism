<template>
  <div class="video-wrapper" :style="style">
    <video :key="source" ref="video" class="video-player" :height="height" :width="width" :autoplay="autoplay"
           :style="style" :poster="poster" :loop="loop" preload="auto" controls playsinline
           @click.stop @keydown.esc.stop.prevent="$emit('close')">
      <source :src="source">
    </video>
  </div>
</template>

<script>
export default {
  name: "PVideoPlayer",
  props: {
    show: {
      type: Boolean,
      required: false,
      default: false
    },
    poster: {
      type: String,
      required: true,
      default: ""
    },
    source: {
      type: String,
      required: true,
      default: ""
    },
    width: {
      type: Number,
      required: false,
      default: 640
    },
    height: {
      type: Number,
      required: false,
      default: 480
    },
    preload: {
      type: String,
      required: false,
      default: "none"
    },
    autoplay: {
      type: Boolean,
      required: false,
      default: false
    },
    loop: {
      type: Boolean,
      required: false,
      default: false
    },
    success: {
      type: Function,
      default: () => {},
    },
    error: {
      type: Function,
      default: () => {},
    }
  },
  data: () => ({
    refresh: false,
    style: `width: 90vw; height: 90vh`,
  }),
  watch: {
    source: function (src) {
      if (src) {
        this.setSrc(src);
      }
    },
  },
  mounted() {
    document.body.classList.add("player");
    this.render();
  },
  beforeDestroy() {
    document.body.classList.remove("player");
    this.stop();
  },
  methods: {
    videoEl() {
      return this.$el.getElementsByTagName('video')[0];
    },
    updateStyle() {
      this.style = `width: ${this.width.toString()}px; height: ${this.height.toString()}px`;
      this.$el.style.cssText = this.style;
    },
    render() {
      this.updateStyle();
    },
    fullscreen() {
      const el = this.videoEl();
      if (!el) return;

      el.requestFullscreen();
    },
    setSrc(src) {
      if (!src) {
        return;
      }

      this.updateStyle();

      const el = this.videoEl();
      if (!el) return;

      el.src = src;
      el.poster = this.poster;
      el.play();
    },
    pause() {
      const el = this.videoEl();
      if (!el) return;

      el.pause();
    },
    stop() {
      const el = this.videoEl();
      if (!el) return;

      el.pause();
      el.src = "";
      el.poster = "";
      el.load();
    },
  },
};
</script>
