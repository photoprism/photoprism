<template>
  <div id="photoprism" :class="['theme-' + themeName]">
    <p-loading-bar height="4"></p-loading-bar>

    <p-notify></p-notify>

    <v-app :class="appClass">
      <p-navigation></p-navigation>

      <v-main>
        <router-view></router-view>
      </v-main>
    </v-app>

    <p-dialogs></p-dialogs>
  </div>
</template>

<script>
import PLoadingBar from "component/loading-bar.vue";
import PNotify from "component/notify.vue";
import PNavigation from "component/navigation.vue";
import PDialogs from "component/dialogs.vue";

export default {
  name: "App",
  components: {
    PLoadingBar,
    PNotify,
    PNavigation,
    PDialogs,
  },
  data() {
    return {
      themeName: this.$config.themeName,
      subscriptions: [],
      touchStart: 0,
    };
  },
  computed: {
    appClass: function () {
      return [
        this.$route.meta.background,
        this.$vuetify.display.smAndDown ? "small-screen" : "large-screen",
        this.$route.meta.hideNav ? "hide-nav" : "show-nav",
      ];
    },
  },
  created() {
    this.subscriptions["view.refresh"] = this.$event.subscribe("view.refresh", (ev, data) => this.onRefresh(data));
    // this.subscriptions["lightbox"] = this.$event.subscribe("lightbox", (ev, data) => { console.log(ev); });

    this.$config.setVuetify(this.$vuetify);
  },
  mounted() {
    this.$view.enter(this);
  },
  beforeUnmount() {
    for (let i = 0; i < this.subscriptions.length; i++) {
      this.$event.unsubscribe(this.subscriptions[i]);
    }
  },
  unmounted() {
    this.$view.leave(this);
  },
  methods: {
    onRefresh(config) {
      this.themeName = config.themeName;
    },
  },
};
</script>
