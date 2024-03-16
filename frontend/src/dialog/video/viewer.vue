<template>
  <div
    v-if="show"
    class="video-viewer"
    role="dialog"
    @click.stop.prevent="onClose"
    @keydown.esc.stop.prevent="onClose"
  >
    <p-video-player
      v-show="show"
      ref="player"
      :source="source"
      :poster="poster"
      :height="height"
      :width="width"
      :autoplay="true"
      :loop="loop"
      @close="onClose"
    ></p-video-player>
  </div>
</template>
<script>
import Event from "pubsub-js";

export default {
  name: "PVideoViewer",
  data() {
    return {
      show: false,
      source: "",
      poster: `${this.$config.contentUri}/svg/video`,
      defaultWidth: 640,
      defaultHeight: 480,
      width: 640,
      height: 480,
      video: null,
      album: null,
      loop: false,
      subscriptions: [],
    };
  },
  created() {
    this.subscriptions["player.open"] = Event.subscribe("player.open", this.onOpen);
    this.subscriptions["player.pause"] = Event.subscribe("player.pause", this.onPause);
    this.subscriptions["player.close"] = Event.subscribe("player.close", this.onClose);
  },
  beforeDestroy() {
    for (let i = 0; i < this.subscriptions.length; i++) {
      Event.unsubscribe(this.subscriptions[i]);
    }

    this.onClose();
  },
  methods: {
    onOpen(ev, params) {
      const fullscreen = !!params.fullscreen;

      this.video = params.video;
      this.album = params.album;

      this.play(fullscreen);
    },
    onPause() {
      if (this.$refs.player) {
        this.$refs.player.pause();
      }
    },
    onClose() {
      if (this.$refs.player) {
        this.$refs.player.stop();
      }

      this.show = false;
    },
    play(fullscreen) {
      if (!this.video) {
        this.$notify.error(this.$gettext("No video selected"));
        return;
      }

      const params = this.video.videoParams();

      if (params.error) {
        this.$notify.error(params.error);
        return;
      }

      // Set video parameters.
      this.loop = params.loop;
      this.width = params.width;
      this.height = params.height;
      this.poster = params.poster;
      this.source = params.uri;

      // Play video.
      this.show = true;

      if (fullscreen) {
        this.$nextTick(() => this.$refs.player.fullscreen());
      }
    },
  },
};
</script>
