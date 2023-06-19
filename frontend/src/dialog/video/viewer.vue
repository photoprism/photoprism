<template>
  <div v-if="show" class="video-viewer" role="dialog" @click.stop.prevent="onClose" @keydown.esc.stop.prevent="onClose">
      <p-video-player v-show="show" ref="player" :videos="videos" :index="index" :album="album" @close="onClose"></p-video-player>
  </div>
</template>
<script>
import Event from "pubsub-js";

export default {
  name: 'PVideoViewer',
  data() {
    return {
      show: false,
      source: "",
      poster: `${this.$config.contentUri}/svg/video`,
      defaultWidth: 640,
      defaultHeight: 480,
      width: 640,
      height: 480,
      index: 0,
      videos: [],
      album: null,
      loop: false,
      subscriptions: [],
    };
  },
  created() {
    this.subscriptions['player.open'] = Event.subscribe('player.open', this.onOpen);
    this.subscriptions['player.pause'] = Event.subscribe('player.pause', this.onPause);
    this.subscriptions['player.close'] = Event.subscribe('player.close', this.onClose);
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
      const hasQueue = params.videos && params.videos.length > 0;

      this.videos = hasQueue ? params.videos : [];
      this.album = params.album ? params.album : null;

      if(params.index && params.index < this.videos.length) {
        this.index = params.index;
      } else {
        this.index = 0;
      }

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
      if (!this.videos) {
        this.$notify.error(this.$gettext("No videos found to play"));
        return;
      }

      // Play video.
      this.show = true;

      if (fullscreen) {
        this.$nextTick(() => this.$refs.player.fullscreen());
      }
    },
  },
};
</script>
