<template>
  <div class="video-wrapper" :style="style" @click.stop.prevent>
    <video :key="source" ref="video" class="video-player"
           preload="auto" controls autoplay playsinline @click.stop
           @keydown.esc.stop.prevent="$emit('close')">
      <source :src="video.url()">
    </video>
  </div>
</template>

<script>
import Video from "model/video";
import Plyr from 'plyr';

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
      required: false,
      default: ""
    },
    index: {
      type: Number,
      required: false,
      default: 0
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
    },
    videos: {
      type: Array,
      required: false,
      default: () => [],
    },
    album: {
      type: Object,
      required: false,
      default: () => {},
    },
  },
  data() {
    const c = this.$config;

    return {
      refresh: false,
      style: `width: 90vw; height: 90vh`,
      source: "",
      video: new Video(false),
      player: false,
      current: this.index,
      options: {
        iconUrl: `${c.staticUri}/video/plyr.svg`,
        controls: ['play-large', 'play', 'progress', 'current-time', 'duration', 'mute', 'volume', 'download', 'settings', 'airplay', 'fullscreen'],
        settings: ['loop'],
        captions: { active: false, language: 'auto', update: false },
        hideControls: true,
        enabled: true,
        autoplay: true,
        clickToPlay: true,
        disableContextMenu: false,
        resetOnEnd: true,
        toggleInvert: true,
        blankVideo: `${c.staticUri}/video/404.mp4`,
        loop: {
          active: false,
        },
      },
    };
  },
  watch: {
    source: function (src) {
      if (src) {
        this.setSrc(src);
      }
    },
    videos: function () {
      this.onVideos();
    },
  },
  mounted() {
    document.body.classList.add("player");
    this.render();
    window.addEventListener('keyup', this.onKeyUp);
  },
  beforeDestroy() {
    document.body.classList.remove("player");
    this.stop();
  },
  beforeUnmount() {
    try {
      if (this.player) {
        this.player.destroy();
      }
    } catch (e) {
      console.log(e);
    }
    window.removeEventListener('keyup', this.onKeyUp);
  },
  methods: {
    videoEl() {
      return this.$el.getElementsByTagName('video')[0];
    },
    updateStyle() {
      if (!this.video || !this.video.Width) {
        return;
      }

      const size = this.video.playerSize();

      this.style = `width: ${size.width.toString()}px; height: ${size.height.toString()}px`;
      const plyrEl = this.$el.getElementsByClassName('plyr')[0];
      if (plyrEl) {
        plyrEl.style.cssText = this.style;
      }
      this.$el.style.cssText = this.style;
    },
    currentVideo() {
      if(typeof this.videos[this.current] === 'undefined') {
        return Video.notFound();
      }

      return this.videos[this.current];
    },
    onVideos() {
      this.current = this.play;
      /*
      const video = this.currentVideo();

      if (!video || video.Error) {
        this.$notify.error(this.$gettext("Not Found"));
        return;
      }

      // Set video parameters.
      const size = video.playerSize();
      this.loop = video.loop();
      this.width = size.width;
      this.height = size.height;
      this.poster = video.posterUrl();
      this.source = video.url(); */
    },
    render() {
      const el = this.videoEl();
      if (!el) return;

      this.player = new Plyr(el, this.options);
      // this.player.on("ended", (ev) => { console.log("event.ended", ev); });

      this.play();
    },
    fullscreen() {
      const el = this.videoEl();
      if (!el) return;

      el.requestFullscreen();
    },
    play() {
      this.video = this.currentVideo();

      if (!this.video) {
        console.log("render: No current video");
        return;
      }

      this.updateStyle();
      this.player.source = {
        type: 'video',
        title: this.video.Title,
        sources: [
          {
            src: this.video.url(),
            // type: this.video.Mime,
          },
        ],
        poster: this.video.posterUrl("fit_720"),
      };
      this.player.loop = this.videos.length === 0 && this.video.loop();
      this.player.play();
    },
    onPrev(ev) {
      if(this.videos.length < 2) {
        this.current = 0;
        return;
      } else if(this.current <= 0) {
        return;
      }

      this.player.stop();
      this.current--;
      this.play();
    },
    onNext(ev) {
      if(this.videos.length < 2) {
        this.current = 0;
        return;
      } else if(this.current >= this.videos.length - 1) {
        return;
      }

      this.player.stop();
      this.current++;
      this.play();
    },
    onKeyUp(ev) {
      switch(ev.key) {
        case "Escape": this.$emit('close'); break;
        case "ArrowLeft": this.onPrev(ev); break;
        case "ArrowRight": this.onNext(ev); break;
      }
    },
    setSrc(src) {
      // console.log("setSrc", src);
      if (!src) {
        return;
      }

      const el = this.videoEl();
      if (!el) return;

      el.src = src;
      el.poster = this.poster;
      el.play();

      this.updateStyle();

      // console.log("el", el);

      /* this.player = new Plyr(el);
      console.log("this.player", this.player);
      this.player.play();
      this.player.source = src;
      this.player.poster = this.poster;
      */
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
