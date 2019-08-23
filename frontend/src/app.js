import Vue from "vue";
import Vuetify from "vuetify";
import Router from "vue-router";
import PhotoPrism from "photoprism.vue";
import Routes from "routes";
import Api from "common/api";
import Config from "common/config";
import Clipboard from "common/clipboard";
import Components from "component/components";
import Dialogs from "dialog/dialogs";
import Maps from "maps/components";
import Alert from "common/alert";
import Viewer from "common/viewer";
import Session from "common/session";
import Event from "pubsub-js";
import VueMoment from "vue-moment";
import VueInfiniteScroll from "vue-infinite-scroll";
import VueFullscreen from "vue-fullscreen";
import VueFilters from "vue2-filters";

// Initialize helpers
const session = new Session(window.localStorage);
const config = new Config(window.localStorage, window.appConfig);
const viewer = new Viewer();
const clipboard = new Clipboard(window.localStorage, "photo_clipboard");

// Assign helpers to VueJS prototype
Vue.prototype.$event = Event;
Vue.prototype.$alert = Alert;
Vue.prototype.$viewer = viewer;
Vue.prototype.$session = session;
Vue.prototype.$api = Api;
Vue.prototype.$config = config;
Vue.prototype.$clipboard = clipboard;

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
Vue.use(VueMoment);
Vue.use(VueInfiniteScroll);
Vue.use(VueFullscreen);
Vue.use(VueFilters);
Vue.use(Components);
Vue.use(Dialogs);
Vue.use(Maps);
Vue.use(Router);

// Configure client-side routing
export const router = new Router({
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
