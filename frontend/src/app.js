import Vue from "vue";
import Vuetify from "vuetify";
import Router from "vue-router";
import PhotoPrism from "photoprism.vue";
import Routes from "routes";
import Api from "common/api";
import Socket from "common/websocket";
import Config from "common/config";
import Clipboard from "common/clipboard";
import Components from "component/components";
import Dialogs from "dialog/dialogs";
import Maps from "maps/components";
import Alert from "common/alert";
import Viewer from "common/viewer";
import Session from "session";
import Event from "pubsub-js";
import VueLuxon from "vue-luxon";
import VueInfiniteScroll from "vue-infinite-scroll";
import VueFullscreen from "vue-fullscreen";
import VueFilters from "vue2-filters";
import GetTextPlugin from "vue-gettext";
import translations from "./i18n/translations.json";
import { Settings } from "luxon";

// Initialize helpers
const config = new Config(window.localStorage, window.appConfig);
const viewer = new Viewer();
const clipboard = new Clipboard(window.localStorage, "photo_clipboard");
const isPublic = config.getValue("public");

// Assign helpers to VueJS prototype
Vue.prototype.$event = Event;
Vue.prototype.$alert = Alert;
Vue.prototype.$viewer = viewer;
Vue.prototype.$session = Session;
Vue.prototype.$api = Api;
Vue.prototype.$socket = Socket;
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

Vue.config.language = "en";
Settings.defaultLocale = Vue.config.language;

// Register other VueJS plugins
Vue.use(GetTextPlugin, {translations: translations, silent: false, defaultLanguage: Vue.config.language});
Vue.use(VueLuxon);
Vue.use(VueInfiniteScroll);
Vue.use(VueFullscreen);
Vue.use(VueFilters);
Vue.use(Components);
Vue.use(Dialogs);
Vue.use(Maps);
Vue.use(Router);

// Configure client-side routing
const router = new Router({
    routes: Routes,
    mode: "history",
    saveScrollPosition: true,
});

router.beforeEach((to, from, next) => {
    if(to.matched.some(record => record.meta.admin)) {
        if (isPublic || Session.isAdmin()) {
            next();
        } else {
            next({
                name: "login",
                params: { nextUrl: to.fullPath },
            });
        }
    } else if(to.matched.some(record => record.meta.auth)) {
        if (isPublic || Session.isUser()) {
            next();
        } else {
            next({
                name: "login",
                params: { nextUrl: to.fullPath },
            });
        }
    } else {
        next();
    }
});

// Run app
/* eslint-disable no-unused-vars */
const app = new Vue({
    el: "#photoprism",
    router,
    render: h => h(PhotoPrism),
});
