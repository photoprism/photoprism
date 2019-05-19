import Vue from "vue";
import Vuetify from "vuetify";
import Router from "vue-router";
import PhotoPrism from "app/photoprism.vue";
import Routes from "routes";
import Api from "common/api";
import Config from "common/config";
import Components from "component/register";
import Maps from "maps/register";
import Alert from "common/alert";
import Gallery from "common/gallery";
import Session from "common/session";
import Event from "pubsub-js";
import Moment from "vue-moment";
import InfiniteScroll from "vue-infinite-scroll";
import VueTruncate from "vue-truncate-filter";
import VueFullscreen from "vue-fullscreen";

// Initialize client-side session
const session = new Session(window.localStorage);
const config = new Config(window.localStorage, window.appConfig);

// Set global helpers
Vue.prototype.$event = Event;
Vue.prototype.$alert = Alert;
Vue.prototype.$gallery = new Gallery;
Vue.prototype.$session = session;
Vue.prototype.$api = Api;
Vue.prototype.$config = config;

// Register Vuetify
Vue.use(Vuetify, {
    theme: {
        primary: "#FFD600",
        secondary: "#b0bec5",
        accent: "#00B8D4",
        error: "#E57373",
        info: "#00B8D4",
        success: "#00BFA5",
        warning: "#FFD600",
        delete: "#E57373",
        love: "#EF5350",
    },
});

// Register other VueJS plugins
Vue.use(Moment);
Vue.use(InfiniteScroll);
Vue.use(VueTruncate);
Vue.use(VueFullscreen);
Vue.use(Components);
Vue.use(Maps);
Vue.use(Router);

// Configure client-side routing
const router = new Router({
    routes: Routes,
    mode: "history",
    saveScrollPosition: true,
});

// Run app
/* eslint-disable no-unused-vars */
const app = new Vue({
    el: "#photoprism",
    router,
    render: h => h(PhotoPrism),
});
