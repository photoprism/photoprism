<template>
  <div id="photoprism" :class="[isRtl ? 'is-rtl' : '', 'theme-' + themeName]">
    <p-loading-bar height="4"></p-loading-bar>

    <p-notify></p-notify>

    <v-app :class="appClass">
      <p-navigation></p-navigation>

      <v-content>
        <router-view></router-view>
      </v-content>
    </v-app>

    <p-video-viewer></p-video-viewer>
    <p-photo-viewer></p-photo-viewer>
  </div>
</template>

<script>
import "css/app.css";
import Event from "pubsub-js";

export default {
  name: "PhotoPrism",
  data() {
    return {
      isRtl: this.$config.rtl(),
      themeName: this.$config.themeName,
      subscriptions: [],
      touchStart: 0,
    };
  },
  computed: {
    appClass: function () {
      return [this.$route.meta.background, this.$vuetify.breakpoint.smAndDown ? "small-screen" : "large-screen", this.$route.meta.hideNav ? "hide-nav" : "show-nav"];
    },
  },
  created() {
    window.addEventListener("touchstart", (e) => this.onTouchStart(e), { passive: true });
    window.addEventListener("touchmove", (e) => this.onTouchMove(e), { passive: true });
    this.subscriptions["view.refresh"] = Event.subscribe("view.refresh", (ev, data) => this.onRefresh(data));
    this.$config.setVuetify(this.$vuetify);
  },
  destroyed() {
    for (let i = 0; i < this.subscriptions.length; i++) {
      Event.unsubscribe(this.subscriptions[i]);
    }
    window.removeEventListener("touchstart", (e) => this.onTouchStart(e), false);
    window.removeEventListener("touchmove", (e) => this.onTouchMove(e), false);
  },
  methods: {
    onRefresh(config) {
      this.isRtl = config.rtl();
      this.themeName = config.themeName;
    },
    onTouchStart(e) {
      this.touchStart = e.touches[0].pageY;
    },
    onTouchMove(e) {
      if (!this.touchStart) return;
      if (document.querySelector(".v-dialog--active") !== null) return;

      const y = e.touches[0].pageY;
      const h = window.document.documentElement.scrollHeight - window.document.documentElement.clientHeight;

      if (window.scrollY >= h - 200 && y < this.touchStart) {
        Event.publish("touchmove.bottom");
        this.touchStart = 0;
      } else if (window.scrollY === 0 && y > this.touchStart + 200) {
        Event.publish("touchmove.top");
        this.touchStart = 0;
      }
    },
  },
};
</script>
